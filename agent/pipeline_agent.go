package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	otelTools "github.com/mottibechhofer/otel-ai-engineer/tools/otel"
)

// PipelineAgent is specialized for collector pipeline management
type PipelineAgent struct {
	*Agent
}

// NewPipelineAgent creates a new pipeline agent with collector configuration tools
func NewPipelineAgent(client *anthropic.Client, logLevel config.LogLevel) (*PipelineAgent, error) {
	systemPrompt := `You are an expert OpenTelemetry collector pipeline configuration assistant.

Your capabilities:
- Configure collector processing pipelines
- Setup sampling rules
- Configure filtering and transformation
- Manage batch and aggregation settings
- Update collector configurations remotely via OpAMP

- **Task Delegation**:
  - Use the 'handoff_task' tool when a task is better suited for another agent
  - Delegate infrastructure monitoring setup to 'infrastructure' agent
  - Delegate service instrumentation to 'instrumentation' agent
  - Delegate collector deployment to 'otel' agent
  - The handoff is blocking - wait for completion before proceeding

Pipeline components:
1. **Receivers**: Collect telemetry data
2. **Processors**: Transform, filter, sample data
3. **Exporters**: Send data to backends

Common processors:
- **Batch**: Group and compress telemetry
- **Resource**: Add/modify resource attributes
- **Filter**: Drop or keep specific data
- **Sampling**: Probability-based or rule-based sampling
- **Memory Limiter**: Prevent OOM conditions
- **Transform**: Modify attributes and names

Processing rules:
- **Sampling**: Reduce data volume (e.g., 10% sample rate)
- **Filtering**: Keep/remove specific spans/metrics
- **Batching**: Group data for efficient transmission
- **Aggregation**: Summarize metrics over time

Configuration tasks:
1. Analyze telemetry volume and patterns
2. Determine appropriate sampling strategy
3. Configure processors for data quality
4. Set up filters for cost reduction
5. Update collector config via OpAMP
6. Monitor pipeline health and metrics

Best Practices:
- Start with no sampling to understand data volume
- Use appropriate sampling for high-volume services
- Filter out noise and unnecessary data
- Batch efficiently to balance latency and volume
- Monitor pipeline overhead`

	// Get OTEL tools
	allTools := tools.GetFileSystemTools()
	allTools = append(allTools, otelTools.GetOtelTools(nil)...) // Will need proper OTEL client

	agent := NewAgent(Config{
		Name:         "PipelineAgent",
		Description:  "An AI agent specialized for OpenTelemetry collector pipeline configuration",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        allTools,
	})

	return &PipelineAgent{Agent: agent}, nil
}

// ConfigurePipeline configures a collector pipeline from a plan
func (pa *PipelineAgent) ConfigurePipeline(ctx context.Context, pipelineName string, collectorID string, configYAML string, rulesJSON string) *RunResult {
	prompt := fmt.Sprintf(`Configure a collector pipeline.

Pipeline details:
- Name: %s
- Collector ID: %s
- Config YAML: %s
- Processing Rules: %s

Tasks:
1. Review the existing collector configuration for %s
2. Apply the processing rules to the pipeline
3. Configure sampling, filtering, and batching as needed
4. Update collector configuration via OpAMP
5. Verify the configuration is applied
6. Monitor pipeline performance

Start by reading the current collector configuration and then apply updates.`, pipelineName, collectorID, configYAML, rulesJSON, collectorID)

	return pa.Run(ctx, prompt)
}
