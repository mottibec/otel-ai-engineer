package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// AutoDiscoverSourcesInput represents the input for auto-discovery
type AutoDiscoverSourcesInput struct {
	GrafanaURL   string `json:"grafana_url"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	LawrenceURL  string `json:"lawrence_url"`
}

// GetAutoDiscoverSourcesTool creates a tool for auto-discovering and configuring data sources
func GetAutoDiscoverSourcesTool() tools.Tool {
	return tools.Tool{
		Name:        "auto_discover_datasources",
		Description: "Automatically discovers OpenTelemetry collectors and other observability backends, then configures appropriate data sources in Grafana. This tool queries the Lawrence API for collectors and creates datasources pointing to OTLP endpoints.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"grafana_url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the Grafana instance (e.g., http://localhost:3000)",
				},
				"username": map[string]interface{}{
					"type":        "string",
					"description": "Grafana admin username",
				},
				"password": map[string]interface{}{
					"type":        "string",
					"description": "Grafana admin password",
				},
				"lawrence_url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the Lawrence OTLP server (e.g., http://lawrence:4318)",
				},
			},
			Required: []string{"grafana_url", "username", "password"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input AutoDiscoverSourcesInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Create Lawrence client to discover collectors
			lawrenceURL := input.LawrenceURL
			if lawrenceURL == "" {
				lawrenceURL = "http://lawrence:8080" // Default
			}

			otelClient := otelclient.NewOtelClient(lawrenceURL, nil)
			
			// Get all agents
			agents, err := otelClient.ListAgents()
			if err != nil {
				return nil, fmt.Errorf("failed to list OTEL agents: %w", err)
			}

			// Discovered datasources
			discovered := []map[string]interface{}{
				{
					"name": "Lawrence OTLP",
					"type": "otlp",
					"url":  "http://lawrence:4318",
					"description": "OpenTelemetry traces, metrics, and logs from Lawrence",
				},
			}

			// Add a datasource for each agent (if needed)
			for _, agent := range agents {
				discovered = append(discovered, map[string]interface{}{
					"name": fmt.Sprintf("OTLP - %s", agent.Name),
					"type": "otlp",
					"url":  fmt.Sprintf("http://%s:4318", agent.Name),
					"description": fmt.Sprintf("OTLP endpoint for collector %s", agent.Name),
				})
			}

			return map[string]interface{}{
				"success": true,
				"discovered_count": len(discovered),
				"datasources": discovered,
				"message": "Found datasources. Use configure_grafana_datasource to add them to Grafana.",
			}, nil
		},
	}
}
