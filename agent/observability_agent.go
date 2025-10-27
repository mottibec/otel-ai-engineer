package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	dc "github.com/mottibechhofer/otel-ai-engineer/tools/dockerclient"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	grafanaTools "github.com/mottibechhofer/otel-ai-engineer/tools/grafana"
	otelTools "github.com/mottibechhofer/otel-ai-engineer/tools/otel"
)

// ObservabilityAgent is a specialized agent for complete observability infrastructure
type ObservabilityAgent struct {
	*Agent
	otelClient *otelclient.OtelClient
}

// NewObservabilityAgent creates a new observability agent with file system, OTEL, and Grafana tools
func NewObservabilityAgent(client *anthropic.Client, logLevel config.LogLevel) (*ObservabilityAgent, error) {
	systemPrompt := `You are an expert observability infrastructure assistant with access to file system operations, OpenTelemetry collector management, and Grafana visualization tools.

Your capabilities:
- **OpenTelemetry Collector Management**:
  - List all connected OpenTelemetry collectors and their status
  - View and inspect collector configurations (YAML format)
  - Update collector configurations remotely via OpAMP protocol
  - Deploy new OpenTelemetry collector instances to Docker, Kubernetes, or remote servers

- **Grafana Visualization Setup**:
  - Deploy Grafana instances (Docker, Kubernetes, or connect to existing)
  - Auto-discover data sources from OpenTelemetry collectors
  - Configure datasources (OTLP, Prometheus, Loki, Tempo)
  - Generate pre-built dashboards (RED/USE metrics)
  - Analyze application code to generate custom dashboards
  - Create intelligent alert rules based on application structure

- **File System Operations**:
  - Read, write, edit configuration files
  - Search for files and directories
  - Analyze application codebases
  - Create and manage configurations

When setting up complete observability:

1. **Initial Assessment**:
   - Analyze the application codebase to understand structure
   - List existing collectors and Grafana instances
   - Identify what observability infrastructure exists

2. **Collector Deployment** (if needed):
   - Deploy OTEL collectors with proper configuration
   - Configure collectors to send data to Lawrence OTLP server
   - Verify collectors are running and connected

3. **Grafana Deployment**:
   - Deploy Grafana instance (Docker recommended for development)
   - Auto-discover available data sources
   - Configure datasources pointing to OTEL endpoints

4. **Dashboard Generation**:
   - Analyze code to detect frameworks and critical operations
   - Generate golden signals dashboards (Rate, Errors, Duration, Utilization, Saturation)
   - Create custom dashboards based on application-specific patterns

5. **Alert Configuration**:
   - Analyze code for critical paths (auth, payments, data processing)
   - Create context-aware alerts based on business logic
   - Set up standard alerts for common issues

Best Practices:
- Use auto-discover tools to reduce manual configuration
- Analyze code before generating dashboards/alerts for better context
- Start with golden signals dashboards, then add custom dashboards
- Create alerts incrementally and test them
- Provide clear explanations of what you're doing and why
- Handle errors gracefully and suggest troubleshooting steps`

	// Get Lawrence API URL from environment
	lawrenceURL := os.Getenv("LAWRENCE_API_URL")
	if lawrenceURL == "" {
		lawrenceURL = "http://lawrence:8080"
	}

	// Create OTEL client
	otelClient := otelclient.NewOtelClient(lawrenceURL, &http.Client{
		Timeout: 30 * time.Second,
	})

	// Create Docker client
	dockerClient, err := dc.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to create Docker client: %v", err)
		// Continue without Docker client - other deployers will fail gracefully
		dockerClient = nil
	}

	// Get all tools
	allTools := []tools.Tool{}
	allTools = append(allTools, tools.GetFileSystemTools()...)
	allTools = append(allTools, otelTools.GetOtelTools(otelClient)...)
	allTools = append(allTools, grafanaTools.GetGrafanaTools(dockerClient)...)

	agent := NewAgent(Config{
		Name:         "ObservabilityAgent",
		Description:  "An AI agent specialized for complete observability infrastructure setup with OTEL collectors, Grafana visualization, and code analysis",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        allTools,
	})

	return &ObservabilityAgent{
		Agent:      agent,
		otelClient: otelClient,
	}, nil
}

// SetupObservabilityStack sets up a complete observability infrastructure
func (oa *ObservabilityAgent) SetupObservabilityStack(ctx context.Context, codebasePath string, targetEnvironment string) *RunResult {
	prompt := fmt.Sprintf(`Set up complete observability infrastructure for my application.

Codebase path: %s
Target environment: %s

Please:
1. Analyze the codebase to understand the application structure
2. Deploy OTEL collectors if needed
3. Deploy Grafana instance
4. Auto-discover and configure data sources
5. Generate appropriate dashboards based on the application
6. Create intelligent alert rules for critical paths

Start by analyzing the codebase and then proceed with deployment and configuration.`, codebasePath, targetEnvironment)

	return oa.Run(ctx, prompt)
}

// GenerateDashboards analyzes code and generates custom dashboards
func (oa *ObservabilityAgent) GenerateDashboards(ctx context.Context, grafanaURL string, codebasePath string, datasourceUID string) *RunResult {
	prompt := fmt.Sprintf(`Generate Grafana dashboards for my application.

Steps:
1. Analyze the code at: %s
2. Detect frameworks and critical operations
3. Generate custom dashboards based on detected patterns
4. Create the dashboards in Grafana at: %s

Focus on:
- HTTP endpoint monitoring
- Database performance
- Application-specific metrics based on framework detection
- Golden signals (RED/USE metrics)

Use datasource UID: %s`, codebasePath, grafanaURL, datasourceUID)

	return oa.Run(ctx, prompt)
}

// CreateAlerts generates intelligent alert rules based on code analysis
func (oa *ObservabilityAgent) CreateAlerts(ctx context.Context, grafanaURL string, codebasePath string, datasourceUID string) *RunResult {
	prompt := fmt.Sprintf(`Create intelligent alert rules for my application.

Steps:
1. Analyze the code at: %s
2. Identify critical paths (auth, payments, data processing, etc.)
3. Create context-aware alerts based on business logic
4. Set appropriate thresholds for each alert type

Use Grafana at: %s with datasource UID: %s`, codebasePath, grafanaURL, datasourceUID)

	return oa.Run(ctx, prompt)
}
