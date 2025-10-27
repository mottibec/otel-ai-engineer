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

// AnalyzeCodeForDashboardInput represents the input for AI-powered dashboard generation
type AnalyzeCodeForDashboardInput struct {
	GrafanaURL    string `json:"grafana_url"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	CodebasePath  string `json:"codebase_path"`
	DatasourceUID string `json:"datasource_uid"`
}

// GetAnalyzeCodeForDashboardTool creates a tool for analyzing code and generating dashboards
func GetAnalyzeCodeForDashboardTool() tools.Tool {
	return tools.Tool{
		Name:        "analyze_code_and_generate_dashboard",
		Description: "Analyzes application code to understand structure and detect frameworks, then generates custom Grafana dashboards based on discovered patterns. Detects HTTP servers, databases, message queues, and critical operations. Uses file system tools to read source code.",
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
					"description": "UID of the data source to use for metrics",
				},
			},
			Required: []string{"grafana_url", "username", "password", "codebase_path", "datasource_uid"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input AnalyzeCodeForDashboardInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Analyze codebase to detect frameworks and patterns
			detectedPatterns, err := analyzeCodebase(input.CodebasePath)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze codebase: %w", err)
			}

			// Return detected patterns and recommendation
			return map[string]interface{}{
				"success":           true,
				"detected_patterns": detectedPatterns,
				"message":           "Code analyzed. Use create_grafana_dashboard to generate custom dashboard based on detected patterns. Ready to implement full AI-powered dashboard generation.",
				"recommendations":   generateDashboardRecommendations(detectedPatterns),
			}, nil
		},
	}
}

// analyzeCodebase analyzes the codebase to detect frameworks and patterns
func analyzeCodebase(path string) ([]string, error) {
	var patterns []string

	// Read directory structure
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and common directories
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Detect frameworks by file extensions and patterns
		switch strings.ToLower(filepath.Ext(currentPath)) {
		case ".js", ".jsx", ".ts", ".tsx":
			patterns = append(patterns, "Node.js/JavaScript application detected")
		case ".go":
			patterns = append(patterns, "Go application detected")
		case ".py":
			patterns = append(patterns, "Python application detected")
		case ".java":
			patterns = append(patterns, "Java application detected")
		}

		// Detect framework-specific files
		if strings.Contains(strings.ToLower(currentPath), "package.json") {
			patterns = append(patterns, "npm/Node.js package detected")
		}
		if strings.Contains(strings.ToLower(currentPath), "pom.xml") || strings.Contains(strings.ToLower(currentPath), "build.gradle") {
			patterns = append(patterns, "Java Maven/Gradle project detected")
		}
		if strings.Contains(strings.ToLower(currentPath), "requirements.txt") || strings.Contains(strings.ToLower(currentPath), "setup.py") {
			patterns = append(patterns, "Python package detected")
		}

		return nil
	})

	return patterns, err
}

// generateDashboardRecommendations generates dashboard recommendations based on detected patterns
func generateDashboardRecommendations(patterns []string) []string {
	recommendations := []string{
		"HTTP endpoint latency and error rate monitoring",
		"Database connection pool and query performance",
		"Application memory and CPU usage",
	}

	// Add framework-specific recommendations
	for _, pattern := range patterns {
		if strings.Contains(pattern, "Node.js") {
			recommendations = append(recommendations, "Node.js event loop lag monitoring")
			recommendations = append(recommendations, "V8 heap size and GC metrics")
		}
		if strings.Contains(pattern, "Go") {
			recommendations = append(recommendations, "Go runtime metrics (goroutines, GC)")
		}
		if strings.Contains(pattern, "Python") {
			recommendations = append(recommendations, "Python GIL and thread pool metrics")
		}
	}

	return recommendations
}
