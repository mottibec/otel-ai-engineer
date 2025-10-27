package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/grafanaclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// ListDatasourcesInput represents the input for listing data sources
type ListDatasourcesInput struct {
	GrafanaURL string `json:"grafana_url"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// GetListDatasourcesTool creates a tool for listing Grafana data sources
func GetListDatasourcesTool() tools.Tool {
	return tools.Tool{
		Name:        "list_grafana_datasources",
		Description: "Lists all configured data sources in a Grafana instance. Useful for checking existing data sources before adding new ones.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"grafana_url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the Grafana instance (e.g., http://localhost:3000)",
				},
				"username": map[string]interface{}{
					"type":        "string",
					"description": "Grafana admin username",
				},
				"password": map[string]interface{}{
					"type":        "string",
					"description": "Grafana admin password",
				},
			},
			Required: []string{"grafana_url", "username", "password"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input ListDatasourcesInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Create Grafana client
			client := grafanaclient.NewClientWithAuth(input.GrafanaURL, input.Username, input.Password)

			datasources, err := client.ListDatasources()
			if err != nil {
				return nil, fmt.Errorf("failed to list datasources: %w", err)
			}

			// Format response
			result := make([]map[string]interface{}, len(datasources))
			for i, ds := range datasources {
				result[i] = map[string]interface{}{
					"id":   ds.ID,
					"uid":  ds.UID,
					"name": ds.Name,
					"type": ds.Type,
					"url":  ds.URL,
				}
			}

			return map[string]interface{}{
				"total_count": len(datasources),
				"datasources": result,
			}, nil
		},
	}
}
