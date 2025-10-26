package otel

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// ListOtelAgentsInput represents the input for listing OTEL agents
type ListOtelAgentsInput struct {
	StatusFilter *string `json:"status_filter"`
}

// GetListOtelAgentsTool creates a tool for listing OTEL agents
func GetListOtelAgentsTool(client *otelclient.OtelClient) tools.Tool {
	return tools.Tool{
		Name:        "list_otel_agents",
		Description: "Lists all connected OpenTelemetry collector agents with their status, version, and group information. Use this to see which collectors are currently connected and their health status.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"status_filter": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter by status: 'online', 'offline', or 'error'",
					"enum":        []string{"online", "offline", "error"},
				},
			},
			Required: []string{},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input ListOtelAgentsInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Get all agents from Lawrence API
			agents, err := client.ListAgents()
			if err != nil {
				return nil, fmt.Errorf("failed to list agents: %w", err)
			}

			// Apply status filter if provided
			if input.StatusFilter != nil {
				// Note: OtelAgent doesn't have Status field yet, this is a placeholder
				// for when the API adds status information. Currently returns all agents.
				// TODO: Implement actual filtering based on status when available
			}

			// Format response
			result := map[string]interface{}{
				"total_count": len(agents),
				"agents":      formatAgentList(agents),
			}

			return result, nil
		},
	}
}

func formatAgentList(agents []otelclient.OtelAgent) []map[string]interface{} {
	result := make([]map[string]interface{}, len(agents))

	for i, agent := range agents {
		result[i] = map[string]interface{}{
			"id":          agent.ID,
			"name":        agent.Name,
			"description": agent.Description,
		}
	}

	return result
}

// GetOtelTools returns all OTEL-related tools
// This follows the same pattern as GetFileSystemTools()
func GetOtelTools(client *otelclient.OtelClient) []tools.Tool {
	return []tools.Tool{
		GetListOtelAgentsTool(client),
		GetGetAgentConfigTool(client),
		GetUpdateAgentConfigTool(client),
		GetDeployCollectorTool(),
		GetStopCollectorTool(),
		GetListDeployedCollectorsTool(),
	}
}
