package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// AgentWorkService handles business logic for agent work tracking
type AgentWorkService struct {
	storage storage.Storage
}

// NewAgentWorkService creates a new agent work service
func NewAgentWorkService(stor storage.Storage) *AgentWorkService {
	return &AgentWorkService{
		storage: stor,
	}
}

// CreateAgentWork creates a new agent work entry
func (aws *AgentWorkService) CreateAgentWork(ctx context.Context, work *storage.AgentWork) error {
	if work == nil {
		return fmt.Errorf("agent work cannot be nil")
	}
	if work.ResourceType == "" {
		return fmt.Errorf("resource type cannot be empty")
	}
	if work.ResourceID == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}
	if work.RunID == "" {
		return fmt.Errorf("run ID cannot be empty")
	}
	if work.AgentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Generate ID if not provided
	if work.ID == "" {
		work.ID = fmt.Sprintf("work-%d", time.Now().UnixNano())
	}

	// Set defaults
	if work.Status == "" {
		work.Status = storage.AgentWorkStatusRunning
	}
	if work.StartedAt.IsZero() {
		work.StartedAt = time.Now()
	}

	return aws.storage.CreateAgentWork(work)
}

// GetAgentWork retrieves agent work by ID
func (aws *AgentWorkService) GetAgentWork(ctx context.Context, workID string) (*storage.AgentWork, error) {
	if workID == "" {
		return nil, fmt.Errorf("work ID cannot be empty")
	}

	return aws.storage.GetAgentWork(workID)
}

// GetAgentWorkByResource retrieves agent work for a specific resource
func (aws *AgentWorkService) GetAgentWorkByResource(ctx context.Context, resourceType storage.ResourceType, resourceID string) ([]*storage.AgentWork, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resource type cannot be empty")
	}
	if resourceID == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}

	return aws.storage.GetAgentWorkByResource(resourceType, resourceID)
}

// ListAgentWork lists agent work with optional filtering
func (aws *AgentWorkService) ListAgentWork(ctx context.Context, opts storage.AgentWorkListOptions) ([]*storage.AgentWork, error) {
	return aws.storage.ListAgentWork(opts)
}

// UpdateAgentWorkStatus updates the status of agent work, typically called when a run completes
func (aws *AgentWorkService) UpdateAgentWorkStatus(ctx context.Context, runID string, status storage.AgentWorkStatus, errMsg string) error {
	// Find agent work by run ID
	allWork, err := aws.storage.ListAgentWork(storage.AgentWorkListOptions{
		Limit:  1000, // Get all to find by run_id
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("failed to list agent work: %w", err)
	}

	var workToUpdate *storage.AgentWork
	for _, work := range allWork {
		if work.RunID == runID {
			workToUpdate = work
			break
		}
	}

	if workToUpdate == nil {
		// No agent work found for this run, that's okay
		return nil
	}

	update := &storage.AgentWorkUpdate{
		Status: &status,
	}
	if status != storage.AgentWorkStatusRunning {
		now := time.Now()
		update.CompletedAt = &now
	}
	if errMsg != "" {
		update.Error = &errMsg
	}

	return aws.storage.UpdateAgentWork(workToUpdate.ID, update)
}

// UpdateAgentWork updates agent work
func (aws *AgentWorkService) UpdateAgentWork(ctx context.Context, workID string, update *storage.AgentWorkUpdate) error {
	if workID == "" {
		return fmt.Errorf("work ID cannot be empty")
	}
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	return aws.storage.UpdateAgentWork(workID, update)
}

// DeleteAgentWork deletes agent work
func (aws *AgentWorkService) DeleteAgentWork(ctx context.Context, workID string) error {
	if workID == "" {
		return fmt.Errorf("work ID cannot be empty")
	}

	return aws.storage.DeleteAgentWork(workID)
}

// GetActiveWorkForResource gets the currently active (running) agent work for a resource
func (aws *AgentWorkService) GetActiveWorkForResource(ctx context.Context, resourceType storage.ResourceType, resourceID string) (*storage.AgentWork, error) {
	running := storage.AgentWorkStatusRunning
	works, err := aws.storage.ListAgentWork(storage.AgentWorkListOptions{
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Status:       &running,
		Limit:        1,
		Offset:       0,
	})
	if err != nil {
		return nil, err
	}

	if len(works) > 0 {
		return works[0], nil
	}

	return nil, nil // No active work
}

