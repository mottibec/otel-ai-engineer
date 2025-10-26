package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/agent"
	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// RunService handles business logic for agent runs
type RunService struct {
	storage         storage.Storage
	agentRegistry   *agent.Registry
	anthropicClient *anthropic.Client
	logLevel        config.LogLevel
	eventBridge     EventBridge
	activeRuns      ActiveRunManager
}

// EventBridge interface for event emission
type EventBridge interface {
	GetEmitter() events.EventEmitter
}

// ActiveRunManager interface for managing active runs
type ActiveRunManager interface {
	Add(runID string, ctx context.Context, cancel context.CancelFunc)
	Get(runID string) (*ActiveRun, bool)
	Remove(runID string)
}

// ActiveRun represents an active agent run
type ActiveRun struct {
	RunID          string
	Context        context.Context
	CancelFunc     context.CancelFunc
	PendingMessage chan string
}

// Config holds configuration for the run service
type Config struct {
	Storage         storage.Storage
	AgentRegistry   *agent.Registry
	AnthropicClient *anthropic.Client
	LogLevel        config.LogLevel
	EventBridge     EventBridge
	ActiveRuns      ActiveRunManager
}

// NewRunService creates a new run service
func NewRunService(cfg Config) *RunService {
	return &RunService{
		storage:         cfg.Storage,
		agentRegistry:   cfg.AgentRegistry,
		anthropicClient: cfg.AnthropicClient,
		logLevel:        cfg.LogLevel,
		eventBridge:     cfg.EventBridge,
		activeRuns:      cfg.ActiveRuns,
	}
}

// CreateRunRequest represents a request to create a new run
type CreateRunRequest struct {
	AgentID         string
	Prompt          string
	ResumeFromRunID string
}

// CreateRunResponse represents the response from creating a run
type CreateRunResponse struct {
	Run    *storage.Run
	RunID  string
	Status string
}

// CreateRun creates and starts a new agent run
func (s *RunService) CreateRun(ctx context.Context, req CreateRunRequest) (*CreateRunResponse, error) {
	// Validate agent exists
	agentInfo, exists := s.agentRegistry.Get(req.AgentID)
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", req.AgentID)
	}

	// Generate run ID
	runID := fmt.Sprintf("run-%d", time.Now().UnixNano())

	// Create cancellable context
	runCtx, cancel := context.WithCancel(context.Background())

	// Track active run
	s.activeRuns.Add(runID, runCtx, cancel)

	// Get active run for pending messages channel
	activeRun, exists := s.activeRuns.Get(runID)
	if !exists {
		return nil, fmt.Errorf("failed to track active run")
	}

	// Build history if resuming
	var history []anthropic.MessageParam
	if req.ResumeFromRunID != "" {
		var err error
		history, err = s.buildHistoryFromEvents(req.ResumeFromRunID)
		if err != nil {
			log.Printf("Failed to build history: %v", err)
			// Continue without history rather than failing
		}
	}

	// Start agent in background
	go func() {
		defer cancel()
		defer s.activeRuns.Remove(runID)

		_, err := s.agentRegistry.RunAgent(runCtx, agent.RunnerConfig{
			AgentID:         req.AgentID,
			Prompt:          req.Prompt,
			Client:          s.anthropicClient,
			LogLevel:        s.logLevel,
			EventEmitter:    s.eventBridge.GetEmitter(),
			PendingMessages: activeRun.PendingMessage,
			History:         history,
		})
		if err != nil {
			log.Printf("Agent run failed: %v", err)
		}
	}()

	// Wait briefly for run to be created in storage
	time.Sleep(100 * time.Millisecond)

	// Fetch the created run
	runs, err := s.storage.ListRuns(storage.RunListOptions{
		Limit:  1,
		Offset: 0,
	})

	if err != nil || len(runs) == 0 {
		// Return basic response if we can't fetch yet
		return &CreateRunResponse{
			RunID:  runID,
			Status: "started",
			Run: &storage.Run{
				ID:        runID,
				AgentID:   req.AgentID,
				AgentName: agentInfo.Name,
				Prompt:    req.Prompt,
				Status:    storage.RunStatusRunning,
			},
		}, nil
	}

	return &CreateRunResponse{
		Run:    runs[0],
		RunID:  runs[0].ID,
		Status: "created",
	}, nil
}

// ResumeRunRequest represents a request to resume a run
type ResumeRunRequest struct {
	RunID   string
	Message string
}

