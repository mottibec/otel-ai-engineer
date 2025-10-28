package plan

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// ServiceInstrumentInput represents input for instrumenting a service
type ServiceInstrumentInput struct {
	PlanID      string `json:"plan_id"`
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	TargetPath  string `json:"target_path"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
}

// GetServiceInstrumentTool creates a tool for instrumenting services
func GetServiceInstrumentTool() tools.Tool {
	return tools.Tool{
		Name:        "instrument_service_from_plan",
		Description: "Instruments a service from an observability plan by analyzing the codebase, installing OpenTelemetry SDK, and configuring exporters",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"plan_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the observability plan",
				},
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the service to instrument",
				},
				"service_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the service",
				},
				"target_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the service codebase",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Programming language (go, python, java, javascript, etc.)",
				},
				"framework": map[string]interface{}{
					"type":        "string",
					"description": "Framework name (if applicable)",
				},
			},
			Required: []string{"plan_id", "service_id", "service_name", "target_path", "language"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input ServiceInstrumentInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// This tool should guide the ObservabilityAgent to use the handoff_task tool
			// The ObservabilityAgent will use handoff_task to delegate to InstrumentationAgent
			return map[string]interface{}{
				"success":    true,
				"plan_id":    input.PlanID,
				"service_id": input.ServiceID,
				"status":     "pending_delegation",
				"message":    fmt.Sprintf("Use the handoff_task tool to delegate instrumentation of service '%s' to the 'instrumentation' agent", input.ServiceName),
				"suggestion": "Call handoff_task with to_agent_id='instrumentation' and task_description describing the instrumentation task",
			}, nil
		},
	}
}
