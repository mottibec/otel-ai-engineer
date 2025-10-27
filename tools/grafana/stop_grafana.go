package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	"github.com/mottibechhofer/otel-ai-engineer/tools/grafana/deployers"
)

// StopGrafanaInput represents the input for stopping Grafana
type StopGrafanaInput struct {
	InstanceID  string                 `json:"instance_id"`
	TargetType  string                 `json:"target_type"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GetStopGrafanaTool creates a tool for stopping Grafana instances
func GetStopGrafanaTool() tools.Tool {
	return tools.Tool{
		Name:        "stop_grafana",
		Description: "Stops and removes a running Grafana instance. This will shut down the Grafana container/deployment and clean up resources.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"instance_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the Grafana instance to stop",
				},
				"target_type": map[string]interface{}{
					"type":        "string",
					"description": "Deployment target type: 'docker', 'kubernetes', or 'remote'",
					"enum":        []string{"docker", "kubernetes", "remote"},
				},
				"parameters": map[string]interface{}{
					"type":        "object",
					"description": "Target-specific parameters for stopping",
				},
			},
			Required: []string{"instance_id", "target_type"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input StopGrafanaInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Get the appropriate deployer
			deployer, err := getDeployer(deployers.TargetType(input.TargetType))
			if err != nil {
				return nil, fmt.Errorf("unsupported target type %s: %w", input.TargetType, err)
			}

			// Stop the instance
			err = deployer.Stop(input.InstanceID, input.Parameters)
			if err != nil {
				return nil, fmt.Errorf("failed to stop instance: %w", err)
			}

			return map[string]interface{}{
				"success":     true,
				"instance_id": input.InstanceID,
				"message":     fmt.Sprintf("Successfully stopped Grafana instance %s", input.InstanceID),
			}, nil
		},
	}
}
