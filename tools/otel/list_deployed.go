package otel

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	"github.com/mottibechhofer/otel-ai-engineer/tools/otel/deployers"
)

// ListDeployedInput represents the input for listing deployed collectors
type ListDeployedInput struct {
	TargetType string `json:"target_type,omitempty"`
}

// GetListDeployedCollectorsTool creates a tool for listing deployed collectors
func GetListDeployedCollectorsTool() tools.Tool {
	return tools.Tool{
		Name:        "list_deployed_collectors",
		Description: "Lists all OpenTelemetry collector instances deployed to various targets (docker, remote, kubernetes, local). Use this to see which collectors are currently running and their status.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"target_type": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter by target type: 'docker', 'remote', 'kubernetes', or 'local'",
					"enum":        []string{"docker", "remote", "kubernetes", "local"},
				},
			},
			Required: []string{},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input ListDeployedInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			var allCollectors []deployers.CollectorInfo

			// Define target types to check
			targetTypes := []deployers.TargetType{
				deployers.TargetDocker,
				// deployers.TargetRemote,
				// deployers.TargetK8s,
				// deployers.TargetLocal,
			}

			// Filter by target type if specified
			if input.TargetType != "" {
				targetTypes = []deployers.TargetType{deployers.TargetType(input.TargetType)}
			}

			// Get collectors from each target
			for _, targetType := range targetTypes {
				deployer, err := getDeployer(targetType)
				if err != nil {
					// Skip unsupported target types
					continue
				}

				collectors, err := deployer.List()
				if err != nil {
					// Log error but continue with other targets
					continue
				}

				allCollectors = append(allCollectors, collectors...)
			}

			result := make([]map[string]interface{}, len(allCollectors))
			for i, collector := range allCollectors {
				result[i] = map[string]interface{}{
					"collector_id":   collector.CollectorID,
					"collector_name": collector.CollectorName,
					"target_type":    collector.TargetType,
					"status":         collector.Status,
					"deployed_at":    collector.DeployedAt.Format("2006-01-02T15:04:05Z"),
					"config_path":    collector.ConfigPath,
				}
			}

			return map[string]interface{}{
				"total_count": len(allCollectors),
				"collectors":  result,
			}, nil
		},
	}
}

