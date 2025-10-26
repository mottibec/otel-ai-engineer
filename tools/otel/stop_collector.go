package otel

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	"github.com/mottibechhofer/otel-ai-engineer/tools/otel/deployers"
)

// StopCollectorInput represents the input for stopping a collector
type StopCollectorInput struct {
	TargetType    string                 `json:"target_type"`
	CollectorID   string                 `json:"collector_id"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// GetStopCollectorTool creates a tool for stopping a collector
func GetStopCollectorTool() tools.Tool {
	return tools.Tool{
		Name:        "stop_otel_collector",
		Description: "Stops and removes a deployed OpenTelemetry collector instance from the specified target.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"target_type": map[string]interface{}{
					"type":        "string",
					"description": "Deployment target type where the collector is running",
					"enum":        []string{"docker", "remote", "kubernetes", "local"},
				},
				"collector_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the collector to stop",
				},
				"parameters": map[string]interface{}{
					"type":        "object",
					"description": "Target-specific parameters (currently unused for docker)",
				},
			},
			Required: []string{"target_type", "collector_id"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input StopCollectorInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Get the appropriate deployer
			deployer, err := getDeployer(deployers.TargetType(input.TargetType))
			if err != nil {
				return nil, fmt.Errorf("unsupported target type %s: %w", input.TargetType, err)
			}

			// Stop the collector
			err = deployer.Stop(input.CollectorID, input.Parameters)
			if err != nil {
				return nil, fmt.Errorf("failed to stop collector: %w", err)
			}

			return map[string]interface{}{
				"success":     true,
				"collector_id": input.CollectorID,
				"target_type": input.TargetType,
				"message":     fmt.Sprintf("Collector %s successfully stopped and removed", input.CollectorID),
			}, nil
		},
	}
}

