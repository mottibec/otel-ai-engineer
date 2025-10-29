package collector

import (
	"context"
	"encoding/json"
	"fmt"

	dc "github.com/mottibechhofer/otel-ai-engineer/tools/dockerclient"
	otelTools "github.com/mottibechhofer/otel-ai-engineer/tools/otel"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/server/service"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// CollectorService handles business logic for collector management
type CollectorService struct {
	storage          storage.Storage
	agentWorkService *service.AgentWorkService
	otelClient       *otelclient.OtelClient
}

// NewCollectorService creates a new collector service
func NewCollectorService(stor storage.Storage, agentWorkService *service.AgentWorkService, otelClient *otelclient.OtelClient) *CollectorService {
	return &CollectorService{
		storage:          stor,
		agentWorkService: agentWorkService,
		otelClient:       otelClient,
	}
}

// ListCollectors lists deployed collectors with optional target type filter
func (cs *CollectorService) ListCollectors(ctx context.Context, targetType string, enrichWithAgentWork bool) (*ListCollectorsResponse, error) {
	// Use the list_deployed_collectors tool logic
	tool := otelTools.GetListDeployedCollectorsTool()

	var inputJSON json.RawMessage
	if targetType != "" {
		input := map[string]interface{}{
			"target_type": targetType,
		}
		var err error
		inputJSON, err = json.Marshal(input)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input: %w", err)
		}
	} else {
		inputJSON = json.RawMessage("{}")
	}

	result, err := tool.Handler(inputJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to list collectors: %w", err)
	}

	// Extract collectors from result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	// Handle collectors field - could be nil or missing
	var collectorsData []interface{}
	if collectorsField, exists := resultMap["collectors"]; exists && collectorsField != nil {
		switch v := collectorsField.(type) {
		case []interface{}:
			collectorsData = v
		case []map[string]interface{}:
			// Convert []map[string]interface{} to []interface{}
			collectorsData = make([]interface{}, len(v))
			for i, item := range v {
				collectorsData[i] = item
			}
		default:
			// If it's not a recognized type, treat as empty list
			collectorsData = []interface{}{}
		}
	} else {
		// No collectors field or it's nil - return empty list
		collectorsData = []interface{}{}
	}

	// Convert to response format and enrich with agent work
	collectors := make([]CollectorResponse, 0, len(collectorsData))
	for _, c := range collectorsData {
		collectorMap, ok := c.(map[string]interface{})
		if !ok {
			// Skip invalid entries
			continue
		}

		// Safely extract fields with type assertions
		collectorID, ok := collectorMap["collector_id"].(string)
		if !ok || collectorID == "" {
			continue
		}

		collectorName, _ := collectorMap["collector_name"].(string)
		targetTypeFromMap, _ := collectorMap["target_type"].(string)
		status, _ := collectorMap["status"].(string)

		var deployedAt string
		if deployedAtVal, ok := collectorMap["deployed_at"]; ok {
			if deployedAtStr, ok := deployedAtVal.(string); ok {
				deployedAt = deployedAtStr
			}
		}

		collector := CollectorResponse{
			CollectorID:   collectorID,
			CollectorName: collectorName,
			TargetType:    targetTypeFromMap,
			Status:        status,
			DeployedAt:    deployedAt,
		}

		if configPath, ok := collectorMap["config_path"].(string); ok {
			collector.ConfigPath = configPath
		}

		if enrichWithAgentWork {
			cs.enrichCollectorWithAgentWork(ctx, &collector, collectorID)
		}

		collectors = append(collectors, collector)
	}

	return &ListCollectorsResponse{
		TotalCount: len(collectors),
		Collectors: collectors,
	}, nil
}

// ListConnectedAgents lists connected OTEL agents
func (cs *CollectorService) ListConnectedAgents(ctx context.Context, enrichWithAgentWork bool) (*ListConnectedAgentsResponse, error) {
	if cs.otelClient == nil {
		return nil, fmt.Errorf("OTEL client not initialized")
	}

	agents, err := cs.otelClient.ListAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	// Convert to response format and enrich with agent work
	responses := make([]ConnectedAgentResponse, 0, len(agents))
	for _, agent := range agents {
		response := ConnectedAgentResponse{
			ID:          agent.ID,
			Name:        agent.Name,
			Status:      agent.Status,
			Version:     agent.Version,
			LastSeen:    agent.LastSeen,
			GroupID:     agent.GroupID,
			GroupName:   agent.GroupName,
			Description: agent.Description,
		}

		if enrichWithAgentWork {
			cs.enrichAgentWithAgentWork(ctx, &response, agent.ID)
		}

		responses = append(responses, response)
	}

	return &ListConnectedAgentsResponse{
		TotalCount: len(responses),
		Agents:     responses,
	}, nil
}

