package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	"github.com/mottibechhofer/otel-ai-engineer/tools/grafana/deployers"
)

// ListGrafanaInstancesInput represents the input for listing Grafana instances
type ListGrafanaInstancesInput struct {
	TargetType string `json:"target_type"`
}

// GetListGrafanaInstancesTool creates a tool for listing Grafana instances
func GetListGrafanaInstancesTool() tools.Tool {
	return tools.Tool{
		Name:        "list_grafana_instances",
		Description: "Lists all deployed Grafana instances for a specific target type (docker, kubernetes, or remote). Returns instance IDs, names, status, and URLs.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"target_type": map[string]interface{}{
					"type":        "string",
					"description": "Deployment target type: 'docker', 'kubernetes', or 'remote'",
					"enum":        []string{"docker", "kubernetes", "remote"},
				},
			},
			Required: []string{"target_type"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input ListGrafanaInstancesInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Get the appropriate deployer
			deployer, err := getDeployer(deployers.TargetType(input.TargetType))
			if err != nil {
				return nil, fmt.Errorf("unsupported target type %s: %w", input.TargetType, err)
			}

			// List instances
			instances, err := deployer.List()
			if err != nil {
				return nil, fmt.Errorf("failed to list instances: %w", err)
			}

			// Format response
			result := make([]map[string]interface{}, len(instances))
			for i, instance := range instances {
				result[i] = map[string]interface{}{
					"instance_id":   instance.InstanceID,
					"instance_name": instance.InstanceName,
					"target_type":   instance.TargetType,
					"status":        instance.Status,
					"url":           instance.URL,
					"deployed_at":   instance.DeployedAt,
				}
			}

			return map[string]interface{}{
				"total_count": len(instances),
				"instances":   result,
			}, nil
		},
	}
}
