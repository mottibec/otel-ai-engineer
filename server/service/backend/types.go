package backend

import (
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// BackendResponse represents a backend with agent work info
type BackendResponse struct {
	*storage.Backend
	AgentWork []*storage.AgentWork `json:"agent_work,omitempty"`
}

// CreateBackendRequest represents the request to create a backend
type CreateBackendRequest struct {
	BackendType string                 `json:"backend_type"` // "grafana", "prometheus", "jaeger", "custom"
	Name        string                 `json:"name"`
	URL         string                 `json:"url"`
	Username    string                 `json:"username,omitempty"`
	Password    string                 `json:"password,omitempty"`
	Credentials string                 `json:"credentials,omitempty"` // Alternative: encrypted JSON string
	Config      map[string]interface{} `json:"config,omitempty"`
	PlanID      *string                `json:"plan_id,omitempty"`
}

// UpdateBackendRequest represents the request to update a backend
type UpdateBackendRequest struct {
	Name        *string                `json:"name,omitempty"`
	URL         *string                `json:"url,omitempty"`
	Username    *string                `json:"username,omitempty"`
	Password    *string                `json:"password,omitempty"`
	Credentials *string                `json:"credentials,omitempty"`
	Config      *map[string]interface{} `json:"config,omitempty"`
	HealthStatus *string                `json:"health_status,omitempty"`
}

// TestConnectionRequest represents the request to test a backend connection
type TestConnectionRequest struct {
	URL      string `json:"url,omitempty"`       // Optional override
	Username string `json:"username,omitempty"` // Optional override
	Password string `json:"password,omitempty"` // Optional override
}

// TestConnectionResult represents the result of a connection test
type TestConnectionResult struct {
	Healthy    bool                           `json:"healthy"`
	Status     string                         `json:"status"`
	Error      string                         `json:"error,omitempty"`
	Datasources []interface{}                `json:"datasources,omitempty"`
}

// ConfigureGrafanaDatasourceRequest represents the request to configure a Grafana datasource
type ConfigureGrafanaDatasourceRequest struct {
	DatasourceName string                 `json:"datasource_name"`
	DatasourceType string                 `json:"datasource_type"` // "otlp", "prometheus", "loki", "tempo"
	URL            string                 `json:"url"`
	JSONData       map[string]interface{} `json:"json_data,omitempty"`
}

