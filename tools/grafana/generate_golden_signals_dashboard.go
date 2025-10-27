package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// GenerateGoldenSignalsInput represents the input for generating golden signals dashboard
type GenerateGoldenSignalsInput struct {
	GrafanaURL    string `json:"grafana_url"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ServiceName   string `json:"service_name"`
	DatasourceUID string `json:"datasource_uid"`
}

// GetGenerateGoldenSignalsDashboardTool creates a tool for generating golden signals dashboard
func GetGenerateGoldenSignalsDashboardTool() tools.Tool {
	return tools.Tool{
		Name:        "generate_golden_signals_dashboard",
		Description: "Generates a pre-configured dashboard with RED (Rate, Errors, Duration) and USE (Utilization, Saturation, Errors) metrics for a service. This is a standard observability dashboard template.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"grafana_url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the Grafana instance",
				},
				"username": map[string]interface{}{
					"type":        "string",
					"description": "Grafana admin username",
				},
				"password": map[string]interface{}{
					"type":        "string",
					"description": "Grafana admin password",
				},
				"service_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the service to monitor",
				},
				"datasource_uid": map[string]interface{}{
					"type":        "string",
					"description": "UID of the data source to use",
				},
			},
			Required: []string{"grafana_url", "username", "password", "service_name", "datasource_uid"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input GenerateGoldenSignalsInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Note: This would generate a standard dashboard JSON
			// For now, return a message instructing to use create_dashboard with templates
			return map[string]interface{}{
				"message": "Golden signals dashboard requires Grafana dashboard JSON. Use create_grafana_dashboard tool with a pre-built template. Ready to implement full dashboard generation.",
			}, nil
		},
	}
}
