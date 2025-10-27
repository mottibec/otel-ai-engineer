package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/grafanaclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// CreateAlertRuleInput represents the input for creating an alert rule
type CreateAlertRuleInput struct {
	GrafanaURL     string                 `json:"grafana_url"`
	Username       string                 `json:"username"`
	Password       string                 `json:"password"`
	RuleName       string                 `json:"rule_name"`
	Condition      string                 `json:"condition"`
	Data           []interface{}          `json:"data"`
	ExecErrState   string                 `json:"exec_err_state"`
	For            string                 `json:"for"`
	NoDataState    string                 `json:"no_data_state"`
	Annotations    map[string]interface{} `json:"annotations"`
	Labels         map[string]interface{} `json:"labels"`
}

// GetCreateAlertRuleTool creates a tool for creating Grafana alert rules
func GetCreateAlertRuleTool() tools.Tool {
	return tools.Tool{
		Name:        "create_grafana_alert_rule",
		Description: "Creates a new alert rule in Grafana. Define conditions, thresholds, and notification channels for monitoring metrics and triggering alerts.",
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
				"rule_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the alert rule",
				},
				"condition": map[string]interface{}{
					"type":        "string",
					"description": "Alert condition (e.g., A > 5)",
				},
				"data": map[string]interface{}{
					"type":        "array",
					"description": "Array of data queries for the alert",
				},
				"exec_err_state": map[string]interface{}{
					"type":        "string",
					"description": "State when execution error occurs: 'OK', 'Alerting', 'NoData'",
				},
				"for": map[string]interface{}{
					"type":        "string",
					"description": "Duration to wait before firing (e.g., '5m')",
				},
				"no_data_state": map[string]interface{}{
					"type":        "string",
					"description": "State when no data: 'OK', 'Alerting', 'NoData'",
				},
				"annotations": map[string]interface{}{
					"type":        "object",
					"description": "Annotations for the alert",
				},
				"labels": map[string]interface{}{
					"type":        "object",
					"description": "Labels for the alert",
				},
			},
			Required: []string{"grafana_url", "username", "password", "rule_name", "condition", "data"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input CreateAlertRuleInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Create Grafana client
			client := grafanaclient.NewClientWithAuth(input.GrafanaURL, input.Username, input.Password)

			// Create alert rule
			rule := grafanaclient.AlertRule{
				Title:        input.RuleName,
				Condition:    input.Condition,
				Data:         input.Data,
				ExecErrState: input.ExecErrState,
				For:          input.For,
				NoDataState:  input.NoDataState,
				Annotations:   input.Annotations,
				Labels:       input.Labels,
			}

			err := client.CreateAlertRule(rule)
			if err != nil {
				return nil, fmt.Errorf("failed to create alert rule: %w", err)
			}

			return map[string]interface{}{
				"success":   true,
				"rule_name": input.RuleName,
				"message":   "Alert rule created successfully",
			}, nil
		},
	}
}
