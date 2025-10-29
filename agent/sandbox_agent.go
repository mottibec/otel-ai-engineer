package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	sandboxTools "github.com/mottibechhofer/otel-ai-engineer/tools/sandbox"
)

// SandboxAgent is a specialized agent for OpenTelemetry collector sandbox testing
type SandboxAgent struct {
	*Agent
}

// NewSandboxAgent creates a new sandbox testing agent
func NewSandboxAgent(client *anthropic.Client, logLevel config.LogLevel) (*SandboxAgent, error) {
	systemPrompt := `You are an expert OpenTelemetry collector testing and validation assistant. You help users test OpenTelemetry collector configurations in isolated sandbox environments before deploying to production.

Your capabilities:

**Sandbox Management**:
- Create isolated test environments (sandboxes) with OpenTelemetry collectors
- Each sandbox runs in its own Docker network with:
  - An OpenTelemetry Collector instance with custom configuration
  - Telemetrygen containers for generating synthetic telemetry (traces, metrics, logs)
  - Log and metric collection for validation
- List all active sandboxes and their status
- Get detailed information about specific sandboxes

**Telemetry Generation**:
- Generate synthetic traces, metrics, and logs using telemetrygen
- Configure generation rates (traces/metrics/logs per second)
- Set custom duration for telemetry generation
- Auto-validate after generation completes

**Validation & Diagnostics**:
- Validate collector configurations for syntax and structure errors
- Check pipeline configurations (receivers, processors, exporters)
- Verify telemetry is flowing correctly (received vs exported)
- Detect common issues:
  - Missing required configuration sections
  - Improperly configured pipelines
  - Connection failures to backends
  - Queue overflows and backpressure
  - Data loss percentages
- Analyze collector logs for errors and warnings
- Collect internal collector metrics (queue sizes, throughput, resource usage)
- Provide AI-powered recommendations for improvements

**File System Operations**:
- Read, write, and edit collector configuration files
- Create configuration templates
- Search for configuration examples

**Testing Workflow**:

1. **Create a Sandbox**:
   - User provides a collector configuration (YAML)
   - Choose collector version (default: latest)
   - Optionally enable automatic telemetry generation
   - Sandbox is deployed with isolated network

2. **Generate Telemetry**:
   - Start telemetrygen to send synthetic data
   - Configure what types to generate (traces/metrics/logs)
   - Set generation rates and duration
   - Optionally auto-validate after completion

3. **Validate & Diagnose**:
   - Run comprehensive validation checks
   - Collect logs and metrics
   - Identify configuration issues
   - Get AI-powered recommendations
   - Fix issues and re-validate

4. **Iterate**:
   - Update configurations based on findings
   - Re-run validation
   - Compare results

5. **Cleanup**:
   - Stop sandbox when done testing
   - Delete sandbox to free resources

**Best Practices**:

- Always start by creating a sandbox with a clear name describing what you're testing
- Enable telemetry generation appropriate for your configuration (if testing traces pipeline, generate traces)
- Run validation with both logs and metrics collection enabled
- Check validation results for critical and high-severity issues first
- Use AI analysis to get recommendations for fixing issues
- Iterate on configuration until validation passes
- Stop or delete sandboxes when done to free resources

**Common Use Cases**:

1. **Testing New Configurations**:
   - Create sandbox with new config
   - Generate appropriate telemetry
   - Validate and fix issues
   - Deploy to production once validated

2. **Debugging Existing Configurations**:
   - Reproduce issue in sandbox
   - Generate logs and metrics
   - Analyze validation results
   - Test fixes

3. **Learning OpenTelemetry**:
   - Try different receiver/processor/exporter combinations
   - See how data flows through pipelines
   - Experiment with different settings

4. **Performance Testing**:
   - Test with different generation rates
   - Monitor queue sizes and resource usage
   - Identify bottlenecks

**Error Handling**:
- Provide clear explanations when operations fail
- Suggest troubleshooting steps
- Help users interpret validation results
- Guide users to fix configuration issues

Remember: Sandboxes are ephemeral test environments. They're perfect for experimentation and validation before production deployment.`

	// Initialize sandbox tools
	if err := sandboxTools.InitializeSandboxTools(); err != nil {
		return nil, fmt.Errorf("failed to initialize sandbox tools: %w", err)
	}

	// Get tools directly (sandbox tools + file system tools)
	sandboxToolsList := sandboxTools.GetSandboxTools()
	fileSystemTools := tools.GetFileSystemTools()
	allToolsList := append(fileSystemTools, sandboxToolsList...)

	agent := NewAgent(Config{
		Name:         "SandboxAgent",
		Description:  "An AI agent specialized for OpenTelemetry collector testing and validation in sandbox environments",
		Client:       client,
		Model:        anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:    4096,
		SystemPrompt: systemPrompt,
		LogLevel:     logLevel,
		Tools:        allToolsList,
	})

	return &SandboxAgent{
		Agent: agent,
	}, nil
}

// TestConfiguration tests a collector configuration in a sandbox
func (sa *SandboxAgent) TestConfiguration(ctx context.Context, config string, testName string) *RunResult {
	prompt := fmt.Sprintf(`I need to test this OpenTelemetry collector configuration in a sandbox environment.

Configuration:
---
%s
---

Test name: %s

Steps:
1. Create a sandbox with this configuration
2. Start telemetry generation (traces, metrics, and logs) for 30 seconds with auto-validation
3. Review the validation results
4. Provide a summary of any issues found and recommendations
5. Keep the sandbox running so the user can inspect it further

Please execute these steps and report the results.`, config, testName)

	return sa.Run(ctx, prompt)
}

// ValidateConfiguration validates an existing sandbox
func (sa *SandboxAgent) ValidateConfiguration(ctx context.Context, sandboxID string) *RunResult {
	prompt := fmt.Sprintf(`Please validate the sandbox '%s':

1. Get the current sandbox details
2. Get the collector logs (last 100 lines)
3. Get the collector metrics
4. Run validation with AI analysis
5. Provide a detailed summary of findings and recommendations

Focus on critical and high-severity issues first.`, sandboxID)

	return sa.Run(ctx, prompt)
}

// DiagnoseIssue helps diagnose a specific issue in a sandbox
func (sa *SandboxAgent) DiagnoseIssue(ctx context.Context, sandboxID string, issue string) *RunResult {
	prompt := fmt.Sprintf(`I need help diagnosing an issue with sandbox '%s'.

Issue: %s

Please:
1. Get sandbox details and current status
2. Review collector logs for relevant errors
3. Check collector metrics
4. Run validation
5. Analyze the issue and suggest specific fixes
6. If needed, provide an updated configuration that addresses the issue`, sandboxID, issue)

	return sa.Run(ctx, prompt)
}

// CompareConfigurations compares two configurations by testing them in separate sandboxes
func (sa *SandboxAgent) CompareConfigurations(ctx context.Context, config1 string, config2 string, description string) *RunResult {
	prompt := fmt.Sprintf(`I need to compare two OpenTelemetry collector configurations.

Description: %s

Configuration A:
---
%s
---

Configuration B:
---
%s
---

Please:
1. Create sandbox "comparison-a" with Configuration A
2. Create sandbox "comparison-b" with Configuration B
3. Generate the same telemetry to both (30 seconds, auto-validate)
4. Compare validation results side by side
5. Provide a detailed comparison highlighting:
   - Which configuration performs better
   - Any issues specific to each
   - Recommendations on which to use
6. Keep both sandboxes running for further inspection`, description, config1, config2)

	return sa.Run(ctx, prompt)
}
