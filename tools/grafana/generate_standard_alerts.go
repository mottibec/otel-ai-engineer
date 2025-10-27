package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// GenerateStandardAlertsInput represents the input for generating standard alerts
type GenerateStandardAlertsInput struct {
	GrafanaURL    string `json:"grafana_url"`
	Username      string `json:"username"`
	Password       string `json:"password"`
	ServiceName   string `json:"service_name"`
	DatasourceUID string `json:"datasource_uid"`
}

// GetGenerateStandardAlertsTool creates a tool for generating standard alerts
func GetGenerateStandardAlertsTool() tools.Tool {
	return tools.Tool{
		Name:        "generate_standard_alerts",
		Description: "Generates standard alert rules for common issues: high error rate (>5%), high latency (p99 >1s), high CPU/memory utilization. These are pre-configured alerts based on best practices.",
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
			var input GenerateStandardAlertsInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Note: This would generate standard alert rules
			// For now, return a message
			return map[string]interface{}{
				"message": "Standard alert generation requires full Grafana alert configuration. Use create_grafana_alert_rule tool with pre-built alert templates. Ready to implement full alert generation.",
				"standard_alerts": []string{
					"High Error Rate (>5%)",
					"High Latency (p99 >1s)",
					"High CPU Usage",
					"High Memory Usage",
				},
			}, nil
		},
	}
}
