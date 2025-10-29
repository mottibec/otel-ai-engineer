package agent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// HumanInputContext holds dependencies for human input requests
type HumanInputContext struct {
	RunID       string
	AgentID     string
	AgentName   string
	Storage     storage.Storage
	EventEmitter events.EventEmitter
	ResourceType *storage.ResourceType
	ResourceID   *string
	AgentWorkID  *string
}

// HumanInputRequest defines input for the request_human_input tool
type HumanInputRequest struct {
	RequestType string   `json:"request_type"` // "approval", "input", "decision", "information"
	Question    string   `json:"question"`    // The question or request
	Context     string   `json:"context"`     // Additional context
	Options     []string `json:"options,omitempty"` // Optional: predefined choices
}

// CreateHumanInputTool creates the request_human_input tool with context
func CreateHumanInputTool(ctx *HumanInputContext) tools.Tool {
	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"request_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of request: 'approval' (yes/no), 'input' (text), 'decision' (choose from options), 'information' (provide info)",
				"enum":        []string{"approval", "input", "decision", "information"},
			},
			"question": map[string]interface{}{
				"type":        "string",
				"description": "The question or request you need from the human",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "Additional context about why you need this input",
			},
			"options": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional: Predefined options for 'decision' type requests",
			},
		},
		Required: []string{"request_type", "question", "context"},
	}

	handler := func(inputJSON json.RawMessage) (interface{}, error) {
		var input HumanInputRequest
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			return nil, fmt.Errorf("failed to parse human input request: %w", err)
		}

		return executeHumanInputRequest(ctx, input)
	}

	return tools.Tool{
		Name:        "request_human_input",
		Description: "Request manual intervention or input from a human. The agent will pause until a human responds. Use this when you need approval, additional information, or a decision that requires human judgment.",
		Schema:      schema,
		Handler:     handler,
	}
}

// executeHumanInputRequest creates a human action and pauses the agent
func executeHumanInputRequest(ctx *HumanInputContext, input HumanInputRequest) (interface{}, error) {
	// Generate action ID
	actionID := fmt.Sprintf("action-%d", time.Now().UnixNano())

	// Create human action entry
	action := &storage.HumanAction{
		ID:           actionID,
		RunID:        ctx.RunID,
		AgentID:     ctx.AgentID,
		AgentName:   ctx.AgentName,
		ResourceType: ctx.ResourceType,
		ResourceID:  ctx.ResourceID,
		AgentWorkID: ctx.AgentWorkID,
		RequestType: input.RequestType,
		Question:    input.Question,
		Context:     input.Context,
		Options:     input.Options,
		Status:      storage.HumanActionStatusPending,
	}

	if err := ctx.Storage.CreateHumanAction(action); err != nil {
		return nil, fmt.Errorf("failed to create human action: %w", err)
	}

	// Emit event
	if evt, err := events.NewMessageEvent(ctx.RunID, ctx.AgentID, ctx.AgentName, events.MessageData{
		Role: "assistant",
		Content: []events.ContentBlock{
			{
				Type: "text",
				Text: fmt.Sprintf("🤚 **Human input requested**: %s\n\n%s\n\nAwaiting human response...", input.Question, input.Context),
			},
		},
	}); err == nil {
		ctx.EventEmitter.Emit(evt)
	}

	// Return information about the pending request
	// The agent will need to wait (via resume) for the response
	return map[string]interface{}{
		"action_id":    actionID,
		"status":       "pending",
		"message":      "Human input requested. The agent will pause and wait for a human response. Use the resume endpoint with the response to continue.",
	}, nil
}

