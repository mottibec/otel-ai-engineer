package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// Agent represents an AI agent with a specific set of tools and capabilities
type Agent struct {
	name         string
	description  string
	client       *anthropic.Client
	registry     *tools.ToolRegistry
	model        anthropic.Model
	maxTokens    int64
	systemPrompt string
	logger       *Logger
	eventEmitter events.EventEmitter
}

// Config holds the configuration for creating an agent
type Config struct {
	Name         string
	Description  string
	Client       *anthropic.Client
	Model        anthropic.Model
	MaxTokens    int64
	SystemPrompt string
	LogLevel     config.LogLevel
	EventEmitter events.EventEmitter
	Tools        []tools.Tool
}

// NewAgent creates a new agent with the given configuration
func NewAgent(cfg Config) *Agent {
	if cfg.Model == "" {
		cfg.Model = anthropic.ModelClaudeSonnet4_5_20250929
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}

	// Use NoOpEmitter if no emitter is provided
	emitter := cfg.EventEmitter
	if emitter == nil {
		emitter = events.NewNoOpEmitter()
	}

	agent := &Agent{
		name:         cfg.Name,
		description:  cfg.Description,
		client:       cfg.Client,
		registry:     tools.NewRegistry(),
		model:        cfg.Model,
		maxTokens:    cfg.MaxTokens,
		systemPrompt: cfg.SystemPrompt,
		logger:       NewLogger(cfg.LogLevel),
		eventEmitter: emitter,
	}

	// Register tools if provided
	if len(cfg.Tools) > 0 {
		agent.registry.RegisterTools(cfg.Tools)
	}

	return agent
}

// RegisterTool adds a tool to the agent's registry
func (a *Agent) RegisterTool(
	name string,
	description string,
	schema anthropic.ToolInputSchemaParam,
	handler interface{},
) error {
	return tools.RegisterFunc(a.registry, name, description, schema, handler)
}

// AddTools registers multiple tools from another registry
func (a *Agent) AddTools(registry *tools.ToolRegistry) {
	// Get all tools from the source registry and register them
	for _, toolName := range registry.ListTools() {
		// Note: This is a simplified version. In production, you'd want to
		// properly copy the tool definitions
		if a.registry.HasTool(toolName) {
			continue // Skip if already registered
		}
		// The actual implementation would need access to the internal tools map
		// For now, tools should be registered individually
	}
}

// RunResult contains the result of running an agent
type RunResult struct {
	Success      bool
	FinalMessage *anthropic.Message
	Messages     []anthropic.MessageParam
	Error        error
	ToolCalls    int
	Iterations   int
}

// Run executes the agent with the given prompt until completion
func (a *Agent) Run(ctx context.Context, prompt string) *RunResult {
	return a.runInternal(ctx, prompt, "", nil, nil)
}

// RunWithHistory executes the agent with conversation history
func (a *Agent) RunWithHistory(ctx context.Context, prompt string, history []anthropic.MessageParam) *RunResult {
	return a.runInternal(ctx, prompt, "", history, nil)
}

// RunWithPendingMessages executes the agent with support for pending messages from a channel
func (a *Agent) RunWithPendingMessages(ctx context.Context, prompt string, pendingMessages chan string) *RunResult {
	return a.runInternal(ctx, prompt, "", nil, pendingMessages)
}

// RunWithFullConfig executes the agent with full configuration support including:
// - Optional existing run ID (for resuming)
// - Optional conversation history
// - Optional pending messages channel
func (a *Agent) RunWithFullConfig(ctx context.Context, prompt string, existingRunID string, history []anthropic.MessageParam, pendingMessages chan string) *RunResult {
	return a.runInternal(ctx, prompt, existingRunID, history, pendingMessages)
}

