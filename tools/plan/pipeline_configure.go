package plan

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// PipelineConfigureInput represents input for configuring a collector pipeline
type PipelineConfigureInput struct {
	PlanID      string `json:"plan_id"`
	PipelineID  string `json:"pipeline_id"`
	CollectorID string `json:"collector_id"`
	ConfigYAML  string `json:"config_yaml"`
	Rules       string `json:"rules"`
}

// GetPipelineConfigureTool creates a tool for pipeline configuration
func GetPipelineConfigureTool() tools.Tool {
	return tools.Tool{
		Name:        "configure_pipeline_from_plan",
		Description: "Configures a collector pipeline by applying sampling, filtering, and other processing rules",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"plan_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the observability plan",
				},
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the pipeline",
				},
				"collector_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the collector to configure",
				},
				"config_yaml": map[string]interface{}{
					"type":        "string",
					"description": "Collector configuration YAML",
				},
				"rules": map[string]interface{}{
					"type":        "string",
					"description": "Processing rules as JSON (sampling, filtering, etc.)",
				},
			},
			Required: []string{"plan_id", "pipeline_id", "collector_id", "config_yaml"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input PipelineConfigureInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// This would delegate to PipelineAgent
			return map[string]interface{}{
				"success":     true,
				"plan_id":     input.PlanID,
				"pipeline_id": input.PipelineID,
				"status":      "pipeline_configured",
				"message":     fmt.Sprintf("Pipeline configuration started for collector %s", input.CollectorID),
			}, nil
		},
	}
}
