package otel

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// GetAgentConfigInput represents the input for getting agent configuration
type GetAgentConfigInput struct {
	AgentID string `json:"agent_id"`
}

// GetGetAgentConfigTool creates a tool for getting agent configuration
func GetGetAgentConfigTool(client *otelclient.OtelClient) tools.Tool {
	return tools.Tool{
		Name:        "get_otel_agent_config",
		Description: "Retrieves the current YAML configuration for a specific OpenTelemetry collector agent. Use this to inspect the agent's configuration for debugging or before making updates.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the agent to get configuration for",
				},
			},
			Required: []string{"agent_id"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input GetAgentConfigInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			config, err := client.GetAgentConfig(input.AgentID)
			if err != nil {
				return nil, fmt.Errorf("failed to get agent config: %w", err)
			}

			return map[string]interface{}{
				"config_id":      config.ID,
				"config_name":    config.Name,
				"config_version": config.Version,
				"yaml_content":   config.Content,
			}, nil
		},
	}
}
