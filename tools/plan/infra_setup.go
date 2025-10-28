package plan

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// InfrastructureSetupInput represents input for setting up infrastructure monitoring
type InfrastructureSetupInput struct {
	PlanID        string `json:"plan_id"`
	InfraID       string `json:"infra_id"`
	ComponentName string `json:"component_name"`
	ComponentType string `json:"component_type"`
	ReceiverType  string `json:"receiver_type"`
	Host          string `json:"host"`
}

// GetInfrastructureSetupTool creates a tool for infrastructure setup
func GetInfrastructureSetupTool() tools.Tool {
	return tools.Tool{
		Name:        "setup_infrastructure_from_plan",
		Description: "Sets up infrastructure monitoring by configuring and deploying collectors with appropriate receivers",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"plan_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the observability plan",
				},
				"infra_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the infrastructure component",
				},
				"component_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the infrastructure component",
				},
				"component_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of infrastructure (database, cache, queue, host)",
				},
				"receiver_type": map[string]interface{}{
					"type":        "string",
					"description": "OpenTelemetry receiver type (postgresql, mysql, redis, hostmetrics, etc.)",
				},
				"host": map[string]interface{}{
					"type":        "string",
					"description": "Host address for the infrastructure component",
				},
			},
			Required: []string{"plan_id", "infra_id", "component_name", "component_type", "receiver_type", "host"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input InfrastructureSetupInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// This would delegate to InfrastructureAgent
			return map[string]interface{}{
				"success":  true,
				"plan_id":  input.PlanID,
				"infra_id": input.InfraID,
				"status":   "infrastructure_setup_started",
				"message":  fmt.Sprintf("Infrastructure monitoring setup for %s started", input.ComponentName),
			}, nil
		},
	}
}
