package humanaction

import (
	"context"
	"fmt"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/server/service"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// HumanActionService handles business logic for human action management
type HumanActionService struct {
	storage       storage.Storage
	runService    *service.RunService
}

// NewHumanActionService creates a new human action service
func NewHumanActionService(stor storage.Storage, runService *service.RunService) *HumanActionService {
	return &HumanActionService{
		storage:    stor,
		runService: runService,
	}
}

// ListHumanActions lists human actions with optional filtering
func (has *HumanActionService) ListHumanActions(ctx context.Context, opts storage.HumanActionListOptions) ([]*storage.HumanAction, error) {
	actions, err := has.storage.ListHumanActions(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list human actions: %w", err)
	}
	return actions, nil
}

// GetPendingHumanActions gets all pending human actions
func (has *HumanActionService) GetPendingHumanActions(ctx context.Context) ([]*storage.HumanAction, error) {
	actions, err := has.storage.GetPendingHumanActions()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending human actions: %w", err)
	}
	return actions, nil
}

// GetHumanAction retrieves a human action by ID
func (has *HumanActionService) GetHumanAction(ctx context.Context, actionID string) (*storage.HumanAction, error) {
	if actionID == "" {
		return nil, fmt.Errorf("action ID cannot be empty")
	}

	action, err := has.storage.GetHumanAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get human action: %w", err)
	}
	return action, nil
}

// RespondToHumanAction responds to a human action
func (has *HumanActionService) RespondToHumanAction(ctx context.Context, actionID string, req RespondToHumanActionRequest) (*storage.HumanAction, error) {
	if actionID == "" {
		return nil, fmt.Errorf("action ID cannot be empty")
	}

	if req.Response == "" {
		return nil, fmt.Errorf("response is required")
	}

	// Get the action
	action, err := has.storage.GetHumanAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get human action: %w", err)
	}

	if action.Status != storage.HumanActionStatusPending {
		return nil, fmt.Errorf("human action is not pending")
	}

	// Update action with response
	now := time.Now()
	status := storage.HumanActionStatusResponded
	update := &storage.HumanActionUpdate{
		Status:      &status,
		Response:    &req.Response,
		RespondedAt: &now,
	}

	if err := has.storage.UpdateHumanAction(actionID, update); err != nil {
		return nil, fmt.Errorf("failed to update human action: %w", err)
	}

	// If resume is requested, resume the run with the response
	if req.Resume {
		resumeStatus := storage.HumanActionStatusResumed
		resumeUpdate := &storage.HumanActionUpdate{
			Status:    &resumeStatus,
			ResumedAt: &now,
		}
		_ = has.storage.UpdateHumanAction(actionID, resumeUpdate) // Log error but don't fail

		// Resume the run with the human's response as a message
		_, err := has.runService.ResumeRun(ctx, service.ResumeRunRequest{
			RunID:   action.RunID,
			Message: fmt.Sprintf("Human response to '%s': %s", action.Question, req.Response),
		})
		if err != nil {
			// Log error but don't fail the request
			_ = err
		}
	}

	// Get updated action
	updatedAction, err := has.storage.GetHumanAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated human action: %w", err)
	}

	return updatedAction, nil
}

// ResumeFromHumanAction resumes a run from a human action
func (has *HumanActionService) ResumeFromHumanAction(ctx context.Context, actionID string) (*storage.HumanAction, error) {
	if actionID == "" {
		return nil, fmt.Errorf("action ID cannot be empty")
	}

	// Get the action
	action, err := has.storage.GetHumanAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get human action: %w", err)
	}

	if action.Status != storage.HumanActionStatusResponded {
		return nil, fmt.Errorf("human action must be responded to before resuming")
	}

	if action.Response == nil || *action.Response == "" {
		return nil, fmt.Errorf("no response found on human action")
	}

	// Mark as resumed
	now := time.Now()
	status := storage.HumanActionStatusResumed
	update := &storage.HumanActionUpdate{
		Status:    &status,
		ResumedAt: &now,
	}
	if err := has.storage.UpdateHumanAction(actionID, update); err != nil {
		return nil, fmt.Errorf("failed to update human action: %w", err)
	}

	// Resume the run with the human's response
	_, err = has.runService.ResumeRun(ctx, service.ResumeRunRequest{
		RunID:   action.RunID,
		Message: fmt.Sprintf("Human response to '%s': %s", action.Question, *action.Response),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resume run: %w", err)
	}

	// Get updated action
	updatedAction, err := has.storage.GetHumanAction(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated human action: %w", err)
	}

	return updatedAction, nil
}

// DeleteHumanAction deletes a human action
func (has *HumanActionService) DeleteHumanAction(ctx context.Context, actionID string) error {
	if actionID == "" {
		return fmt.Errorf("action ID cannot be empty")
	}

	if err := has.storage.DeleteHumanAction(actionID); err != nil {
		return fmt.Errorf("failed to delete human action: %w", err)
	}

	return nil
}

