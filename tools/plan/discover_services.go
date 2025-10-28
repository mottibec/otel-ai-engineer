package plan

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// DiscoverServicesInput represents input for discovering services in a plan
type DiscoverServicesInput struct {
	PlanID     string `json:"plan_id"`
	CodePath   string `json:"code_path"`
	Environment string `json:"environment,omitempty"`
}

// Storage interface for plan operations - this will be injected
var planStorage storage.Storage

// SetPlanStorage sets the storage for plan tools
func SetPlanStorage(stor storage.Storage) {
	planStorage = stor
}

// GetDiscoverServicesTool creates a tool for discovering services in a codebase
func GetDiscoverServicesTool() tools.Tool {
	return tools.Tool{
		Name:        "discover_services_for_plan",
		Description: "Analyzes a codebase to discover services and adds them to an observability plan. Automatically detects languages, frameworks, and service names.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"plan_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the observability plan to add services to",
				},
				"code_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the codebase directory to analyze",
				},
				"environment": map[string]interface{}{
					"type":        "string",
					"description": "Target environment (production, staging, development)",
				},
			},
			Required: []string{"plan_id", "code_path"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input DiscoverServicesInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			if planStorage == nil {
				return nil, fmt.Errorf("plan storage not configured")
			}

			// This should guide the agent to:
			// 1. Use file system tools to explore the codebase
			// 2. Detect services (entry points, main files, etc.)
			// 3. Detect language and framework
			// 4. Create service entries via API

			return map[string]interface{}{
				"success":      true,
				"plan_id":      input.PlanID,
				"code_path":    input.CodePath,
				"message":      "Use file system tools to analyze the codebase, then create service components via API calls",
				"suggestion":   "1. Read main.go or index.js to detect language 2. Look for package.json, go.mod, requirements.txt 3. Detect framework (Gin, Express, Django, etc.) 4. Call POST /api/plans/{plan_id}/services to add each discovered service",
			}, nil
		},
	}
}
