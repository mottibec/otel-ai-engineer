package otel

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// UpdateAgentConfigInput represents the input for updating agent configuration
type UpdateAgentConfigInput struct {
	AgentID    string `json:"agent_id"`
	YAMLConfig string `json:"yaml_config"`
}

// GetUpdateAgentConfigTool creates a tool for updating agent configuration
func GetUpdateAgentConfigTool(client *otelclient.OtelClient) tools.Tool {
	return tools.Tool{
		Name:        "update_otel_agent_config",
		Description: "Updates the configuration for a specific OpenTelemetry collector agent. The new configuration will be sent to the agent via OpAMP protocol. The agent must support remote configuration capability.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the agent to update",
				},
				"yaml_config": map[string]interface{}{
					"type":        "string",
					"description": "The new YAML configuration content",
				},
			},
			Required: []string{"agent_id", "yaml_config"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input UpdateAgentConfigInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			err := client.UpdateAgentConfig(input.AgentID, input.YAMLConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to update agent config: %w", err)
			}

			return map[string]interface{}{
				"success":  true,
				"agent_id": input.AgentID,
				"message":  fmt.Sprintf("Configuration successfully sent to agent %s. The agent should apply the new configuration within 30 seconds.", input.AgentID),
			}, nil
		},
	}
}
