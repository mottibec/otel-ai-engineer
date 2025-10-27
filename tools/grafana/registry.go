package grafana

import "github.com/mottibechhofer/otel-ai-engineer/tools"

// GetGrafanaTools returns all Grafana-related tools
func GetGrafanaTools() []tools.Tool {
	return []tools.Tool{
		GetDeployGrafanaTool(),
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
