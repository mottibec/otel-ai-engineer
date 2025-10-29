package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	toolService "github.com/mottibechhofer/otel-ai-engineer/server/service/tools"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// AgentService handles business logic for agent management
type AgentService struct {
	storage       storage.Storage
	agentRegistry *agent.Registry
	toolDiscovery *toolService.ToolDiscoveryService
}

// NewAgentService creates a new agent service
func NewAgentService(
	stor storage.Storage,
	agentRegistry *agent.Registry,
	toolDiscovery *toolService.ToolDiscoveryService,
) *AgentService {
	service := &AgentService{
		storage:       stor,
		agentRegistry: agentRegistry,
		toolDiscovery: toolDiscovery,
	}

	// Load and register all custom agents from storage
	service.LoadCustomAgents(context.Background())

	return service
}

// LoadCustomAgents loads all custom agents from storage and registers them
func (as *AgentService) LoadCustomAgents(ctx context.Context) {
	customAgents, err := as.storage.ListCustomAgents()
	if err != nil {
		fmt.Printf("Warning: Failed to load custom agents from storage: %v\n", err)
		return
	}

	for _, customAgent := range customAgents {
		if err := as.registerCustomAgentInRegistry(customAgent); err != nil {
			fmt.Printf("Warning: Failed to register custom agent %s: %v\n", customAgent.ID, err)
		} else {
			fmt.Printf("Loaded custom agent: %s (%s)\n", customAgent.ID, customAgent.Name)
		}
	}
}

// ListAgents returns all agents (built-in and custom) with tool information
func (as *AgentService) ListAgents(ctx context.Context) ([]AgentResponse, error) {
	responses := []AgentResponse{}

	// Get built-in agents
	builtInAgents := as.agentRegistry.List()
	for _, agentInfo := range builtInAgents {
		response := AgentResponse{
			ID:          agentInfo.ID,
			Name:        agentInfo.Name,
			Description: agentInfo.Description,
			Model:       agentInfo.Model,
			Type:        "built-in",
		}

		// Get tools for built-in agent
		if agentInstance, err := as.agentRegistry.Create(agentInfo.ID, nil, config.LogLevelInfo, nil); err == nil {
			toolNames := agentInstance.ListTools()
			response.ToolNames = toolNames

			// Get tool details
			tools := []toolService.ToolInfo{}
			for _, toolName := range toolNames {
				if tool, err := as.toolDiscovery.GetTool(ctx, toolName); err == nil {
					tools = append(tools, *tool)
				}
			}
			response.Tools = tools
		}

		responses = append(responses, response)
	}

	// Get custom agents
	customAgents, err := as.storage.ListCustomAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to list custom agents: %w", err)
	}

	for _, customAgent := range customAgents {
		response := AgentResponse{
			ID:           customAgent.ID,
			Name:         customAgent.Name,
			Description:  customAgent.Description,
			Model:        customAgent.Model,
			SystemPrompt: customAgent.SystemPrompt,
			MaxTokens:    customAgent.MaxTokens,
			ToolNames:    customAgent.ToolNames,
			Type:         "custom",
			CreatedAt:    customAgent.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    customAgent.UpdatedAt.Format(time.RFC3339),
		}

		// Get tool details
		tools := []toolService.ToolInfo{}
		for _, toolName := range customAgent.ToolNames {
			if tool, err := as.toolDiscovery.GetTool(ctx, toolName); err == nil {
				tools = append(tools, *tool)
			}
		}
		response.Tools = tools

		responses = append(responses, response)
	}

	return responses, nil
}

