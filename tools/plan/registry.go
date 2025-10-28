package plan

import "github.com/mottibechhofer/otel-ai-engineer/tools"

// GetPlanTools returns all plan management tools
func GetPlanTools() []tools.Tool {
	return []tools.Tool{
		GetServiceInstrumentTool(),
		GetInfrastructureSetupTool(),
		GetPipelineConfigureTool(),
		GetBackendConnectTool(),
		GetDiscoverServicesTool(),
	}
}
