package deployers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dc "github.com/mottibechhofer/otel-ai-engineer/tools/dockerclient"
)

// DockerDeployer handles Docker-based Grafana deployments
type DockerDeployer struct {
	dockerClient *dc.Client
}

// NewDockerDeployer creates a new Docker deployer
func NewDockerDeployer(dockerClient *dc.Client) (*DockerDeployer, error) {
	if dockerClient == nil {
		return nil, fmt.Errorf("Docker client cannot be nil")
	}

	return &DockerDeployer{
		dockerClient: dockerClient,
	}, nil
}

// GetTargetType returns the target type
func (d *DockerDeployer) GetTargetType() TargetType {
	return TargetDocker
}

// Deploy deploys Grafana as a Docker container
func (d *DockerDeployer) Deploy(config GrafanaDeploymentConfig) (*GrafanaDeploymentResult, error) {
	ctx := context.Background()

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
	networkName := "otel-network"
	if net, ok := config.Parameters["network"].(string); ok && net != "" {
		networkName = net
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

	// Ensure network exists
	if err := d.dockerClient.EnsureNetwork(ctx, networkName); err != nil {
		return nil, fmt.Errorf("failed to ensure network: %w", err)
	}

	// Build environment variables
	env := []string{
		fmt.Sprintf("GF_SECURITY_ADMIN_USER=%s", adminUser),
		fmt.Sprintf("GF_SECURITY_ADMIN_PASSWORD=%s", adminPassword),
		"GF_AUTH_ANONYMOUS_ENABLED=false",
		"GF_SERVER_ROOT_URL=http://localhost:3000/",
	}

	// Build port bindings
	portBinding := fmt.Sprintf("%s:3000", port)
	portMap, err := dc.CreatePortMap([]string{portBinding})
	if err != nil {
		return nil, fmt.Errorf("failed to create port map: %w", err)
	}

	// Build volume binds
	binds := []string{fmt.Sprintf("%s:/etc/grafana/provisioning:ro", provisionDir)}

	// Create container config
	containerConfig := dc.CreateContainerConfig(image, env, nil)

	// Create host config
	hostConfig := dc.CreateHostConfig(portMap, binds, networkName, "unless-stopped")

	// Create network config
	networkingConfig, err := dc.CreateNetworkConfig(networkName)
	if err != nil {
		return nil, fmt.Errorf("failed to create network config: %w", err)
	}

	cli := d.dockerClient.GetClient()

	// Create container
	createResp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := cli.ContainerStart(ctx, createResp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait a moment for container to start
	time.Sleep(2 * time.Second)

	// Check container status
	status, err := d.dockerClient.GetContainerStatus(ctx, containerName)
	if err != nil {
		logs, logErr := d.dockerClient.GetContainerLogs(ctx, containerName, 50)
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
	ctx := context.Background()
	containerName := fmt.Sprintf("grafana-%s", instanceID)
	cli := d.dockerClient.GetClient()

	// Check if container exists
	exists, err := d.dockerClient.ContainerExists(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to check container existence: %w", err)
	}

	if !exists {
		// Container doesn't exist, nothing to do
		return nil
	}

	// Stop the container
	if err := cli.ContainerStop(ctx, containerName, container.StopOptions{}); err != nil {
		// Log but continue to remove
	}

	// Remove the container
	if err := cli.ContainerRemove(ctx, containerName, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// List lists all running Grafana containers
func (d *DockerDeployer) List() ([]GrafanaInstanceInfo, error) {
	ctx := context.Background()
	cli := d.dockerClient.GetClient()

	filtersArgs := filters.NewArgs()
	filtersArgs.Add("name", "grafana-")

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filtersArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var instances []GrafanaInstanceInfo
	for _, cnt := range containers {
		// Extract instance ID from container name (grafana-{id})
		instanceID := strings.TrimPrefix(cnt.Names[0], "/grafana-")

		instances = append(instances, GrafanaInstanceInfo{
			InstanceID:   instanceID,
			InstanceName: instanceID,
			TargetType:   string(TargetDocker),
			Status:       cnt.Status,
			DeployedAt:   time.Unix(cnt.Created, 0),
		})
	}

	return instances, nil
}

// generateAPIKey generates a random API key for Grafana
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
