package collector

import (
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// CollectorResponse represents a collector with agent work info
type CollectorResponse struct {
	CollectorID   string               `json:"collector_id"`
	CollectorName string               `json:"collector_name"`
	TargetType    string               `json:"target_type"`
	Status        string               `json:"status"`
	DeployedAt    string               `json:"deployed_at"`
	ConfigPath    string               `json:"config_path,omitempty"`
	AgentWork     []*storage.AgentWork `json:"agent_work,omitempty"`
}

// ConnectedAgentResponse represents a connected OTEL agent
type ConnectedAgentResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Status      string               `json:"status"`
	Version     string               `json:"version"`
	LastSeen    string               `json:"last_seen,omitempty"`
	GroupID     string               `json:"group_id,omitempty"`
	GroupName   string               `json:"group_name,omitempty"`
	Description string               `json:"description,omitempty"`
	AgentWork   []*storage.AgentWork `json:"agent_work,omitempty"`
}

// DeployCollectorRequest represents the request to deploy a collector
type DeployCollectorRequest struct {
	CollectorName string                 `json:"collector_name"`
	TargetType    string                 `json:"target_type"`
	YAMLConfig    string                 `json:"yaml_config"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
}

// UpdateCollectorConfigRequest represents the request to update collector config
type UpdateCollectorConfigRequest struct {
	YAMLConfig string `json:"yaml_config"`
}

// CollectorConfigResponse represents a collector config with agent work
type CollectorConfigResponse struct {
	ConfigID      string               `json:"config_id"`
	ConfigName    string               `json:"config_name"`
	ConfigVersion string               `json:"config_version"`
	YAMLContent   string               `json:"yaml_content"`
	AgentWork     []*storage.AgentWork `json:"agent_work,omitempty"`
}

// ListCollectorsResponse represents the response for listing collectors
type ListCollectorsResponse struct {
	TotalCount int                 `json:"total_count"`
	Collectors []CollectorResponse `json:"collectors"`
}

// ListConnectedAgentsResponse represents the response for listing connected agents
type ListConnectedAgentsResponse struct {
	TotalCount int                     `json:"total_count"`
	Agents     []ConnectedAgentResponse `json:"agents"`
}

// CollectorLogsResponse represents collector logs
type CollectorLogsResponse struct {
	Logs string `json:"logs"`
	Tail int    `json:"tail"`
}

