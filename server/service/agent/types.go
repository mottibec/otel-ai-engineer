package agent

import (
	"github.com/mottibechhofer/otel-ai-engineer/server/service/tools"
)

// AgentResponse represents an agent with its tools
type AgentResponse struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Model       string                    `json:"model,omitempty"`
	SystemPrompt string                   `json:"system_prompt,omitempty"`
	MaxTokens   int64                     `json:"max_tokens,omitempty"`
	Type        string                    `json:"type"` // "built-in" or "custom"
	Tools       []tools.ToolInfo          `json:"tools,omitempty"`
	ToolNames   []string                  `json:"tool_names,omitempty"` // For custom agents
	CreatedAt   string                    `json:"created_at,omitempty"`
	UpdatedAt   string                    `json:"updated_at,omitempty"`
}

// CreateCustomAgentRequest represents the request to create a custom agent
type CreateCustomAgentRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
	Model        string   `json:"model,omitempty"`
	MaxTokens    int64    `json:"max_tokens,omitempty"`
	ToolNames    []string `json:"tool_names"`
}

// UpdateCustomAgentRequest represents the request to update a custom agent
type UpdateCustomAgentRequest struct {
	Name         *string   `json:"name,omitempty"`
	Description  *string   `json:"description,omitempty"`
	SystemPrompt *string   `json:"system_prompt,omitempty"`
	Model        *string   `json:"model,omitempty"`
	MaxTokens    *int64    `json:"max_tokens,omitempty"`
	ToolNames    *[]string `json:"tool_names,omitempty"`
}

// CreateMetaAgentRequest represents the request to create a meta-agent
type CreateMetaAgentRequest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	SystemPrompt      string   `json:"system_prompt,omitempty"`
	Model             string   `json:"model,omitempty"`
	AvailableToolNames []string `json:"available_tool_names"` // Tools the meta-agent can use
}

