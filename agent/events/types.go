package events

import (
	"encoding/json"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// EventType represents the type of agent event
type EventType string

const (
	EventRunStart    EventType = "run_start"
	EventRunEnd      EventType = "run_end"
	EventIteration   EventType = "iteration"
	EventMessage     EventType = "message"
	EventToolCall    EventType = "tool_call"
	EventToolResult  EventType = "tool_result"
	EventFileChange  EventType = "file_change"
	EventError       EventType = "error"
	EventAPIRequest  EventType = "api_request"
	EventAPIResponse EventType = "api_response"
)

// AgentEvent represents a single event in the agent's execution
type AgentEvent struct {
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	AgentID   string          `json:"agent_id"`
	AgentName string          `json:"agent_name"`
	RunID     string          `json:"run_id"`
	Type      EventType       `json:"type"`
	Data      json.RawMessage `json:"data"`
}

// RunStartData contains data for run start events
type RunStartData struct {
	Prompt       string `json:"prompt"`
	Model        string `json:"model"`
	MaxTokens    int64  `json:"max_tokens"`
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// RunEndData contains data for run end events
type RunEndData struct {
	Success         bool   `json:"success"`
	Error           string `json:"error,omitempty"`
	TotalToolCalls  int    `json:"total_tool_calls"`
	TotalIterations int    `json:"total_iterations"`
	Duration        string `json:"duration"`
}

// IterationData contains data for iteration events
type IterationData struct {
	Iteration     int `json:"iteration"`
	TotalMessages int `json:"total_messages"`
}

// MessageData contains data for message events
type MessageData struct {
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason,omitempty"`
	Model      string         `json:"model,omitempty"`
	Usage      *UsageInfo     `json:"usage,omitempty"`
}

// ContentBlock represents a content block in a message
type ContentBlock struct {
	Type       string          `json:"type"` // text, tool_use, tool_result
	Text       string          `json:"text,omitempty"`
	ToolUse    *ToolUseInfo    `json:"tool_use,omitempty"`
	ToolResult *ToolResultInfo `json:"tool_result,omitempty"`
}

// ToolUseInfo contains information about a tool use
type ToolUseInfo struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResultInfo contains information about a tool result
type ToolResultInfo struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// UsageInfo contains token usage information
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ToolCallData contains data for tool call events
type ToolCallData struct {
	ToolUseID string          `json:"tool_use_id"`
	ToolName  string          `json:"tool_name"`
	Input     json.RawMessage `json:"input"`
}

// ToolResultData contains data for tool result events
type ToolResultData struct {
	ToolUseID string          `json:"tool_use_id"`
	ToolName  string          `json:"tool_name"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
	IsError   bool            `json:"is_error"`
	Duration  string          `json:"duration"`
}

// FileChangeData contains data for file change events
type FileChangeData struct {
	Operation  string `json:"operation"` // read, write, edit, delete
	FilePath   string `json:"file_path"`
	OldContent string `json:"old_content,omitempty"`
	NewContent string `json:"new_content,omitempty"`
}

// ErrorData contains data for error events
type ErrorData struct {
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// APIRequestData contains data for API request events
type APIRequestData struct {
	Model     string `json:"model"`
	MaxTokens int64  `json:"max_tokens"`
	ToolCount int    `json:"tool_count"`
}

// APIResponseData contains data for API response events
type APIResponseData struct {
	StopReason   string     `json:"stop_reason"`
	Model        string     `json:"model"`
	Usage        *UsageInfo `json:"usage"`
	ContentCount int        `json:"content_count"`
}

// Helper functions to create events with typed data

// NewRunStartEvent creates a run start event
func NewRunStartEvent(runID, agentID, agentName string, data RunStartData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventRunStart,
		Data:      dataJSON,
	}, nil
}

// NewRunEndEvent creates a run end event
func NewRunEndEvent(runID, agentID, agentName string, data RunEndData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventRunEnd,
		Data:      dataJSON,
	}, nil
}

// NewIterationEvent creates an iteration event
func NewIterationEvent(runID, agentID, agentName string, data IterationData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventIteration,
		Data:      dataJSON,
	}, nil
}

// NewMessageEvent creates a message event
func NewMessageEvent(runID, agentID, agentName string, data MessageData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventMessage,
		Data:      dataJSON,
	}, nil
}

// NewToolCallEvent creates a tool call event
func NewToolCallEvent(runID, agentID, agentName string, data ToolCallData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventToolCall,
		Data:      dataJSON,
	}, nil
}

// NewToolResultEvent creates a tool result event
func NewToolResultEvent(runID, agentID, agentName string, data ToolResultData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventToolResult,
		Data:      dataJSON,
	}, nil
}

// NewAPIRequestEvent creates an API request event
func NewAPIRequestEvent(runID, agentID, agentName string, data APIRequestData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventAPIRequest,
		Data:      dataJSON,
	}, nil
}

// NewAPIResponseEvent creates an API response event
func NewAPIResponseEvent(runID, agentID, agentName string, data APIResponseData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventAPIResponse,
		Data:      dataJSON,
	}, nil
}

// NewErrorEvent creates an error event
func NewErrorEvent(runID, agentID, agentName string, data ErrorData) (*AgentEvent, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &AgentEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		RunID:     runID,
		Type:      EventError,
		Data:      dataJSON,
	}, nil
}

// Helper to convert Anthropic message to MessageData
func MessageDataFromAnthropic(msg *anthropic.Message) MessageData {
	content := make([]ContentBlock, 0, len(msg.Content))

	for _, block := range msg.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			content = append(content, ContentBlock{
				Type: "text",
				Text: v.Text,
			})
		case anthropic.ToolUseBlock:
			content = append(content, ContentBlock{
				Type: "tool_use",
				ToolUse: &ToolUseInfo{
					ID:    v.ID,
					Name:  v.Name,
					Input: v.Input,
				},
			})
		}
	}

	return MessageData{
		Role:       "assistant",
		Content:    content,
		StopReason: string(msg.StopReason),
		Model:      string(msg.Model),
		Usage: &UsageInfo{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
		},
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405.000000")
}
