package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	dc "github.com/mottibechhofer/otel-ai-engineer/tools/dockerclient"
	grafanaTools "github.com/mottibechhofer/otel-ai-engineer/tools/grafana"
	otelTools "github.com/mottibechhofer/otel-ai-engineer/tools/otel"
	planTools "github.com/mottibechhofer/otel-ai-engineer/tools/plan"
	sandboxTools "github.com/mottibechhofer/otel-ai-engineer/tools/sandbox"
)

// ToolDiscoveryService collects and provides access to all available tools
type ToolDiscoveryService struct {
	tools        map[string]ToolInfo   // Tool metadata
	toolObjects  map[string]tools.Tool // Actual tool objects with handlers
	categories   map[string][]string   // category -> tool names
	mu           sync.RWMutex
	dockerClient *dc.Client
	otelClient   *otelclient.OtelClient
}

// ToolInfo represents information about a tool for API responses
type ToolInfo struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Schema      anthropic.ToolInputSchemaParam `json:"schema"`
	Category    string                         `json:"category"`
}

// ListToolsResponse represents the response for listing all tools
type ListToolsResponse struct {
	Tools      []ToolInfo            `json:"tools"`
	ByCategory map[string][]ToolInfo `json:"by_category"`
}

// NewToolDiscoveryService creates a new tool discovery service
func NewToolDiscoveryService(dockerClient *dc.Client, otelClient *otelclient.OtelClient) *ToolDiscoveryService {
	service := &ToolDiscoveryService{
		tools:        make(map[string]ToolInfo),
		toolObjects:  make(map[string]tools.Tool),
		categories:   make(map[string][]string),
		dockerClient: dockerClient,
		otelClient:   otelClient,
	}

	// Initialize tools
	service.refreshTools()

	return service
}

// refreshTools collects all tools from all packages
func (s *ToolDiscoveryService) refreshTools() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing
	s.tools = make(map[string]ToolInfo)
	s.toolObjects = make(map[string]tools.Tool)
	s.categories = make(map[string][]string)

	// Collect filesystem tools
	s.collectToolsFromList(tools.GetFileSystemTools(), "filesystem")

	// Collect plan tools
	s.collectToolsFromList(planTools.GetPlanTools(), "plan")

	// Collect sandbox tools (may fail if not initialized, that's okay)
	sandboxToolsList := sandboxTools.GetSandboxTools()
	s.collectToolsFromList(sandboxToolsList, "sandbox")

	// Collect Grafana tools (requires docker client)
	if s.dockerClient != nil {
		grafanaToolsList := grafanaTools.GetGrafanaTools(s.dockerClient)
		s.collectToolsFromList(grafanaToolsList, "grafana")
	}

	// Collect OTEL tools (requires otel client)
	if s.otelClient != nil {
		otelToolsList := otelTools.GetOtelTools(s.otelClient)
		s.collectToolsFromList(otelToolsList, "otel")
	}
}

// collectToolsFromList adds tools from a list to the catalog
func (s *ToolDiscoveryService) collectToolsFromList(toolList []tools.Tool, category string) {
	for _, tool := range toolList {
		// Convert schema to JSON-serializable format
		schemaJSON, err := json.Marshal(tool.Schema)
		if err != nil {
			continue // Skip tools that can't be serialized
		}

		var schema anthropic.ToolInputSchemaParam
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			continue
		}

		info := ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
			Schema:      schema,
			Category:    category,
		}

		s.tools[tool.Name] = info
		s.toolObjects[tool.Name] = tool // Store actual tool object (copy) with handler
		s.categories[category] = append(s.categories[category], tool.Name)
	}
}

// GetAllTools returns all available tools
func (s *ToolDiscoveryService) GetAllTools(ctx context.Context) (*ListToolsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice
	toolList := make([]ToolInfo, 0, len(s.tools))
	for _, info := range s.tools {
		toolList = append(toolList, info)
	}

	// Build by category map
	byCategory := make(map[string][]ToolInfo)
	for category, toolNames := range s.categories {
		categoryTools := make([]ToolInfo, 0, len(toolNames))
		for _, toolName := range toolNames {
			if info, exists := s.tools[toolName]; exists {
				categoryTools = append(categoryTools, info)
			}
		}
		byCategory[category] = categoryTools
	}

	return &ListToolsResponse{
		Tools:      toolList,
		ByCategory: byCategory,
	}, nil
}

// GetTool returns a specific tool by name
func (s *ToolDiscoveryService) GetTool(ctx context.Context, name string) (*ToolInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tool, exists := s.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return &tool, nil
}

// GetToolsByCategory returns all tools in a specific category
func (s *ToolDiscoveryService) GetToolsByCategory(ctx context.Context, category string) ([]ToolInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	toolNames, exists := s.categories[category]
	if !exists {
		return []ToolInfo{}, nil
	}

	tools := make([]ToolInfo, 0, len(toolNames))
	for _, toolName := range toolNames {
		if info, exists := s.tools[toolName]; exists {
			tools = append(tools, info)
		}
	}

	return tools, nil
}

// HasTool checks if a tool exists
func (s *ToolDiscoveryService) HasTool(ctx context.Context, name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.tools[name]
	return exists
}

// ValidateToolNames validates that all tool names exist in the catalog
func (s *ToolDiscoveryService) ValidateToolNames(ctx context.Context, toolNames []string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, toolName := range toolNames {
		if _, exists := s.tools[toolName]; !exists {
			return fmt.Errorf("tool not found: %s", toolName)
		}
	}

	return nil
}

// Refresh forces a refresh of the tool catalog
func (s *ToolDiscoveryService) Refresh() {
	s.refreshTools()
}

// GetToolsByName returns actual Tool objects with handlers for the given tool names
func (s *ToolDiscoveryService) GetToolsByName(toolNames []string) []tools.Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]tools.Tool, 0, len(toolNames))
	for _, toolName := range toolNames {
		if toolObj, exists := s.toolObjects[toolName]; exists {
			// Return a copy of the tool
			result = append(result, toolObj)
		}
	}

	return result
}
