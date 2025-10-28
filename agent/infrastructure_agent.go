package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	otelTools "github.com/mottibechhofer/otel-ai-engineer/tools/otel"
)

// InfrastructureAgent is specialized for infrastructure monitoring setup
type InfrastructureAgent struct {
	*Agent
}

// NewInfrastructureAgent creates a new infrastructure agent with collector deployment tools
func NewInfrastructureAgent(client *anthropic.Client, logLevel config.LogLevel) (*InfrastructureAgent, error) {
	systemPrompt := `You are an expert infrastructure monitoring assistant specializing in OpenTelemetry collector configuration and deployment.

Your capabilities:
- Deploy OpenTelemetry collectors
- Configure infrastructure-specific receivers
- Setup host metrics collection
- Monitor databases, caches, queues, and other infrastructure components
- Configure custom receivers and processors

- **Task Delegation**:
  - Use the 'handoff_task' tool when a task is better suited for another agent
  - Delegate service instrumentation to 'instrumentation' agent
  - Delegate backend connectivity validation to 'backend' agent
  - Delegate collector pipeline configuration to 'pipeline' agent
  - The handoff is blocking - wait for completion before proceeding

Supported receiver types:
1. **Host Metrics**: CPU, memory, disk, network utilization
2. **Database Receivers**:
   - postgresql (query metrics, connection stats)
   - mysql (performance metrics, query statistics)
   - mongodb (operations, connection pool)
3. **Cache Receivers**:
   - redis (performance, memory, evictions)
   - memcached (operations, memory)
4. **Message Queue Receivers**:
   - kafka (consumer lag, broker metrics)
   - rabbitmq (queues, messages)
5. **HTTP Receivers**: webhooks, metrics endpoints

Deployment process:
1. **Receiver Selection**:
   - Identify component type
   - Select appropriate receiver
   - Determine required configuration

2. **Collector Configuration**:
   - Create collector YAML config
   - Configure receivers with connection details
   - Setup processors (batch, resource, transform)
   - Configure exporters

3. **Deployment**:
   - Deploy collector to appropriate target
   - Verify receiver connectivity
   - Check metrics are being collected
   - Monitor collector health

Best Practices:
- Use receivers that don't require code changes
- Minimize collector overhead on infrastructure
- Configure appropriate scrape intervals
- Set meaningful resource attributes
- Handle authentication properly`

	// Get OTEL tools
	allTools := tools.GetFileSystemTools()
	allTools = append(allTools, otelTools.GetOtelTools(nil)...) // Will need proper OTEL client

	agent := NewAgent(Config{
		Name:         "InfrastructureAgent",
		Description:  "An AI agent specialized for infrastructure monitoring with OpenTelemetry collectors",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        allTools,
	})

	return &InfrastructureAgent{Agent: agent}, nil
}

// SetupInfrastructureMonitoring configures infrastructure monitoring from a plan component
func (ia *InfrastructureAgent) SetupInfrastructureMonitoring(ctx context.Context, componentName string, componentType string, receiverType string, host string) *RunResult {
	prompt := fmt.Sprintf(`Setup infrastructure monitoring for an infrastructure component.

Component details:
- Name: %s
- Type: %s
- Receiver: %s
- Host: %s

Tasks:
1. Generate collector configuration YAML with the %s receiver
2. Configure receiver connection to %s
3. Set up appropriate processors (batch, resource)
4. Deploy collector if not already running
5. Verify metrics are being collected
6. Test receiver connectivity

Provide the complete collector configuration and deployment status.`, componentName, componentType, receiverType, host, receiverType, host)

	return ia.Run(ctx, prompt)
}
