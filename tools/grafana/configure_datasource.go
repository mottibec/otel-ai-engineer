package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/grafanaclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// ConfigureDatasourceInput represents the input for configuring a data source
type ConfigureDatasourceInput struct {
	GrafanaURL      string            `json:"grafana_url"`
	Username        string            `json:"username"`
	Password        string            `json:"password"`
	DatasourceName string            `json:"datasource_name"`
	DatasourceType string            `json:"datasource_type"`
	URL             string            `json:"url"`
	JSONData        map[string]interface{} `json:"json_data"`
}

// GetConfigureDatasourceTool creates a tool for configuring Grafana data sources
func GetConfigureDatasourceTool() tools.Tool {
	return tools.Tool{
		Name:        "configure_grafana_datasource",
		Description: "Configures a data source in Grafana. Supports OTLP (traces, metrics, logs), Prometheus, Loki, Tempo, and other Grafana-compatible data sources.",
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
				"datasource_name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the data source",
				},
				"datasource_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of data source: 'otlp', 'prometheus', 'loki', 'tempo', 'influxdb', 'elasticsearch', etc.",
					"enum":        []string{"otlp", "prometheus", "loki", "tempo", "influxdb", "elasticsearch"},
				},
				"url": map[string]interface{}{
					"type":        "string",
					"description": "Data source URL (e.g., http://lawrence:4318 for OTLP HTTP)",
				},
				"json_data": map[string]interface{}{
					"type":        "object",
					"description": "Additional JSON data for the data source configuration",
				},
			},
			Required: []string{"grafana_url", "username", "password", "datasource_name", "datasource_type", "url"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input ConfigureDatasourceInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Create Grafana client
			client := grafanaclient.NewClientWithAuth(input.GrafanaURL, input.Username, input.Password)

			// Map datasource type to Grafana type
			grafanaType := mapDatasourceTypeToGrafana(input.DatasourceType)
			
			// Create datasource
			datasource := grafanaclient.Datasource{
				Name: input.DatasourceName,
				Type: grafanaType,
				URL:  input.URL,
				JSONData: input.JSONData,
			}

			result, err := client.CreateDatasource(datasource)
			if err != nil {
				return nil, fmt.Errorf("failed to create datasource: %w", err)
			}

			return map[string]interface{}{
				"success":         true,
				"datasource_id":   result.ID,
				"datasource_uid": result.UID,
				"datasource_name": result.Name,
				"datasource_type": result.Type,
				"url":             result.URL,
			}, nil
		},
	}
}

// mapDatasourceTypeToGrafana maps our datasource type names to Grafana's type names
func mapDatasourceTypeToGrafana(dsType string) string {
	switch dsType {
	case "otlp":
		return "grafana-otlp-datasource"
	case "prometheus":
		return "prometheus"
	case "loki":
		return "loki"
	case "tempo":
		return "tempo"
	case "influxdb":
		return "influxdb"
	case "elasticsearch":
		return "elasticsearch"
	default:
		return dsType
	}
}
