package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

// InstrumentationAgent is specialized for service instrumentation
type InstrumentationAgent struct {
	*Agent
}

// NewInstrumentationAgent creates a new instrumentation agent with filesystem tools
func NewInstrumentationAgent(client *anthropic.Client, logLevel config.LogLevel) (*InstrumentationAgent, error) {
	systemPrompt := `You are an expert service instrumentation assistant specializing in OpenTelemetry SDK integration.

Your capabilities:
- Analyze codebases to detect programming language and framework
- Detect existing instrumentation patterns
- Install appropriate OpenTelemetry SDKs
- Configure exporters for observability data
- Generate instrumentation code
- Validate instrumentation setup

- **Task Delegation**:
  - Use the 'handoff_task' tool when a task is better suited for another agent
  - Delegate infrastructure monitoring setup to 'infrastructure' agent
  - Delegate backend connectivity to 'backend' agent
  - Delegate collector deployment to 'otel' agent
  - The handoff is blocking - wait for completion before proceeding

Instrumentation process:
1. **Codebase Analysis**:
   - Detect programming language (Go, Python, Java, Node.js, etc.)
   - Identify framework (Django, Flask, Express, Gin, etc.)
   - Scan for existing OpenTelemetry instrumentation

2. **SDK Installation**:
   - Add OpenTelemetry SDK dependencies
   - Configure package managers (npm, pip, go.mod, etc.)
   - Install required packages

3. **Configuration**:
   - Set up OpenTelemetry auto-instrumentation where available
   - Configure exporters (OTLP, HTTP, console)
   - Set service name, resource attributes
   - Configure sampling and context propagation

4. **Manual Instrumentation**:
   - Add custom spans for critical business logic
   - Instrument database queries, HTTP calls, message queues
   - Add custom metrics and logs

5. **Validation**:
   - Verify instrumentation compiles
   - Check for common configuration issues
   - Validate exporter connectivity

Best Practices:
- Prefer auto-instrumentation when available
- Add meaningful span names and attributes
- Use semantic conventions
- Keep instrumentation overhead low
- Document changes clearly`

	agent := NewAgent(Config{
		Name:         "InstrumentationAgent",
		Description:  "An AI agent specialized for instrumenting services with OpenTelemetry",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        tools.GetFileSystemTools(),
	})

	return &InstrumentationAgent{Agent: agent}, nil
}

// InstrumentService instruments a service from an observability plan
func (ia *InstrumentationAgent) InstrumentService(ctx context.Context, serviceName string, targetPath string, language string, framework string) *RunResult {
	prompt := fmt.Sprintf(`Instrument a service for observability.

Service details:
- Name: %s
- Target path: %s
- Language: %s
- Framework: %s

Tasks:
1. Read and analyze the codebase at %s
2. Detect existing instrumentation patterns
3. Install appropriate OpenTelemetry SDK
4. Configure auto-instrumentation
5. Add manual instrumentation for critical operations
6. Configure OTLP exporter
7. Verify the instrumentation setup

Start by analyzing the codebase structure and existing dependencies.`, serviceName, targetPath, language, framework, targetPath)

	return ia.Run(ctx, prompt)
}