// GetAgent retrieves an agent by ID with tool information
func (as *AgentService) GetAgent(ctx context.Context, agentID string) (*AgentResponse, error) {
	// Try built-in first
	if agentInfo, exists := as.agentRegistry.Get(agentID); exists {
		response := &AgentResponse{
			ID:          agentInfo.ID,
			Name:        agentInfo.Name,
			Description: agentInfo.Description,
			Model:       agentInfo.Model,
			Type:        "built-in",
		}

		// Get tools
		if agentInstance, err := as.agentRegistry.Create(agentInfo.ID, nil, config.LogLevelInfo, nil); err == nil {
			toolNames := agentInstance.ListTools()
			response.ToolNames = toolNames

			// Get tool details
			tools := []toolService.ToolInfo{}
			for _, toolName := range toolNames {
				if tool, err := as.toolDiscovery.GetTool(ctx, toolName); err == nil {
					tools = append(tools, *tool)
				}
			}
			response.Tools = tools
		}

		return response, nil
	}

	// Try custom agent
	customAgent, err := as.storage.GetCustomAgent(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	response := &AgentResponse{
		ID:           customAgent.ID,
		Name:         customAgent.Name,
		Description:  customAgent.Description,
		Model:        customAgent.Model,
		SystemPrompt: customAgent.SystemPrompt,
		MaxTokens:    customAgent.MaxTokens,
		ToolNames:    customAgent.ToolNames,
		Type:         "custom",
		CreatedAt:    customAgent.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    customAgent.UpdatedAt.Format(time.RFC3339),
	}

	// Get tool details
	tools := []toolService.ToolInfo{}
	for _, toolName := range customAgent.ToolNames {
		if tool, err := as.toolDiscovery.GetTool(ctx, toolName); err == nil {
			tools = append(tools, *tool)
		}
	}
	response.Tools = tools

	return response, nil
}

// GetAgentTools retrieves tools for a specific agent
func (as *AgentService) GetAgentTools(ctx context.Context, agentID string) ([]toolService.ToolInfo, error) {
	agent, err := as.GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	return agent.Tools, nil
}

// CreateCustomAgent creates a new custom agent
func (as *AgentService) CreateCustomAgent(ctx context.Context, req CreateCustomAgentRequest) (*AgentResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if len(req.ToolNames) == 0 {
		return nil, fmt.Errorf("at least one tool is required")
	}

	// Validate tool names exist
	if err := as.toolDiscovery.ValidateToolNames(ctx, req.ToolNames); err != nil {
		return nil, fmt.Errorf("invalid tool names: %w", err)
	}

	// Check if built-in agent with same ID exists (using name as ID check)
	if _, exists := as.agentRegistry.Get(req.Name); exists {
		return nil, fmt.Errorf("an agent with name '%s' already exists as built-in agent", req.Name)
	}

	// Generate ID
	agentID := fmt.Sprintf("custom-%d", time.Now().UnixNano())

	// Set defaults
	if req.Model == "" {
		req.Model = "claude-sonnet-4-5-20250929"
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	now := time.Now()
	customAgent := &storage.CustomAgent{
		ID:           agentID,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		MaxTokens:    req.MaxTokens,
		ToolNames:    req.ToolNames,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := as.storage.CreateCustomAgent(customAgent); err != nil {
		return nil, fmt.Errorf("failed to create custom agent: %w", err)
	}

	// Register in agent registry
	if err := as.registerCustomAgentInRegistry(customAgent); err != nil {
		// Log error but don't fail - agent is stored
		fmt.Printf("Warning: Failed to register custom agent in registry: %v\n", err)
	} else {
		fmt.Printf("Successfully registered custom agent %s in registry\n", customAgent.ID)
	}

	// Build response
	response := &AgentResponse{
		ID:           customAgent.ID,
		Name:         customAgent.Name,
		Description:  customAgent.Description,
		Model:        customAgent.Model,
		SystemPrompt: customAgent.SystemPrompt,
		MaxTokens:    customAgent.MaxTokens,
		ToolNames:    customAgent.ToolNames,
		Type:         "custom",
		CreatedAt:    customAgent.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    customAgent.UpdatedAt.Format(time.RFC3339),
	}

	// Get tool details
	tools := []toolService.ToolInfo{}
	for _, toolName := range customAgent.ToolNames {
		if tool, err := as.toolDiscovery.GetTool(ctx, toolName); err == nil {
			tools = append(tools, *tool)
		}
	}
	response.Tools = tools

	return response, nil
}

// UpdateCustomAgent updates a custom agent
func (as *AgentService) UpdateCustomAgent(ctx context.Context, agentID string, req UpdateCustomAgentRequest) (*AgentResponse, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID cannot be empty")
	}

	// Verify agent exists
	if _, err := as.storage.GetCustomAgent(agentID); err != nil {
		return nil, fmt.Errorf("failed to get custom agent: %w", err)
	}

	// Validate tool names if provided
	if req.ToolNames != nil {
		if err := as.toolDiscovery.ValidateToolNames(ctx, *req.ToolNames); err != nil {
			return nil, fmt.Errorf("invalid tool names: %w", err)
		}
	}

	// Build update
	update := &storage.CustomAgentUpdate{
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		MaxTokens:    req.MaxTokens,
		ToolNames:    req.ToolNames,
	}

	if err := as.storage.UpdateCustomAgent(agentID, update); err != nil {
		return nil, fmt.Errorf("failed to update custom agent: %w", err)
	}

	// Get updated agent
	updated, err := as.storage.GetCustomAgent(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated agent: %w", err)
	}

	// Re-register in registry if tool names changed
	if req.ToolNames != nil {
		if err := as.registerCustomAgentInRegistry(updated); err != nil {
			fmt.Printf("Warning: Failed to re-register custom agent in registry: %v\n", err)
		}
	}

	// Build response
	response := &AgentResponse{
		ID:           updated.ID,
		Name:         updated.Name,
		Description:  updated.Description,
		Model:        updated.Model,
		SystemPrompt: updated.SystemPrompt,
		MaxTokens:    updated.MaxTokens,
		ToolNames:    updated.ToolNames,
		Type:         "custom",
		CreatedAt:    updated.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    updated.UpdatedAt.Format(time.RFC3339),
	}

	// Get tool details
	tools := []toolService.ToolInfo{}
	for _, toolName := range updated.ToolNames {
		if tool, err := as.toolDiscovery.GetTool(ctx, toolName); err == nil {
			tools = append(tools, *tool)
		}
	}
	response.Tools = tools

	return response, nil
}

// DeleteCustomAgent deletes a custom agent
func (as *AgentService) DeleteCustomAgent(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Check if it exists
	if _, err := as.storage.GetCustomAgent(agentID); err != nil {
		return fmt.Errorf("custom agent not found")
	}

	if err := as.storage.DeleteCustomAgent(agentID); err != nil {
		return fmt.Errorf("failed to delete custom agent: %w", err)
	}

	return nil
}

// CreateMetaAgent creates a meta-agent (placeholder for now)
func (as *AgentService) CreateMetaAgent(ctx context.Context, req CreateMetaAgentRequest) (*AgentResponse, error) {
	// For now, treat meta-agent as a special type of custom agent
	// TODO: Implement actual meta-agent logic
	return as.CreateCustomAgent(ctx, CreateCustomAgentRequest{
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		ToolNames:    req.AvailableToolNames,
	})
}

// registerCustomAgentInRegistry registers a custom agent in the agent registry
func (as *AgentService) registerCustomAgentInRegistry(customAgent *storage.CustomAgent) error {
	// Convert model string
	var model string
	if customAgent.Model != "" {
		model = customAgent.Model
	} else {
		model = "claude-sonnet-4-5-20250929"
	}

	// Register custom agent in registry
	as.agentRegistry.RegisterCustomAgent(
		agent.AgentInfo{
			ID:          customAgent.ID,
			Name:        customAgent.Name,
			Description: customAgent.Description,
			Model:       model,
		},
		customAgent.ToolNames,
		customAgent.SystemPrompt,
		customAgent.MaxTokens,
		func(toolNames []string) []tools.Tool {
			// Get actual Tool objects with handlers from tool discovery service
			return as.toolDiscovery.GetToolsByName(toolNames)
		},
	)

	return nil
}
