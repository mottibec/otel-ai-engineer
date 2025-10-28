package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	grafanaTools "github.com/mottibechhofer/otel-ai-engineer/tools/grafana"
)

// BackendAgent is specialized for backend connectivity and validation
type BackendAgent struct {
	*Agent
}

// NewBackendAgent creates a new backend agent with backend management tools
func NewBackendAgent(client *anthropic.Client, logLevel config.LogLevel) (*BackendAgent, error) {
	systemPrompt := `You are an expert observability backend connectivity assistant.

Your capabilities:
- Connect to observability backends (Grafana, Prometheus, Jaeger, etc.)
- Validate backend connectivity and health
- Configure datasources and exporters
- Setup exporters in collectors
- Verify data flow from collectors to backends
- Create visualizations and dashboards

- **Task Delegation**:
  - Use the 'handoff_task' tool when a task is better suited for another agent
  - Delegate infrastructure monitoring setup to 'infrastructure' agent
  - Delegate service instrumentation to 'instrumentation' agent
  - Delegate collector pipeline configuration to 'pipeline' agent
  - The handoff is blocking - wait for completion before proceeding

Supported backends:
1. **Grafana**: Visualization and alerting
   - Configure OTLP, Prometheus, Loki, Tempo datasources
   - Create datasources automatically
   - Validate datasource connectivity
   - Generate pre-built dashboards

2. **Prometheus**: Metrics storage
   - Validate Prometheus endpoint
   - Configure Prometheus remote write
   - Verify metric ingestion

3. **Jaeger/Tempo**: Trace storage
   - Validate trace endpoint
   - Configure OTLP exporter
   - Verify trace ingestion

4. **Custom Backends**:
   - Validate HTTP/OTLP endpoints
   - Configure custom exporters
   - Test connectivity

Connection tasks:
1. **Validation**:
   - Test backend URL accessibility
   - Validate credentials
   - Check endpoint health
   - Verify authentication

2. **Configuration**:
   - Setup exporters in collector config
   - Configure datasources (Grafana)
   - Configure resource attributes
   - Set appropriate endpoints

3. **Data Flow Verification**:
   - Send test telemetry data
   - Verify data arrives at backend
   - Check data completeness
   - Validate timestamps and attributes

4. **Health Monitoring**:
   - Regular health checks
   - Monitor connection status
   - Track data flow metrics
   - Alert on connectivity issues

Best Practices:
- Use secure authentication
- Validate credentials before configuring
- Test connectivity regularly
- Monitor data flow
- Handle connection failures gracefully`

	// Get Grafana tools
	allTools := tools.GetFileSystemTools()
	allTools = append(allTools, grafanaTools.GetGrafanaTools(nil)...) // Will need proper Docker client

	agent := NewAgent(Config{
		Name:         "BackendAgent",
		Description:  "An AI agent specialized for observability backend connectivity and validation",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        allTools,
	})

	return &BackendAgent{Agent: agent}, nil
}

// ConnectBackend connects and validates an observability backend
func (ba *BackendAgent) ConnectBackend(ctx context.Context, backendName string, backendType string, url string) *RunResult {
	prompt := fmt.Sprintf(`Connect to and validate an observability backend.

Backend details:
- Name: %s
- Type: %s
- URL: %s

Tasks:
1. Validate the backend URL is accessible
2. Check authentication credentials
3. Verify backend health endpoint
4. Configure exporter for %s
5. Send test telemetry to verify data flow
6. Monitor backend connectivity

Start by testing the backend URL connectivity and authentication.`, backendName, backendType, url, backendType)

	return ba.Run(ctx, prompt)
}
