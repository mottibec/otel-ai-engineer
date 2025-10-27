package deployers

import (
	"fmt"
)

// RemoteDeployer handles configuration of existing remote Grafana instances
type RemoteDeployer struct {
	url      string
	username string
	password string
}

// NewRemoteDeployer creates a new remote deployer
func NewRemoteDeployer() (*RemoteDeployer, error) {
	// Note: In a real implementation, this would:
	// 1. Accept credentials for the remote Grafana instance
	// 2. Verify connectivity
	// 3. Validate authentication
	// 4. Return configured deployer

	return &RemoteDeployer{}, nil
}

// GetTargetType returns the target type
func (d *RemoteDeployer) GetTargetType() TargetType {
	return TargetRemote
}

// Deploy configures a remote Grafana instance (no deployment, just configuration)
func (d *RemoteDeployer) Deploy(config GrafanaDeploymentConfig) (*GrafanaDeploymentResult, error) {
	// TODO: Implement remote configuration
	// This would involve:
	// 1. Connect to remote Grafana instance
	// 2. Configure datasources programmatically
	// 3. Upload dashboards via API
	// 4. Configure alert rules
	// 5. Return configuration result

	return nil, fmt.Errorf("remote configuration not yet implemented")
}

// Stop stops configuration of a remote Grafana instance
func (d *RemoteDeployer) Stop(instanceID string, params map[string]interface{}) error {
	// Note: For remote targets, "stop" means to remove configured resources
	// This would involve:
	// 1. Remove dashboards
	// 2. Remove datasources (optionally)
	// 3. Remove alert rules
	// 4. Clean up provisioning

	return fmt.Errorf("remote stop not yet implemented")
}

// List lists remote Grafana instances
func (d *RemoteDeployer) List() ([]GrafanaInstanceInfo, error) {
	// TODO: Implement remote list
	// For remote instances, this would return:
	// - List of configured/detected remote Grafana instances
	// - Their connection status and configuration state

	return nil, fmt.Errorf("remote list not yet implemented")
}