// ResumeRun resumes an existing run with a new message
func (s *RunService) ResumeRun(ctx context.Context, req ResumeRunRequest) (*storage.Run, error) {
	// Get the original run
	run, err := s.storage.GetRun(req.RunID)
	if err != nil {
		return nil, fmt.Errorf("run not found: %w", err)
	}

	// Normalize agent ID (convert display name to registry ID)
	agentID := s.normalizeAgentID(run.AgentID)

	// Verify agent exists
	_, exists := s.agentRegistry.Get(agentID)
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	// Build history from events
	history, err := s.buildHistoryFromEvents(req.RunID)
	if err != nil {
		return nil, fmt.Errorf("failed to build history: %w", err)
	}

	// Update run status to running
	status := storage.RunStatusRunning
	update := &storage.RunUpdate{
		Status:       &status,
		ClearEndTime: true,
	}
	if err := s.storage.UpdateRun(req.RunID, update); err != nil {
		return nil, fmt.Errorf("failed to update run: %w", err)
	}

	// Create cancellable context
	runCtx, cancel := context.WithCancel(context.Background())

	// Track active run
	s.activeRuns.Add(req.RunID, runCtx, cancel)

	// Get active run for pending messages
	activeRun, exists := s.activeRuns.Get(req.RunID)
	if !exists {
		return nil, fmt.Errorf("failed to track active run")
	}

	// Start agent in background with existing run ID
	go func() {
		defer cancel()
		defer s.activeRuns.Remove(req.RunID)

		_, err := s.agentRegistry.RunAgent(runCtx, agent.RunnerConfig{
			AgentID:         agentID,
			Prompt:          req.Message,
			Client:          s.anthropicClient,
			LogLevel:        s.logLevel,
			EventEmitter:    s.eventBridge.GetEmitter(),
			PendingMessages: activeRun.PendingMessage,
			History:         history,
			RunID:           req.RunID, // Use existing run ID
		})
		if err != nil {
			log.Printf("Agent run failed: %v", err)
		}
	}()

	// Wait briefly for agent to start
	time.Sleep(100 * time.Millisecond)

	// Get updated run
	updatedRun, err := s.storage.GetRun(req.RunID)
	if err != nil {
		// Return basic response if we can't fetch yet
		return &storage.Run{
			ID:     req.RunID,
			Status: storage.RunStatusRunning,
		}, nil
	}

	return updatedRun, nil
}

// StopRun stops an active run
func (s *RunService) StopRun(ctx context.Context, runID string) error {
	activeRun, exists := s.activeRuns.Get(runID)
	if !exists {
		return fmt.Errorf("run not found or not active")
	}

	// Cancel the run
	if activeRun.CancelFunc != nil {
		activeRun.CancelFunc()
	}

	// Update status
	status := storage.RunStatusCancelled
	update := &storage.RunUpdate{
		Status: &status,
	}
	return s.storage.UpdateRun(runID, update)
}

// PauseRun pauses an active run
func (s *RunService) PauseRun(ctx context.Context, runID string) error {
	run, err := s.storage.GetRun(runID)
	if err != nil {
		return fmt.Errorf("run not found: %w", err)
	}

	if run.Status != storage.RunStatusRunning {
		return fmt.Errorf("cannot pause run in status %s", run.Status)
	}

	status := storage.RunStatusPaused
	update := &storage.RunUpdate{
		Status: &status,
	}
	return s.storage.UpdateRun(runID, update)
}

// AddInstruction adds an instruction to a running or paused run
func (s *RunService) AddInstruction(ctx context.Context, runID string, instruction string) error {
	run, err := s.storage.GetRun(runID)
	if err != nil {
		return fmt.Errorf("run not found: %w", err)
	}

	// Create user message event
	messageData := events.MessageData{
		Role: "user",
		Content: []events.ContentBlock{
			{
				Type: "text",
				Text: instruction,
			},
		},
	}

	event, err := events.NewMessageEvent(runID, run.AgentID, run.AgentName, messageData)
	if err != nil {
		return fmt.Errorf("failed to create message event: %w", err)
	}

	// Store event
	if err := s.storage.AddEvent(runID, event); err != nil {
		return fmt.Errorf("failed to add event: %w", err)
	}

	// Handle based on run status
	if run.Status == storage.RunStatusPaused {
		// Resume if paused
		status := storage.RunStatusRunning
		update := &storage.RunUpdate{
			Status: &status,
		}
		return s.storage.UpdateRun(runID, update)
	} else if run.Status == storage.RunStatusRunning {
		// Queue message if running
		if activeRun, exists := s.activeRuns.Get(runID); exists {
			select {
			case activeRun.PendingMessage <- instruction:
				// Queued successfully
			default:
				log.Printf("Warning: Pending message channel full for run %s", runID)
			}
		}
	}

	return nil
}

// normalizeAgentID converts agent display names to registry IDs
func (s *RunService) normalizeAgentID(agentID string) string {
	// Map common display names to registry IDs
	switch agentID {
	case "OtelAgent", "OTEL Management Agent":
		return "otel"
	case "CodingAgent", "Coding Agent":
		return "coding"
	default:
		return agentID
	}
}

// buildHistoryFromEvents converts message events into conversation history
func (s *RunService) buildHistoryFromEvents(runID string) ([]anthropic.MessageParam, error) {
	runEvents, err := s.storage.GetEvents(runID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	history := []anthropic.MessageParam{}

	for _, event := range runEvents {
		if event.Type != events.EventMessage {
			continue
		}

		var messageData events.MessageData
		if err := json.Unmarshal(event.Data, &messageData); err != nil {
			continue
		}

		// Only include user messages to avoid tool_use without results issues
		if messageData.Role != "user" {
			continue
		}

		// Build content blocks
		content := make([]anthropic.ContentBlockParamUnion, 0, len(messageData.Content))
		for _, block := range messageData.Content {
			if block.Type == "text" && block.Text != "" {
				textBlock := anthropic.TextBlockParam{Text: block.Text}
				blockJSON, err := json.Marshal(textBlock)
				if err != nil {
					continue
				}
				var paramBlock anthropic.ContentBlockParamUnion
				if err := json.Unmarshal(blockJSON, &paramBlock); err != nil {
					continue
				}
				content = append(content, paramBlock)
			}
		}

		if len(content) == 0 {
			continue
		}

		// Create user message
		msg := anthropic.NewUserMessage(content...)
		history = append(history, msg)
	}

	return history, nil
}
