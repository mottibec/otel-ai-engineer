package dockerclient

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Client wraps the Docker client with helper methods
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client connected to the Docker socket
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

// GetClient returns the underlying Docker client
func (c *Client) GetClient() *client.Client {
	return c.cli
}

// EnsureNetwork ensures a Docker network exists, creating it if necessary
// It handles various edge cases including:
// - Network already exists (returns success)
// - Network created concurrently by another process (handles gracefully)
// - Docker Compose prefixed network names
// - Falls back to CLI if API fails
func (c *Client) EnsureNetwork(ctx context.Context, networkName string) error {
	// Try API-based approach first
	err := c.ensureNetworkViaAPI(ctx, networkName)
	if err == nil {
		return nil
	}

	// If API fails, try CLI fallback
	return c.ensureNetworkViaCLI(ctx, networkName, err)
}

// ensureNetworkViaAPI uses the Docker API to ensure network exists
func (c *Client) ensureNetworkViaAPI(ctx context.Context, networkName string) error {
	// List all networks to check for existence (including Compose-prefixed ones)
	networkListOptions := network.ListOptions{}
	allNetworks, err := c.cli.NetworkList(ctx, networkListOptions)
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	// Check if network exists (exact match or ends with network name)
	for _, net := range allNetworks {
		if net.Name == networkName {
			return nil // Network exists
		}
		// Handle Docker Compose prefixed networks (e.g., "project_otel-network")
		if strings.HasSuffix(net.Name, "_"+networkName) || strings.HasSuffix(net.Name, "-"+networkName) {
			return nil // Network exists with prefix
		}
	}

	// Network doesn't exist, create it
	networkOptions := network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			"com.docker.compose.project": "otel-ai-engineer",
		},
	}

	_, err = c.cli.NetworkCreate(ctx, networkName, networkOptions)
	if err != nil {
		// Check if error is because network was created concurrently
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "network with name") {
			// Network was created by another process, verify it exists now
			time.Sleep(100 * time.Millisecond) // Brief wait for network to be available
			return c.verifyNetworkExists(ctx, networkName)
		}
		return fmt.Errorf("failed to create network: %w", err)
	}

	return nil
}

// verifyNetworkExists verifies that a network exists after a potential race condition
func (c *Client) verifyNetworkExists(ctx context.Context, networkName string) error {
	// Retry a few times with short delays
	for i := 0; i < 5; i++ {
		networkListOptions := network.ListOptions{}
		allNetworks, err := c.cli.NetworkList(ctx, networkListOptions)
		if err == nil {
			for _, net := range allNetworks {
				if net.Name == networkName {
					return nil
				}
				if strings.HasSuffix(net.Name, "_"+networkName) || strings.HasSuffix(net.Name, "-"+networkName) {
					return nil
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("network %s not found after creation", networkName)
}

// ensureNetworkViaCLI uses Docker CLI as a fallback when API fails
func (c *Client) ensureNetworkViaCLI(ctx context.Context, networkName string, apiErr error) error {
	// Check if network exists via CLI
	checkCmd := exec.CommandContext(ctx, "docker", "network", "inspect", networkName)
	if err := checkCmd.Run(); err == nil {
		// Network exists via CLI check
		return nil
	}

	// Network doesn't exist, create it via CLI
	createCmd := exec.CommandContext(ctx, "docker", "network", "create", "--driver", "bridge", networkName)
	output, err := createCmd.CombinedOutput()
	if err != nil {
		// Check if network was created concurrently
		if strings.Contains(string(output), "already exists") {
			return nil // Network exists, success
		}
		return fmt.Errorf("API error: %w; CLI fallback also failed: %s", apiErr, string(output))
	}

	return nil
}

// GetNetworkID returns the network ID for a given network name
func (c *Client) GetNetworkID(ctx context.Context, networkName string) (string, error) {
	networkListOptions := network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	}

	networks, err := c.cli.NetworkList(ctx, networkListOptions)
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}

	if len(networks) == 0 {
		return "", fmt.Errorf("network %s not found", networkName)
	}

	return networks[0].ID, nil
}

// ContainerExists checks if a container with the given name exists
func (c *Client) ContainerExists(ctx context.Context, containerName string) (bool, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return false, err
	}

	return len(containers) > 0, nil
}

// GetContainerStatus gets the status of a container
func (c *Client) GetContainerStatus(ctx context.Context, containerName string) (string, error) {
	inspect, err := c.cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", err
	}

	return inspect.State.Status, nil
}

// GetContainerLogs retrieves logs from a container
func (c *Client) GetContainerLogs(ctx context.Context, containerName string, tail int) (string, error) {
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
		Follow:     false,
	}

	logs, err := c.cli.ContainerLogs(ctx, containerName, logOptions)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := make([]byte, 4096)
	var output strings.Builder
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return output.String(), nil
}

// ParsePortBinding parses a port mapping string like "3000:3000" into nat.Port and nat.PortBinding
func ParsePortBinding(portMapping string) (nat.Port, nat.PortBinding, error) {
	parts := strings.Split(portMapping, ":")
	if len(parts) != 2 {
		return "", nat.PortBinding{}, fmt.Errorf("invalid port mapping: %s", portMapping)
	}

	port, err := nat.NewPort("tcp", parts[1])
	if err != nil {
		return "", nat.PortBinding{}, fmt.Errorf("invalid port: %w", err)
	}

	binding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: parts[0],
	}

	return port, binding, nil
}

// CreatePortMap creates a nat.PortMap from a map of port mappings
func CreatePortMap(portMappings []string) (nat.PortMap, error) {
	portMap := nat.PortMap{}
	for _, mapping := range portMappings {
		port, binding, err := ParsePortBinding(mapping)
		if err != nil {
			return nil, err
		}
		portMap[port] = []nat.PortBinding{binding}
	}
	return portMap, nil
}

// CreateContainerConfig creates a container.Config from parameters
func CreateContainerConfig(image string, env []string, cmd []string) *container.Config {
	return &container.Config{
		Image: image,
		Env:   env,
		Cmd:   cmd,
	}
}

// CreateHostConfig creates a container.HostConfig from parameters
func CreateHostConfig(portMap nat.PortMap, binds []string, networkMode string, restartPolicy string) *container.HostConfig {
	return &container.HostConfig{
		PortBindings: portMap,
		Binds:        binds,
		NetworkMode:  container.NetworkMode(networkMode),
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode(restartPolicy),
		},
	}
}

// CreateNetworkConfig creates a network.NetworkingConfig from network name
func CreateNetworkConfig(networkName string) (*network.NetworkingConfig, error) {
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}, nil
}

// Close closes the Docker client connection
func (c *Client) Close() error {
	return c.cli.Close()
}

