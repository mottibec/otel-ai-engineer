package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// HandoffInput defines input for the handoff tool
type HandoffInput struct {
	ToAgentID       string `json:"to_agent_id"`
	TaskDescription string `json:"task_description"`
}

// HandoffContext holds dependencies for handoff execution
type HandoffContext struct {
	ParentRunID   string
	ParentAgentID string
	Registry      *Registry
	Client        *anthropic.Client
	LogLevel      config.LogLevel
	EventEmitter  events.EventEmitter
	Storage       storage.Storage
}

// CreateHandoffTool creates the handoff tool with context
func CreateHandoffTool(ctx *HandoffContext) tools.Tool {
	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"to_agent_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the agent to delegate this task to (e.g., 'coding', 'otel')",
			},
			"task_description": map[string]interface{}{
				"type":        "string",
				"description": "Clear description of the task to delegate to the sub-agent",
			},
		},
		Required: []string{"to_agent_id", "task_description"},
	}

	handler := func(inputJSON json.RawMessage) (interface{}, error) {
		var input HandoffInput
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			return nil, fmt.Errorf("failed to parse handoff input: %w", err)
		}

		return executeHandoff(ctx, input)
	}

	return tools.Tool{
		Name:        "handoff_task",
		Description: "Delegate a task to another specialized agent. The current agent will pause until the sub-agent completes. Use this when a task requires expertise from a different agent type.",
		Schema:      schema,
		Handler:     handler,
	}
}

// executeHandoff runs the sub-agent and returns results
func executeHandoff(ctx *HandoffContext, input HandoffInput) (interface{}, error) {
	// 1. Validate target agent exists
	if !ctx.Registry.Has(input.ToAgentID) {
		return nil, fmt.Errorf("agent not found: %s", input.ToAgentID)
	}

	// 2. Generate sub-run ID
	subRunID := fmt.Sprintf("run-%d", time.Now().UnixNano())

	// 3. Emit handoff start event
	if evt, err := events.NewAgentHandoffEvent(ctx.ParentRunID, ctx.ParentAgentID,
		ctx.ParentAgentID, events.HandoffData{
		ParentRunID:     ctx.ParentRunID,
		SubRunID:        subRunID,
		FromAgentID:     ctx.ParentAgentID,
		ToAgentID:       input.ToAgentID,
		TaskDescription: input.TaskDescription,
	}); err == nil {
		ctx.EventEmitter.Emit(evt)
	}

	// 4. Create and run sub-agent (blocking)
	startTime := time.Now()
	result, err := ctx.Registry.RunAgent(context.Background(), RunnerConfig{
		AgentID:      input.ToAgentID,
		Prompt:       input.TaskDescription,
		Client:       ctx.Client,
		LogLevel:     ctx.LogLevel,
		EventEmitter: ctx.EventEmitter,
		RunID:        subRunID,
		Storage:      ctx.Storage,
	})

	duration := time.Since(startTime)

	// 5. Extract summary from result
	summary := extractSummary(result)
	success := result != nil && result.Success

	// 6. Update parent run with sub-run reference
	if ctx.Storage != nil {
		updateParentWithSubRun(ctx.Storage, ctx.ParentRunID, subRunID)
	}

	// 7. Emit handoff complete event
	if evt, e := events.NewAgentHandoffCompleteEvent(ctx.ParentRunID, ctx.ParentAgentID,
		ctx.ParentAgentID, events.HandoffCompleteData{
		ParentRunID: ctx.ParentRunID,
		SubRunID:    subRunID,
		FromAgentID: ctx.ParentAgentID,
		ToAgentID:   input.ToAgentID,
		Success:     success,
		Summary:     summary,
		Error:       getErrorString(result, err),
		Duration:    duration.String(),
	}); e == nil {
		ctx.EventEmitter.Emit(evt)
	}

	// 8. Return summary to parent agent
	return map[string]interface{}{
		"success":    success,
		"summary":    summary,
		"sub_run_id": subRunID,
		"agent_id":   input.ToAgentID,
		"duration":   duration.String(),
		"error":      getErrorString(result, err),
	}, nil
}

// extractSummary extracts a summary from the sub-agent's final result
func extractSummary(result *RunResult) string {
	// Extract final text message from assistant
	if result == nil || result.FinalMessage == nil {
		return "No response"
	}

	for _, block := range result.FinalMessage.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			return textBlock.Text
		}
	}

	return "Task completed"
}

// getErrorString extracts error information
func getErrorString(result *RunResult, err error) string {
	if err != nil {
		return err.Error()
	}
	if result != nil && result.Error != nil {
		return result.Error.Error()
	}
	return ""
}

// updateParentWithSubRun updates the parent run to include the sub-run ID
func updateParentWithSubRun(stor storage.Storage, parentRunID, subRunID string) {
	// Fetch parent, update sub_run_ids, save
	parent, err := stor.GetRun(parentRunID)
	if err != nil {
		return
	}

	// Append sub_run_id if not already in list
	subRunIDs := make([]string, len(parent.SubRunIDs))
	copy(subRunIDs, parent.SubRunIDs)
	subRunIDs = append(subRunIDs, subRunID)

	update := &storage.RunUpdate{
		SubRunIDs: &subRunIDs,
	}
	stor.UpdateRun(parentRunID, update)
}

