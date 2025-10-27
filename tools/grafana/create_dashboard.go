package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/grafanaclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// CreateDashboardInput represents the input for creating a dashboard
type CreateDashboardInput struct {
	GrafanaURL    string                 `json:"grafana_url"`
	Username      string                 `json:"username"`
	Password      string                 `json:"password"`
	DashboardJSON map[string]interface{} `json:"dashboard_json"`
	FolderName    string                 `json:"folder_name"`
	Overwrite     bool                   `json:"overwrite"`
}

// GetCreateDashboardTool creates a tool for creating Grafana dashboards
func GetCreateDashboardTool() tools.Tool {
	return tools.Tool{
		Name:        "create_grafana_dashboard",
		Description: "Creates a new dashboard in Grafana from a JSON configuration. Supports both pre-built templates and custom dashboard JSON. The dashboard must follow Grafana's dashboard schema.",
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
				"dashboard_json": map[string]interface{}{
					"type":        "object",
					"description": "Dashboard JSON configuration following Grafana schema",
				},
				"folder_name": map[string]interface{}{
					"type":        "string",
					"description": "Folder name to organize the dashboard",
				},
				"overwrite": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to overwrite existing dashboard with same UID",
				},
			},
			Required: []string{"grafana_url", "username", "password", "dashboard_json"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input CreateDashboardInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Create Grafana client
			client := grafanaclient.NewClientWithAuth(input.GrafanaURL, input.Username, input.Password)

			// Create dashboard from JSON
			dashboard := grafanaclient.Dashboard{
				Dashboard: input.DashboardJSON,
			}

			result, err := client.CreateDashboard(dashboard)
			if err != nil {
				return nil, fmt.Errorf("failed to create dashboard: %w", err)
			}

			return map[string]interface{}{
				"success":   true,
				"dashboard": result,
				"message":   "Dashboard created successfully",
			}, nil
		},
	}
}