// GetCollector gets a collector by ID, checking both deployed collectors and connected agents
func (cs *CollectorService) GetCollector(ctx context.Context, collectorID string, enrichWithAgentWork bool) (*CollectorResponse, error) {
	if collectorID == "" {
		return nil, fmt.Errorf("collector ID cannot be empty")
	}

	// Try to find in deployed collectors
	tool := otelTools.GetListDeployedCollectorsTool()
	result, err := tool.Handler(json.RawMessage("{}"))
	if err != nil {
		return nil, fmt.Errorf("failed to list deployed collectors: %w", err)
	}

	// Extract collectors from result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	// Handle collectors field - could be nil, missing, or different types
	var collectorsData []interface{}
	if collectorsField, exists := resultMap["collectors"]; exists && collectorsField != nil {
		switch v := collectorsField.(type) {
		case []interface{}:
			collectorsData = v
		case []map[string]interface{}:
			// Convert []map[string]interface{} to []interface{}
			collectorsData = make([]interface{}, len(v))
			for i, item := range v {
				collectorsData[i] = item
			}
		default:
			// If it's not a recognized type, treat as empty list
			collectorsData = []interface{}{}
		}
	} else {
		// No collectors field or it's nil - return empty list
		collectorsData = []interface{}{}
	}

	var foundCollector map[string]interface{}
	for _, c := range collectorsData {
		collectorMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		// Safely extract collector_id for comparison
		currCollectorID, ok := collectorMap["collector_id"].(string)
		if !ok {
			continue
		}

		if currCollectorID == collectorID {
			foundCollector = collectorMap
			break
		}
	}

	if foundCollector == nil {
		// Try as connected agent ID
		if cs.otelClient != nil {
			agents, err := cs.otelClient.ListAgents()
			if err == nil {
				for _, agent := range agents {
					if agent.ID == collectorID {
						// Convert to collector format
						collector := CollectorResponse{
							CollectorID:   agent.ID,
							CollectorName: agent.Name,
							TargetType:    "connected",
							Status:        agent.Status,
						}

						if enrichWithAgentWork {
							cs.enrichCollectorWithAgentWork(ctx, &collector, collectorID)
						}

						return &collector, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("collector not found")
	}

	// Safely extract fields with type assertions
	collectorIDFromMap, _ := foundCollector["collector_id"].(string)
	collectorName, _ := foundCollector["collector_name"].(string)
	targetType, _ := foundCollector["target_type"].(string)
	status, _ := foundCollector["status"].(string)

	var deployedAt string
	if deployedAtVal, ok := foundCollector["deployed_at"]; ok {
		if deployedAtStr, ok := deployedAtVal.(string); ok {
			deployedAt = deployedAtStr
		}
	}

	collector := CollectorResponse{
		CollectorID:   collectorIDFromMap,
		CollectorName: collectorName,
		TargetType:    targetType,
		Status:        status,
		DeployedAt:    deployedAt,
	}

	if configPath, ok := foundCollector["config_path"].(string); ok {
		collector.ConfigPath = configPath
	}

	if enrichWithAgentWork {
		cs.enrichCollectorWithAgentWork(ctx, &collector, collectorID)
	}

	return &collector, nil
}

// GetCollectorConfig gets the configuration for a collector
func (cs *CollectorService) GetCollectorConfig(ctx context.Context, collectorID string, enrichWithAgentWork bool) (*CollectorConfigResponse, error) {
	if collectorID == "" {
		return nil, fmt.Errorf("collector ID cannot be empty")
	}

	if cs.otelClient == nil {
		return nil, fmt.Errorf("OTEL client not initialized")
	}

	config, err := cs.otelClient.GetAgentConfig(collectorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent config: %w", err)
	}

	response := &CollectorConfigResponse{
		ConfigID:      config.ID,
		ConfigName:    config.Name,
		ConfigVersion: fmt.Sprintf("%d", config.Version),
		YAMLContent:   config.Content,
	}

	if enrichWithAgentWork {
		resourceType := storage.ResourceTypeCollector
		works, err := cs.agentWorkService.GetAgentWorkByResource(ctx, resourceType, collectorID)
		if err == nil && len(works) > 0 {
			response.AgentWork = works
		}
	}

	return response, nil
}

// UpdateCollectorConfig updates the configuration for a collector
func (cs *CollectorService) UpdateCollectorConfig(ctx context.Context, collectorID string, req UpdateCollectorConfigRequest) (*CollectorConfigResponse, error) {
	if collectorID == "" {
		return nil, fmt.Errorf("collector ID cannot be empty")
	}

	if req.YAMLConfig == "" {
		return nil, fmt.Errorf("yaml_config is required")
	}

	if cs.otelClient == nil {
		return nil, fmt.Errorf("OTEL client not initialized")
	}

	if err := cs.otelClient.UpdateAgentConfig(collectorID, req.YAMLConfig); err != nil {
		return nil, fmt.Errorf("failed to update agent config: %w", err)
	}

	// Get updated config
	return cs.GetCollectorConfig(ctx, collectorID, true)
}

// DeployCollector deploys a new collector
func (cs *CollectorService) DeployCollector(ctx context.Context, req DeployCollectorRequest) (map[string]interface{}, error) {
	if req.CollectorName == "" {
		return nil, fmt.Errorf("collector_name is required")
	}
	if req.TargetType == "" {
		req.TargetType = "docker" // default
	}
	if req.YAMLConfig == "" {
		return nil, fmt.Errorf("yaml_config is required")
	}

	// Use the deploy_otel_collector tool logic
	tool := otelTools.GetDeployCollectorTool()

	input := map[string]interface{}{
		"collector_name": req.CollectorName,
		"target_type":    req.TargetType,
		"yaml_config":    req.YAMLConfig,
	}
	if req.Parameters != nil {
		input["parameters"] = req.Parameters
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	result, err := tool.Handler(inputJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy collector: %w", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	return resultMap, nil
}

// StopCollector stops a deployed collector
func (cs *CollectorService) StopCollector(ctx context.Context, collectorID string, targetType string) (map[string]interface{}, error) {
	if collectorID == "" {
		return nil, fmt.Errorf("collector ID cannot be empty")
	}

	// Get target_type from parameter (default to docker)
	if targetType == "" {
		targetType = "docker"
	}

	// Use the stop_otel_collector tool logic
	tool := otelTools.GetStopCollectorTool()

	input := map[string]interface{}{
		"target_type":  targetType,
		"collector_id": collectorID,
		"parameters":   map[string]interface{}{},
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	result, err := tool.Handler(inputJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to stop collector: %w", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	return resultMap, nil
}

// GetCollectorLogs gets logs for a Docker collector
func (cs *CollectorService) GetCollectorLogs(ctx context.Context, collectorID string, tail int) (*CollectorLogsResponse, error) {
	if collectorID == "" {
		return nil, fmt.Errorf("collector ID cannot be empty")
	}

	if tail <= 0 {
		tail = 100 // default
	}

	// First, get collector info to determine target type
	collector, err := cs.GetCollector(ctx, collectorID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get collector: %w", err)
	}

	// Only support Docker collectors for now (container logs)
	if collector.TargetType != "docker" {
		return nil, fmt.Errorf("logs only available for Docker collectors")
	}

	// Construct container name: otel-collector-{collectorID}
	containerName := fmt.Sprintf("otel-collector-%s", collectorID)

	// Use dockerclient to get logs
	dockerClient, err := dc.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	logs, err := dockerClient.GetContainerLogs(ctx, containerName, tail)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return &CollectorLogsResponse{
		Logs: logs,
		Tail: tail,
	}, nil
}

// enrichCollectorWithAgentWork enriches a collector response with agent work
func (cs *CollectorService) enrichCollectorWithAgentWork(ctx context.Context, collector *CollectorResponse, collectorID string) {
	if cs.agentWorkService == nil {
		return
	}

	resourceType := storage.ResourceTypeCollector
	works, err := cs.agentWorkService.GetAgentWorkByResource(ctx, resourceType, collectorID)
	if err == nil && len(works) > 0 {
		collector.AgentWork = works
	}
}

// enrichAgentWithAgentWork enriches a connected agent response with agent work
func (cs *CollectorService) enrichAgentWithAgentWork(ctx context.Context, agent *ConnectedAgentResponse, agentID string) {
	if cs.agentWorkService == nil {
		return
	}

	resourceType := storage.ResourceTypeCollector
	works, err := cs.agentWorkService.GetAgentWorkByResource(ctx, resourceType, agentID)
	if err == nil && len(works) > 0 {
		agent.AgentWork = works
	}
}

