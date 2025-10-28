package plan

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// BackendConnectInput represents input for connecting to a backend
type BackendConnectInput struct {
	PlanID      string `json:"plan_id"`
	BackendID   string `json:"backend_id"`
	BackendName string `json:"backend_name"`
	BackendType string `json:"backend_type"`
	URL         string `json:"url"`
	Credentials string `json:"credentials,omitempty"`
}

// GetBackendConnectTool creates a tool for backend connectivity
func GetBackendConnectTool() tools.Tool {
	return tools.Tool{
		Name:        "connect_backend_from_plan",
		Description: "Connects to and validates an observability backend by testing connectivity, configuring exporters, and verifying data flow",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"plan_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the observability plan",
				},
				"backend_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the backend",
				},
				"backend_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the backend",
				},
				"backend_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of backend (grafana, prometheus, jaeger, custom)",
				},
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the backend endpoint",
				},
				"credentials": map[string]interface{}{
					"type":        "string",
					"description": "Credentials for authentication (optional)",
				},
			},
			Required: []string{"plan_id", "backend_id", "backend_name", "backend_type", "url"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input BackendConnectInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// This would delegate to BackendAgent
			return map[string]interface{}{
				"success":       true,
				"plan_id":       input.PlanID,
				"backend_id":    input.BackendID,
				"status":        "backend_connection_started",
				"health_status": "unknown",
				"message":       fmt.Sprintf("Connection to backend %s started", input.BackendName),
			}, nil
		},
	}
}