// runInternal is the core run method that all public run methods delegate to
func (a *Agent) runInternal(ctx context.Context, prompt string, existingRunID string, history []anthropic.MessageParam, pendingMessages chan string) *RunResult {
	// Start with history, then add the new prompt (if not empty)
	messages := make([]anthropic.MessageParam, 0, len(history)+1)
	messages = append(messages, history...)
	// Only add the prompt if it's not empty (empty prompt is used for resuming with existing history)
	if prompt != "" {
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))
	}

	result := &RunResult{
		Messages: messages,
	}

	// Use existing run ID if provided, otherwise generate a new one
	var runID string
	if existingRunID != "" {
		runID = existingRunID
	} else {
		runID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}

	startTime := time.Now()

	// Only emit run start event if this is a new run (no existing run ID)
	if existingRunID == "" {
		if evt, err := events.NewRunStartEvent(runID, a.name, a.name, events.RunStartData{
			Prompt:       prompt,
			Model:        string(a.model),
			MaxTokens:    a.maxTokens,
			SystemPrompt: a.systemPrompt,
		}); err == nil {
			a.eventEmitter.Emit(evt)
		}
	}

	maxIterations := 50 // Prevent infinite loops

	for result.Iterations = 0; result.Iterations < maxIterations; result.Iterations++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			a.logger.Info("Run cancelled by user")
			result.Error = ctx.Err()
			result.Success = false

			// Emit cancelled event
			if evt, err := events.NewRunEndEvent(runID, a.name, a.name, events.RunEndData{
				Success:         false,
				Error:           "Run was cancelled by user",
				TotalToolCalls:  result.ToolCalls,
				TotalIterations: result.Iterations,
				Duration:        time.Since(startTime).String(),
			}); err == nil {
				a.eventEmitter.Emit(evt)
			}

			return result
		default:
		}

		// Check for pending messages if channel is provided
		if pendingMessages != nil {
			select {
			case pendingMsg := <-pendingMessages:
				a.logger.Info("Received pending message from channel")
				// Add the pending message to the conversation
				messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(pendingMsg)))
				result.Messages = messages

				// Emit message event for the pending message
				if evt, err := events.NewMessageEvent(runID, a.name, a.name, events.MessageData{
					Role: "user",
					Content: []events.ContentBlock{
						{Type: "text", Text: pendingMsg},
					},
				}); err == nil {
					a.eventEmitter.Emit(evt)
				}
			default:
				// No pending messages right now, continue
			}
		}

		// Log iteration start
		a.logger.LogIteration(result.Iterations+1, len(result.Messages))

		// Emit iteration event
		if evt, err := events.NewIterationEvent(runID, a.name, a.name, events.IterationData{
			Iteration:     result.Iterations + 1,
			TotalMessages: len(result.Messages),
		}); err == nil {
			a.eventEmitter.Emit(evt)
		}

		// Create message params
		params := anthropic.MessageNewParams{
			Model:     a.model,
			MaxTokens: a.maxTokens,
			Messages:  result.Messages,
			Tools:     a.registry.GetToolUnionParams(),
		}

		// Add system prompt if provided
		if a.systemPrompt != "" {
			params.System = []anthropic.TextBlockParam{
				{Text: a.systemPrompt},
			}
		}

		// Log API request
		a.logger.LogAPIRequest(a.model, a.maxTokens, len(a.registry.ListTools()))

		// Emit API request event
		if evt, err := events.NewAPIRequestEvent(runID, a.name, a.name, events.APIRequestData{
			Model:     string(a.model),
			MaxTokens: a.maxTokens,
			ToolCount: len(a.registry.ListTools()),
		}); err == nil {
			a.eventEmitter.Emit(evt)
		}

		// Call Claude API
		message, err := a.client.Messages.New(ctx, params)
		if err != nil {
			a.logger.Error("API call failed: %v", err)
			result.Error = fmt.Errorf("API call failed: %w", err)

			// Parse the error for better display
			errorMessage := a.parseAPIError(err)

			// Emit error event
			if evt, err := events.NewErrorEvent(runID, a.name, a.name, events.ErrorData{
				Message: errorMessage,
			}); err == nil {
				a.eventEmitter.Emit(evt)
			}

			// Emit run end event
			if evt, err := events.NewRunEndEvent(runID, a.name, a.name, events.RunEndData{
				Success:         false,
				Error:           result.Error.Error(),
				TotalToolCalls:  result.ToolCalls,
				TotalIterations: result.Iterations + 1,
				Duration:        time.Since(startTime).String(),
			}); err == nil {
				a.eventEmitter.Emit(evt)
			}

			return result
		}

		// Log API response
		a.logger.LogAPIResponse(message)

		// Emit API response event
		if evt, err := events.NewAPIResponseEvent(runID, a.name, a.name, events.APIResponseData{
			StopReason: string(message.StopReason),
			Model:      string(message.Model),
			Usage: &events.UsageInfo{
				InputTokens:  int(message.Usage.InputTokens),
				OutputTokens: int(message.Usage.OutputTokens),
			},
			ContentCount: len(message.Content),
		}); err == nil {
			a.eventEmitter.Emit(evt)
		}

		// Emit message event
		if evt, err := events.NewMessageEvent(runID, a.name, a.name, events.MessageDataFromAnthropic(message)); err == nil {
			a.eventEmitter.Emit(evt)
		}

		result.FinalMessage = message

		// Check stop reason
		a.logger.LogStopReason(string(message.StopReason))

		switch message.StopReason {
		case "end_turn":
			// Task completed successfully
			result.Success = true
			a.logger.LogTaskComplete(result.Iterations+1, result.ToolCalls)

			// Emit run end event
			if evt, err := events.NewRunEndEvent(runID, a.name, a.name, events.RunEndData{
				Success:         true,
				TotalToolCalls:  result.ToolCalls,
				TotalIterations: result.Iterations + 1,
				Duration:        time.Since(startTime).String(),
			}); err == nil {
				a.eventEmitter.Emit(evt)
			}

			return result

		case "tool_use":
			// Execute tools
			toolResults, err := a.executeTools(message, runID)
			if err != nil {
				a.logger.Error("Tool execution failed: %v", err)
				result.Error = fmt.Errorf("tool execution failed: %w", err)
				a.logger.LogTaskFailed(result.Error, result.Iterations+1)

				// Emit error event
				if evt, err := events.NewErrorEvent(runID, a.name, a.name, events.ErrorData{
					Message: result.Error.Error(),
				}); err == nil {
					a.eventEmitter.Emit(evt)
				}

				// Emit run end event
				if evt, err := events.NewRunEndEvent(runID, a.name, a.name, events.RunEndData{
					Success:         false,
					Error:           result.Error.Error(),
					TotalToolCalls:  result.ToolCalls,
					TotalIterations: result.Iterations + 1,
					Duration:        time.Since(startTime).String(),
				}); err == nil {
					a.eventEmitter.Emit(evt)
				}

				return result
			}

			result.ToolCalls += len(toolResults)

			// Add assistant message with tool use, then tool results
			assistantContent := make([]anthropic.ContentBlockParamUnion, 0, len(message.Content))
			for _, block := range message.Content {
				blockJSON, err := json.Marshal(block)
				if err != nil {
					result.Error = fmt.Errorf("failed to marshal content block: %w", err)
					return result
				}

				var paramBlock anthropic.ContentBlockParamUnion
				if err := json.Unmarshal(blockJSON, &paramBlock); err != nil {
					result.Error = fmt.Errorf("failed to unmarshal content block: %w", err)
					return result
				}

				assistantContent = append(assistantContent, paramBlock)
			}

			result.Messages = append(result.Messages, anthropic.NewAssistantMessage(assistantContent...))

			// Add tool results
			result.Messages = append(result.Messages, toolResults...)

		case "max_tokens":
			result.Error = fmt.Errorf("reached max tokens")

			// Emit run end event
			if evt, err := events.NewRunEndEvent(runID, a.name, a.name, events.RunEndData{
				Success:         false,
				Error:           result.Error.Error(),
				TotalToolCalls:  result.ToolCalls,
				TotalIterations: result.Iterations + 1,
				Duration:        time.Since(startTime).String(),
			}); err == nil {
				a.eventEmitter.Emit(evt)
			}

			return result

		default:
			result.Error = fmt.Errorf("unexpected stop reason: %s", message.StopReason)

			// Emit run end event
			if evt, err := events.NewRunEndEvent(runID, a.name, a.name, events.RunEndData{
				Success:         false,
				Error:           result.Error.Error(),
				TotalToolCalls:  result.ToolCalls,
				TotalIterations: result.Iterations + 1,
				Duration:        time.Since(startTime).String(),
			}); err == nil {
				a.eventEmitter.Emit(evt)
			}

			return result
		}
	}

	result.Error = fmt.Errorf("reached maximum iterations (%d)", maxIterations)

	// Emit run end event
	if evt, err := events.NewRunEndEvent(runID, a.name, a.name, events.RunEndData{
		Success:         false,
		Error:           result.Error.Error(),
		TotalToolCalls:  result.ToolCalls,
		TotalIterations: result.Iterations,
		Duration:        time.Since(startTime).String(),
	}); err == nil {
		a.eventEmitter.Emit(evt)
	}

	return result
}

