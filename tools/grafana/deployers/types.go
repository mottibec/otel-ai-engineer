package deployers

import "time"

// TargetType represents the deployment target for Grafana
type TargetType string

const (
	TargetDocker TargetType = "docker"
	TargetK8s    TargetType = "kubernetes"
	TargetRemote TargetType = "remote"
)

// GrafanaDeploymentConfig holds configuration for deploying Grafana
type GrafanaDeploymentConfig struct {
	TargetType    TargetType             `json:"target_type"`
	InstanceName  string                 `json:"instance_name"`
	AdminUser     string                 `json:"admin_user"`
	AdminPassword string                 `json:"admin_password"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// GrafanaDeploymentResult contains information about a Grafana deployment
type GrafanaDeploymentResult struct {
	Success    bool      `json:"success"`
	InstanceID string    `json:"instance_id"`
	TargetType string    `json:"target_type"`
	Status     string    `json:"status"`
	Message    string    `json:"message,omitempty"`
	URL        string    `json:"url,omitempty"`
	APIKey     string    `json:"api_key,omitempty"`
	DeployedAt time.Time `json:"deployed_at"`
}

// GrafanaInstanceInfo contains information about a deployed Grafana instance
type GrafanaInstanceInfo struct {
	InstanceID   string    `json:"instance_id"`
	InstanceName string    `json:"instance_name"`
	TargetType   string    `json:"target_type"`
	Status       string    `json:"status"`
	URL          string    `json:"url,omitempty"`
	DeployedAt   time.Time `json:"deployed_at"`
	StartedAt    time.Time `json:"started_at,omitempty"`
}

// GrafanaDeployer is the interface all deployment targets must implement
type GrafanaDeployer interface {
	// Deploy deploys a new Grafana instance
	Deploy(config GrafanaDeploymentConfig) (*GrafanaDeploymentResult, error)

	// Stop stops and removes a Grafana instance
	Stop(instanceID string, params map[string]interface{}) error

	// List lists all running Grafana instances for this target
	List() ([]GrafanaInstanceInfo, error)

	// GetTargetType returns the target type this deployer handles
	GetTargetType() TargetType
}
