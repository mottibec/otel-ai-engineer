package grafana

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// AnalyzeCodeForAlertsInput represents the input for AI-powered alert generation
type AnalyzeCodeForAlertsInput struct {
	GrafanaURL    string `json:"grafana_url"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	CodebasePath  string `json:"codebase_path"`
	DatasourceUID string `json:"datasource_uid"`
}

// GetAnalyzeCodeForAlertsTool creates a tool for analyzing code and generating alerts
func GetAnalyzeCodeForAlertsTool() tools.Tool {
	return tools.Tool{
		Name:        "analyze_code_and_generate_alerts",
		Description: "Analyzes application code to identify critical paths and business logic, then suggests context-aware alert rules. Detects authentication flows, payment endpoints, data processing pipelines, and suggests appropriate alerts.",
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
				"codebase_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the codebase directory",
				},
				"datasource_uid": map[string]interface{}{
					"type":        "string",
					"description": "UID of the data source to use for alerts",
				},
			},
			Required: []string{"grafana_url", "username", "password", "codebase_path", "datasource_uid"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input AnalyzeCodeForAlertsInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Analyze codebase to identify critical paths
			criticalPaths, err := analyzeCriticalPaths(input.CodebasePath)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze codebase: %w", err)
			}

			// Return detected critical paths and alert suggestions
			return map[string]interface{}{
				"success":         true,
				"critical_paths":  criticalPaths,
				"message":         "Code analyzed. Use create_grafana_alert_rule to create alerts based on critical paths. Ready to implement full AI-powered alert generation.",
				"alert_suggestions": generateAlertSuggestions(criticalPaths),
			}, nil
		},
	}
}

// analyzeCriticalPaths analyzes the codebase to identify critical operations
func analyzeCriticalPaths(path string) ([]string, error) {
	var criticalPaths []string

	// Read directory structure
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		lowerPath := strings.ToLower(currentPath)

		// Detect critical operations by file/directory names
		if strings.Contains(lowerPath, "auth") || strings.Contains(lowerPath, "login") {
			criticalPaths = append(criticalPaths, "Authentication/authorization flows")
		}
		if strings.Contains(lowerPath, "payment") || strings.Contains(lowerPath, "billing") {
			criticalPaths = append(criticalPaths, "Payment processing")
		}
		if strings.Contains(lowerPath, "database") || strings.Contains(lowerPath, "db") {
			criticalPaths = append(criticalPaths, "Database operations")
		}
		if strings.Contains(lowerPath, "api") || strings.Contains(lowerPath, "endpoint") {
			criticalPaths = append(criticalPaths, "API endpoints")
		}
		if strings.Contains(lowerPath, "worker") || strings.Contains(lowerPath, "job") {
			criticalPaths = append(criticalPaths, "Background jobs/workers")
		}

		return nil
	})

	return criticalPaths, err
}

// generateAlertSuggestions generates alert suggestions based on critical paths
func generateAlertSuggestions(criticalPaths []string) []map[string]interface{} {
	suggestions := []map[string]interface{}{
		{
			"name":       "High Error Rate",
			"threshold":  "error_rate > 0.05",
			"description": "Alert when error rate exceeds 5%",
		},
		{
			"name":       "High Latency",
			"threshold":  "p99_latency > 1s",
			"description": "Alert when 99th percentile latency exceeds 1 second",
		},
	}

	// Add context-aware alerts
	for _, path := range criticalPaths {
		if strings.Contains(path, "Authentication") {
			suggestions = append(suggestions, map[string]interface{}{
				"name":       "Auth Failure Rate",
				"threshold":  "auth_failures > 0.01",
				"description": "Alert on authentication failures >1%",
			})
		}
		if strings.Contains(path, "Payment") {
			suggestions = append(suggestions, map[string]interface{}{
				"name":       "Payment Failure Rate",
				"threshold":  "payment_failures > 0.01",
				"description": "Alert on payment failures >1%",
			})
		}
		if strings.Contains(path, "Database") {
			suggestions = append(suggestions, map[string]interface{}{
				"name":       "Database Connection Pool",
				"threshold":  "db_connections > 0.8",
				"description": "Alert when connection pool >80%",
			})
		}
	}

	return suggestions
}