// executeTools processes all tool use blocks in a message
func (a *Agent) executeTools(message *anthropic.Message, runID string) ([]anthropic.MessageParam, error) {
	var toolResultBlocks []anthropic.ContentBlockParamUnion

	for _, block := range message.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.ToolUseBlock:
			// Log tool execution start
			a.logger.LogToolExecution(variant.Name, variant.Input)

			// Emit tool call event
			if evt, err := events.NewToolCallEvent(runID, a.name, a.name, events.ToolCallData{
				ToolUseID: variant.ID,
				ToolName:  variant.Name,
				Input:     variant.Input,
			}); err == nil {
				a.eventEmitter.Emit(evt)
			}

			// Execute the tool using the registry
			toolStartTime := time.Now()
			result, err := a.registry.Execute(variant.Name, variant.Input)
			toolDuration := time.Since(toolStartTime)

			// Log tool result
			a.logger.LogToolResult(variant.Name, result, err)

			// Emit tool result event
			var resultJSON json.RawMessage
			if result != nil {
				resultJSON, _ = json.Marshal(result)
			}

			if evt, toolErr := events.NewToolResultEvent(runID, a.name, a.name, events.ToolResultData{
				ToolUseID: variant.ID,
				ToolName:  variant.Name,
				Result:    resultJSON,
				Error: func() string {
					if err != nil {
						return err.Error()
					}
					return ""
				}(),
				IsError:  err != nil,
				Duration: toolDuration.String(),
			}); toolErr == nil {
				a.eventEmitter.Emit(evt)
			}

			if err != nil {
				// Return error result
				toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(
					variant.ID,
					fmt.Sprintf("Error: %v", err),
					true, // is_error
				))
			} else {
				// Marshal result to JSON
				resultJSON, err := json.Marshal(result)
				if err != nil {
					toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(
						variant.ID,
						fmt.Sprintf("Error marshaling result: %v", err),
						true,
					))
				} else {
					toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(
						variant.ID,
						string(resultJSON),
						false,
					))
				}
			}
		}
	}

	return []anthropic.MessageParam{anthropic.NewUserMessage(toolResultBlocks...)}, nil
}

// parseAPIError parses Anthropic API errors to extract meaningful information
func (a *Agent) parseAPIError(err error) string {
	errStr := err.Error()
	
	// Check if this is a structured error (likely from Anthropic SDK)
	// The error format from the SDK often includes detailed information
	
	// Rate limit errors
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate_limit") {
		return errStr
	}
	
	// Authentication errors
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "authentication") {
		return "API authentication failed. Please check your API key."
	}
	
	// API error format from Anthropic SDK
	if strings.Contains(errStr, "API call failed:") {
		// Return the full error as it contains structured information
		return errStr
	}
	
	// Default: return the full error
	return fmt.Sprintf("API call failed: %v", err)
}

// GetName returns the agent's name
func (a *Agent) GetName() string {
	return a.name
}

// GetDescription returns the agent's description
func (a *Agent) GetDescription() string {
	return a.description
}

// GetRegistry returns the agent's tool registry
func (a *Agent) GetRegistry() *tools.ToolRegistry {
	return a.registry
}

// ListTools returns all tools available to this agent
func (a *Agent) ListTools() []string {
	return a.registry.ListTools()
}
