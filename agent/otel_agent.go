package agent

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	otelTools "github.com/mottibechhofer/otel-ai-engineer/tools/otel"
)

// OtelAgent is a specialized agent for OpenTelemetry collector management
type OtelAgent struct {
	*Agent
	otelClient *otelclient.OtelClient
}

// NewOtelAgent creates a new OTEL management agent with file system and OTEL tools
func NewOtelAgent(client *anthropic.Client, logLevel config.LogLevel) (*OtelAgent, error) {
	systemPrompt := `You are an expert OpenTelemetry collector management assistant with access to both file system operations and OTEL agent management tools.

Your capabilities:
- **OTEL Agent Management**:
  - List all connected OpenTelemetry collectors and their status
  - View and inspect collector configurations (YAML format)
  - Update collector configurations remotely via OpAMP protocol
  - Monitor collector health and status

- **Collector Deployment**:
  - Deploy new OpenTelemetry collector instances to Docker containers (and other targets in future)
  - Stop and remove deployed collectors
  - List all deployed collector instances and their status
  - Manage collector lifecycle from deployment to removal

- **File System Operations**:
  - Read, write, edit configuration files
  - Search for files and directories
  - Create and manage collector configuration files
  - Backup configurations before making changes

When working on OTEL management tasks:

1. **Initial Assessment**:
   - List all connected collectors to understand the environment
   - List all deployed collector instances
   - Check current configurations for each collector

2. **Deployment**:
   - Create collector configuration files with proper OpAMP setup
   - Deploy collectors with appropriate network and environment settings
   - Verify collectors connect to Lawrence OpAMP server after deployment
   - Monitor deployment status

3. **Configuration Management**:
   - Read existing configurations to understand current state
   - Use file operations to create new configuration templates
   - Test configurations before deploying
   - Always backup configurations before making changes

4. **Updates**:
   - Modify configurations incrementally
   - Update collectors via OpAMP protocol
   - Verify changes were applied successfully

5. **Monitoring**:
   - Regularly check collector status
   - Verify collectors are receiving and processing telemetry

Best Practices:
- Always list collectors first to get their IDs
- Use file operations to create and validate configuration files
- Deploy collectors with proper OpAMP configuration pointing to Lawrence server
- Make incremental changes and verify each step
- Backup configurations before updates
- Provide clear explanations of what you're doing and why
- Handle errors gracefully and suggest troubleshooting steps`

	// Get Lawrence API URL from environment
	lawrenceURL := os.Getenv("LAWRENCE_API_URL")
	if lawrenceURL == "" {
		lawrenceURL = "http://lawrence:8080" // Default in docker-compose
	}

	// Create OTEL client
	otelClient := otelclient.NewOtelClient(lawrenceURL, &http.Client{
		Timeout: 30 * time.Second,
	})

	// Get OTEL tools
	otelToolSet := otelTools.GetOtelTools(otelClient)

	// Combine with file system tools for configuration management
	allTools := append(
		tools.GetFileSystemTools(),
		otelToolSet...,
	)

	agent := NewAgent(Config{
		Name:         "OtelAgent",
		Description:  "An AI agent specialized for OpenTelemetry collector management with file system and OTEL tools",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        allTools,
	})

	return &OtelAgent{
		Agent:      agent,
		otelClient: otelClient,
	}, nil
}

// ManageCollectors manages OTEL collectors
func (oa *OtelAgent) ManageCollectors(ctx context.Context, task string) *RunResult {
	return oa.Run(ctx, task)
}

// UpdateCollectorConfig updates a collector's configuration
func (oa *OtelAgent) UpdateCollectorConfig(ctx context.Context, agentID string, configPath string) *RunResult {
	prompt := fmt.Sprintf(`I need to update the configuration for OTEL collector '%s'.

Steps:
1. List all collectors to verify '%s' exists
2. Get the current configuration for this collector
3. Read the configuration file from '%s'
4. Update the collector's configuration with the new configuration
5. Verify the update was successful

Please execute these steps.`, agentID, agentID, configPath)

	return oa.Run(ctx, prompt)
}

// MonitorCollectors monitors the health and status of collectors
func (oa *OtelAgent) MonitorCollectors(ctx context.Context) *RunResult {
	prompt := `Please monitor all connected OTEL collectors:

1. List all collectors and their status
2. Get configuration information for each collector
3. Report on their health and any issues
4. Provide recommendations if any collectors need attention

Please provide a comprehensive status report.`

	return oa.Run(ctx, prompt)
}
