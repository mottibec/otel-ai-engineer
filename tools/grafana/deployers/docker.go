package deployers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DockerDeployer handles Docker-based Grafana deployments
type DockerDeployer struct {
	dockerPath string
}

// NewDockerDeployer creates a new Docker deployer
func NewDockerDeployer() (*DockerDeployer, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	return &DockerDeployer{
		dockerPath: dockerPath,
	}, nil
}

// GetTargetType returns the target type
func (d *DockerDeployer) GetTargetType() TargetType {
	return TargetDocker
}

// Deploy deploys Grafana as a Docker container
func (d *DockerDeployer) Deploy(config GrafanaDeploymentConfig) (*GrafanaDeploymentResult, error) {
	// Generate unique instance ID
	instanceID := fmt.Sprintf("%s-%d", config.InstanceName, time.Now().Unix())
	containerName := fmt.Sprintf("grafana-%s", instanceID)

	// Create temporary provisioning directory
	provisionDir := "/tmp/grafana-provisioning"
	if err := os.MkdirAll(provisionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create provisioning directory: %w", err)
	}

	// Generate API key
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Get deployment parameters
	network := "otel-network"
	if net, ok := config.Parameters["network"].(string); ok && net != "" {
		network = net
	}

	image := "grafana/grafana:latest"
	if img, ok := config.Parameters["image"].(string); ok && img != "" {
		image = img
	}

	port := "3000"
	if p, ok := config.Parameters["port"].(string); ok && p != "" {
		port = p
	}

	adminUser := "admin"
	if config.AdminUser != "" {
		adminUser = config.AdminUser
	}

	adminPassword := "admin"
	if config.AdminPassword != "" {
		adminPassword = config.AdminPassword
	}

	// Build docker run command
	args := []string{
		"run",
		"-d",
		"--name", containerName,
		"--network", network,
		"-p", fmt.Sprintf("%s:3000", port),
		"-e", "GF_SECURITY_ADMIN_USER=" + adminUser,
		"-e", "GF_SECURITY_ADMIN_PASSWORD=" + adminPassword,
		"-e", "GF_AUTH_ANONYMOUS_ENABLED=false",
		"-e", "GF_SERVER_ROOT_URL=http://localhost:3000/",
		"-v", fmt.Sprintf("%s:/etc/grafana/provisioning:ro", provisionDir),
		"--restart", "unless-stopped",
		image,
	}

	// Execute docker run
	cmd := exec.Command(d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start docker container: %s - %w", string(output), err)
	}

	// Wait a moment for container to start
	time.Sleep(2 * time.Second)

	// Check container status
	status, err := d.getContainerStatus(containerName)
	if err != nil {
		logs, logErr := d.getContainerLogs(containerName)
		if logErr != nil {
			return nil, fmt.Errorf("container failed to start: %w (logs unavailable)", err)
		}
		return nil, fmt.Errorf("container failed to start: %w\nLogs: %s", err, logs)
	}

	url := fmt.Sprintf("http://localhost:%s", port)

	return &GrafanaDeploymentResult{
		Success:    true,
		InstanceID: instanceID,
		TargetType: string(TargetDocker),
		Status:     status,
		Message:    fmt.Sprintf("Grafana deployed in container %s", containerName),
		URL:        url,
		APIKey:     apiKey,
		DeployedAt: time.Now(),
	}, nil
}

// Stop stops and removes a Grafana container
func (d *DockerDeployer) Stop(instanceID string, params map[string]interface{}) error {
	containerName := fmt.Sprintf("grafana-%s", instanceID)

	// First, try to stop the container
	stopCmd := exec.Command(d.dockerPath, "stop", containerName)
	if err := stopCmd.Run(); err != nil && !strings.Contains(err.Error(), "No such container") {
		// Log but continue to remove
	}

	// Remove the container
	rmCmd := exec.Command(d.dockerPath, "rm", containerName)
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// List lists all running Grafana containers
func (d *DockerDeployer) List() ([]GrafanaInstanceInfo, error) {
	args := []string{
		"ps",
		"--filter", "name=grafana-",
		"--format", "{{.Names}},{{.Status}},{{.CreatedAt}}",
	}

	cmd := exec.Command(d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var instances []GrafanaInstanceInfo
	if len(strings.TrimSpace(string(output))) == 0 {
		return instances, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}

		containerName := parts[0]
		status := parts[1]

		// Extract instance ID from container name (grafana-{id})
		instanceID := strings.TrimPrefix(containerName, "grafana-")

		instances = append(instances, GrafanaInstanceInfo{
			InstanceID:   instanceID,
			InstanceName: instanceID,
			TargetType:   string(TargetDocker),
			Status:       status,
			DeployedAt:   time.Now(),
		})
	}

	return instances, nil
}

// getContainerStatus gets the status of a container
func (d *DockerDeployer) getContainerStatus(containerName string) (string, error) {
	cmd := exec.Command(d.dockerPath, "inspect", "-f", "{{.State.Status}}", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getContainerLogs retrieves logs from a container
func (d *DockerDeployer) getContainerLogs(containerName string) (string, error) {
	cmd := exec.Command(d.dockerPath, "logs", "--tail", "50", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// generateAPIKey generates a random API key for Grafana
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
