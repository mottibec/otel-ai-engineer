package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/config"
)

// AgentInfo contains metadata about an agent type
type AgentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
}

// AgentFactory is a function that creates an agent instance
type AgentFactory func(client *anthropic.Client, logLevel config.LogLevel, emitter events.EventEmitter) (*Agent, error)

// Registry manages available agent types
type Registry struct {
	mu        sync.RWMutex
	agents    map[string]AgentInfo
	factories map[string]AgentFactory
}

// NewRegistry creates a new agent registry
func NewRegistry() *Registry {
	r := &Registry{
		agents:    make(map[string]AgentInfo),
		factories: make(map[string]AgentFactory),
	}

	// Register built-in agents
	r.registerBuiltInAgents()

	return r
}

// registerBuiltInAgents registers all built-in agent types
func (r *Registry) registerBuiltInAgents() {
	// Register CodingAgent
	r.Register(AgentInfo{
		ID:          "coding",
		Name:        "Coding Agent",
		Description: "An AI agent specialized for coding tasks with file system access",
		Model:       string(anthropic.ModelClaudeSonnet4_5_20250929),
	}, func(client *anthropic.Client, logLevel config.LogLevel, emitter events.EventEmitter) (*Agent, error) {
		codingAgent, err := NewCodingAgent(client, logLevel)
		if err != nil {
			return nil, err
		}
		// Set the event emitter
		codingAgent.eventEmitter = emitter
		return codingAgent.Agent, nil
	})

	// Register OtelAgent
	r.Register(AgentInfo{
		ID:          "otel",
		Name:        "OTEL Management Agent",
		Description: "An AI agent specialized for OpenTelemetry collector management with file system and OTEL tools",
		Model:       string(anthropic.ModelClaudeSonnet4_5_20250929),
	}, func(client *anthropic.Client, logLevel config.LogLevel, emitter events.EventEmitter) (*Agent, error) {
		otelAgent, err := NewOtelAgent(client, logLevel)
		if err != nil {
			return nil, err
		}
		// Set the event emitter
		otelAgent.eventEmitter = emitter
		return otelAgent.Agent, nil
	})
}

// Register adds a new agent type to the registry
func (r *Registry) Register(info AgentInfo, factory AgentFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agents[info.ID] = info
	r.factories[info.ID] = factory
}

// List returns all available agent types
func (r *Registry) List() []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]AgentInfo, 0, len(r.agents))
	for _, info := range r.agents {
		agents = append(agents, info)
	}
	return agents
}

// Get retrieves agent info by ID
func (r *Registry) Get(id string) (AgentInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.agents[id]
	return info, ok
}

// Create creates a new agent instance by ID
func (r *Registry) Create(id string, client *anthropic.Client, logLevel config.LogLevel, emitter events.EventEmitter) (*Agent, error) {
	r.mu.RLock()
	factory, ok := r.factories[id]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("agent type not found: %s", id)
	}

	return factory(client, logLevel, emitter)
}

// Has checks if an agent type exists
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.agents[id]
	return ok
}

// RunnerConfig holds configuration for running an agent
type RunnerConfig struct {
	AgentID         string
	Prompt          string
	Client          *anthropic.Client
	LogLevel        config.LogLevel
	EventEmitter    events.EventEmitter
	PendingMessages chan string
	History         []anthropic.MessageParam
	RunID           string // Optional: existing run ID for resuming
}

// RunAgent is a helper function to create and run an agent in one call
func (r *Registry) RunAgent(ctx context.Context, cfg RunnerConfig) (*RunResult, error) {
	agent, err := r.Create(cfg.AgentID, cfg.Client, cfg.LogLevel, cfg.EventEmitter)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// If RunID is provided (resuming) or we have both history and pending messages,
	// use RunWithFullConfig which handles all cases
	if cfg.RunID != "" || (len(cfg.History) > 0 && cfg.PendingMessages != nil) {
		result := agent.RunWithFullConfig(ctx, cfg.Prompt, cfg.RunID, cfg.History, cfg.PendingMessages)
		return result, nil
	}

	// If history is provided, use RunWithHistory
	if len(cfg.History) > 0 {
		result := agent.RunWithHistory(ctx, cfg.Prompt, cfg.History)
		return result, nil
	}

	// If pending messages channel is provided, use RunWithPendingMessages
	if cfg.PendingMessages != nil {
		result := agent.RunWithPendingMessages(ctx, cfg.Prompt, cfg.PendingMessages)
		return result, nil
	}

	// Otherwise use standard Run
	result := agent.Run(ctx, cfg.Prompt)
	return result, nil
}
