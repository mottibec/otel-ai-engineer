package grafana

import (
	dc "github.com/mottibechhofer/otel-ai-engineer/tools/dockerclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// GetGrafanaTools returns all Grafana-related tools
func GetGrafanaTools(dockerClient *dc.Client) []tools.Tool {
	// Set the shared Docker client
	SetDockerClient(dockerClient)
	
	return []tools.Tool{
		GetDeployGrafanaTool(dockerClient),
		GetStopGrafanaTool(),
		GetListGrafanaInstancesTool(),
		GetConfigureDatasourceTool(),
		GetListDatasourcesTool(),
		GetAutoDiscoverSourcesTool(),
		GetCreateDashboardTool(),
		GetGenerateGoldenSignalsDashboardTool(),
		GetAnalyzeCodeForDashboardTool(),
		GetCreateAlertRuleTool(),
		GetGenerateStandardAlertsTool(),
		GetAnalyzeCodeForAlertsTool(),
	}
}
