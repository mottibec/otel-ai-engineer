package deployers

import "time"

// TargetType represents the deployment target
type TargetType string

const (
	TargetDocker   TargetType = "docker"
	TargetRemote   TargetType = "remote"
	TargetK8s      TargetType = "kubernetes"
	TargetLocal    TargetType = "local"
)

// DeploymentConfig holds configuration for deploying a collector
type DeploymentConfig struct {
	TargetType     TargetType              `json:"target_type"`
	CollectorName string                   `json:"collector_name"`
	YAMLConfig     string                   `json:"yaml_config"`
	Parameters     map[string]interface{}   `json:"parameters"`
}

// DeploymentResult contains information about a deployment
type DeploymentResult struct {
	Success      bool      `json:"success"`
	CollectorID string    `json:"collector_id"`
	TargetType  string    `json:"target_type"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
	DeployedAt  time.Time `json:"deployed_at"`
}

// CollectorInfo contains information about a deployed collector
type CollectorInfo struct {
	CollectorID  string    `json:"collector_id"`
	CollectorName string   `json:"collector_name"`
	TargetType   string    `json:"target_type"`
	Status       string    `json:"status"`
	DeployedAt   time.Time `json:"deployed_at"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	ConfigPath   string    `json:"config_path,omitempty"`
}

// Deployer is the interface all deployment targets must implement
type Deployer interface {
	// Deploy deploys a new collector instance
	Deploy(config DeploymentConfig) (*DeploymentResult, error)
	
	// Stop stops a running collector
	Stop(collectorID string, params map[string]interface{}) error
	
	// List lists all running collectors for this target
	List() ([]CollectorInfo, error)
	
	// GetTargetType returns the target type this deployer handles
	GetTargetType() TargetType
}

