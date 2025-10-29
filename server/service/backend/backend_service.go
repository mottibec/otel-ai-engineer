package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/grafanaclient"
	grafanaTools "github.com/mottibechhofer/otel-ai-engineer/tools/grafana"
	"github.com/mottibechhofer/otel-ai-engineer/server/service"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// BackendService handles business logic for backend management
type BackendService struct {
	storage         storage.Storage
	agentWorkService *service.AgentWorkService
}

// NewBackendService creates a new backend service
func NewBackendService(stor storage.Storage, agentWorkService *service.AgentWorkService) *BackendService {
	return &BackendService{
		storage:          stor,
		agentWorkService: agentWorkService,
	}
}

// ListBackends retrieves all backends with optional enrichment
func (bs *BackendService) ListBackends(ctx context.Context, enrichWithAgentWork bool) ([]BackendResponse, error) {
	backends, err := bs.storage.ListAllBackends()
	if err != nil {
		return nil, fmt.Errorf("failed to list backends: %w", err)
	}

	responses := make([]BackendResponse, 0, len(backends))
	for _, backend := range backends {
		response := BackendResponse{
			Backend: backend,
		}

		if enrichWithAgentWork {
			bs.enrichWithAgentWork(ctx, &response)
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetBackend retrieves a backend by ID with optional enrichment
func (bs *BackendService) GetBackend(ctx context.Context, backendID string, enrichWithAgentWork bool) (*BackendResponse, error) {
	if backendID == "" {
		return nil, fmt.Errorf("backend ID cannot be empty")
	}

	backend, err := bs.storage.GetBackend(backendID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend: %w", err)
	}

	response := &BackendResponse{
		Backend: backend,
	}

	if enrichWithAgentWork {
		bs.enrichWithAgentWorkForBackend(ctx, response, backendID)
	}

	return response, nil
}

// CreateBackend creates a new backend
func (bs *BackendService) CreateBackend(ctx context.Context, req CreateBackendRequest) (*BackendResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.URL == "" {
		return nil, fmt.Errorf("url is required")
	}
	if req.BackendType == "" {
		return nil, fmt.Errorf("backend_type is required")
	}

	// Generate ID
	backendID := fmt.Sprintf("backend-%d", time.Now().UnixNano())

	// Store credentials as JSON if provided
	credentialsJSON := req.Credentials
	if credentialsJSON == "" && (req.Username != "" || req.Password != "") {
		creds := map[string]string{
			"username": req.Username,
			"password": req.Password,
		}
		credsBytes, err := json.Marshal(creds)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal credentials: %w", err)
		}
		credentialsJSON = string(credsBytes)
	}

	// Serialize config if provided
	configJSON := ""
	if req.Config != nil {
		configBytes, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		configJSON = string(configBytes)
	}

	backend := &storage.Backend{
		ID:           backendID,
		PlanID:       req.PlanID,
		BackendType:  req.BackendType,
		Name:         req.Name,
		URL:          req.URL,
		Credentials:  credentialsJSON,
		HealthStatus: "unknown",
		Config:       configJSON,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := bs.storage.CreateBackend(backend); err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	response := &BackendResponse{
		Backend: backend,
	}

	// Enrich with agent work
	bs.enrichWithAgentWorkForBackend(ctx, response, backendID)

	return response, nil
}

// UpdateBackend updates an existing backend
func (bs *BackendService) UpdateBackend(ctx context.Context, backendID string, req UpdateBackendRequest) (*BackendResponse, error) {
	if backendID == "" {
		return nil, fmt.Errorf("backend ID cannot be empty")
	}

	backend, err := bs.storage.GetBackend(backendID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend: %w", err)
	}

	// Update fields
	if req.Name != nil {
		backend.Name = *req.Name
	}
	if req.URL != nil {
		backend.URL = *req.URL
	}
	if req.Credentials != nil {
		backend.Credentials = *req.Credentials
	} else if req.Username != nil || req.Password != nil {
		// Update credentials from username/password
		var creds map[string]string
		if backend.Credentials != "" {
			if err := json.Unmarshal([]byte(backend.Credentials), &creds); err != nil {
				// If unmarshal fails, create new map
				creds = make(map[string]string)
			}
		}
		if creds == nil {
			creds = make(map[string]string)
		}
		if req.Username != nil {
			creds["username"] = *req.Username
		}
		if req.Password != nil {
			creds["password"] = *req.Password
		}
		credsBytes, err := json.Marshal(creds)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal credentials: %w", err)
		}
		backend.Credentials = string(credsBytes)
	}
	if req.Config != nil {
		configBytes, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		backend.Config = string(configBytes)
	}
	if req.HealthStatus != nil {
		backend.HealthStatus = *req.HealthStatus
		now := time.Now()
		backend.LastCheck = &now
	}

	backend.UpdatedAt = time.Now()

	if err := bs.storage.UpdateBackend(backendID, backend); err != nil {
		return nil, fmt.Errorf("failed to update backend: %w", err)
	}

	// Get updated backend
	updatedBackend, err := bs.storage.GetBackend(backendID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated backend: %w", err)
	}

	response := &BackendResponse{
		Backend: updatedBackend,
	}

	// Enrich with agent work
	bs.enrichWithAgentWorkForBackend(ctx, response, backendID)

	return response, nil
}

// DeleteBackend deletes a backend
func (bs *BackendService) DeleteBackend(ctx context.Context, backendID string) error {
	if backendID == "" {
		return fmt.Errorf("backend ID cannot be empty")
	}

	if err := bs.storage.DeleteBackend(backendID); err != nil {
		return fmt.Errorf("failed to delete backend: %w", err)
	}

	return nil
}

// TestConnection tests the connection to a backend
func (bs *BackendService) TestConnection(ctx context.Context, backendID string, req TestConnectionRequest) (*TestConnectionResult, error) {
	if backendID == "" {
		return nil, fmt.Errorf("backend ID cannot be empty")
	}

	backend, err := bs.storage.GetBackend(backendID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend: %w", err)
	}

	url := req.URL
	if url == "" {
		url = backend.URL
	}

	username := req.Username
	password := req.Password

	// If not provided, try to get from stored credentials
	if username == "" || password == "" {
		if backend.Credentials != "" {
			var creds map[string]string
			if err := json.Unmarshal([]byte(backend.Credentials), &creds); err == nil {
				if username == "" {
					username = creds["username"]
				}
				if password == "" {
					password = creds["password"]
				}
			}
		}
	}

	// Test connection based on backend type
	var healthy bool
	var errorMsg string
	var datasources []interface{}

	switch backend.BackendType {
	case "grafana":
		if username == "" || password == "" {
			return nil, fmt.Errorf("username and password required for Grafana")
		}

		client := grafanaclient.NewClientWithAuth(url, username, password)

		// Try to get health and list datasources
		health, err := client.GetHealth()
		if err != nil {
			healthy = false
			errorMsg = err.Error()
		} else {
			healthy = true
			// Try to list datasources as additional validation
			ds, err := client.ListDatasources()
			if err == nil {
				// Convert to []interface{} for JSON serialization
				datasources = make([]interface{}, len(ds))
				for i, d := range ds {
					datasources[i] = d
				}
			}
			_ = health // Health check successful
		}

	case "prometheus", "jaeger", "custom":
		// Basic HTTP health check
		healthResp, err := http.Get(url + "/health")
		if err != nil {
			healthy = false
			errorMsg = err.Error()
		} else {
			healthResp.Body.Close()
			healthy = healthResp.StatusCode < 400
			if !healthy {
				errorMsg = fmt.Sprintf("HTTP %d", healthResp.StatusCode)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported backend type: %s", backend.BackendType)
	}

	// Update backend health status
	now := time.Now()
	status := "unhealthy"
	if healthy {
		status = "healthy"
	}
	backend.HealthStatus = status
	backend.LastCheck = &now
	if err := bs.storage.UpdateBackend(backendID, backend); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	result := &TestConnectionResult{
		Healthy: healthy,
		Status:  status,
	}

	if !healthy && errorMsg != "" {
		result.Error = errorMsg
	}

	if len(datasources) > 0 {
		result.Datasources = datasources
	}

	return result, nil
}

// ConfigureGrafanaDatasource configures a Grafana datasource
func (bs *BackendService) ConfigureGrafanaDatasource(ctx context.Context, backendID string, req ConfigureGrafanaDatasourceRequest) (map[string]interface{}, error) {
	if backendID == "" {
		return nil, fmt.Errorf("backend ID cannot be empty")
	}

	backend, err := bs.storage.GetBackend(backendID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend: %w", err)
	}

	if backend.BackendType != "grafana" {
		return nil, fmt.Errorf("only Grafana backends support datasource configuration")
	}

	if req.DatasourceName == "" || req.DatasourceType == "" || req.URL == "" {
		return nil, fmt.Errorf("datasource_name, datasource_type, and url are required")
	}

	// Get credentials
	var username, password string
	if backend.Credentials != "" {
		var creds map[string]string
		if err := json.Unmarshal([]byte(backend.Credentials), &creds); err == nil {
			username = creds["username"]
			password = creds["password"]
		}
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("Grafana credentials not configured. Please set username and password")
	}

	// Use the configure_datasource tool logic
	tool := grafanaTools.GetConfigureDatasourceTool()

	input := map[string]interface{}{
		"grafana_url":      backend.URL,
		"username":         username,
		"password":         password,
		"datasource_name": req.DatasourceName,
		"datasource_type":  req.DatasourceType,
		"url":              req.URL,
	}
	if req.JSONData != nil {
		input["json_data"] = req.JSONData
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	result, err := tool.Handler(inputJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to configure datasource: %w", err)
	}

	// Update backend with datasource UID if returned
	resultMap, ok := result.(map[string]interface{})
	if ok {
		if uid, ok := resultMap["datasource_uid"].(string); ok {
			backend.DatasourceUID = uid
			_ = bs.storage.UpdateBackend(backendID, backend) // Log error but don't fail
		}
	}

	return resultMap, nil
}

// enrichWithAgentWork enriches a backend response with agent work
func (bs *BackendService) enrichWithAgentWork(ctx context.Context, response *BackendResponse) {
	if response.Backend == nil {
		return
	}
	bs.enrichWithAgentWorkForBackend(ctx, response, response.Backend.ID)
}

// enrichWithAgentWorkForBackend enriches a backend response with agent work for a specific backend ID
func (bs *BackendService) enrichWithAgentWorkForBackend(ctx context.Context, response *BackendResponse, backendID string) {
	if bs.agentWorkService == nil {
		return
	}

	resourceType := storage.ResourceTypeBackend
	works, err := bs.agentWorkService.GetAgentWorkByResource(ctx, resourceType, backendID)
	if err == nil && len(works) > 0 {
		response.AgentWork = works
	}
}

