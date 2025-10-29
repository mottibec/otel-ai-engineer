package sandbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Validator validates sandbox configurations and operations
type Validator struct {
	orchestrator *DockerOrchestrator
	logger       Logger
}

// NewValidator creates a new validator
func NewValidator(orchestrator *DockerOrchestrator, logger Logger) *Validator {
	return &Validator{
		orchestrator: orchestrator,
		logger:       logger,
	}
}

// Validate runs validation checks on a sandbox
func (v *Validator) Validate(ctx context.Context, sandbox *Sandbox, req ValidateSandboxRequest) (*ValidationResult, error) {
	startedAt := time.Now()

	result := &ValidationResult{
		ID:        fmt.Sprintf("validation-%d", time.Now().Unix()),
		SandboxID: sandbox.ID,
		Status:    ValidationStatusRunning,
		StartedAt: startedAt,
		Checks:    []ValidationCheck{},
		Summary:   ValidationSummary{},
		Issues:    []ValidationIssue{},
	}

	// Run all validation checks
	v.checkContainerHealth(ctx, sandbox, result)
	v.checkCollectorConfiguration(ctx, sandbox, result)
	v.checkPipelineConfiguration(ctx, sandbox, result)
	v.checkTelemetryFlow(ctx, sandbox, result)

	// Collect logs if requested
	if req.CollectLogs {
		logs, err := v.orchestrator.GetContainerLogs(ctx, sandbox.CollectorContainerID, 100)
		if err != nil {
			v.logger.Error("Failed to collect logs", err, nil)
		} else {
			result.CollectorLogs = logs
			v.analyzeLogs(logs, result)
		}
	}

	// Collect metrics if requested
	if req.CollectMetrics {
		metrics, err := v.orchestrator.GetCollectorMetrics(ctx, sandbox.CollectorContainerID)
		if err != nil {
			v.logger.Error("Failed to collect metrics", err, nil)
		} else {
			result.CollectorMetrics = *metrics
			v.analyzeMetrics(metrics, result)
		}
	}

	// Calculate summary
	v.calculateSummary(result)

	// Determine overall status
	if result.Summary.Critical > 0 {
		result.Status = ValidationStatusFailed
	} else if result.Summary.Failed > 0 {
		result.Status = ValidationStatusPartial
	} else {
		result.Status = ValidationStatusPassed
	}

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startedAt)

	// Run AI analysis if requested
	if req.AIAnalysis {
		v.runAIAnalysis(result)
	}

	return result, nil
}

// checkContainerHealth verifies the collector container is running
func (v *Validator) checkContainerHealth(ctx context.Context, sandbox *Sandbox, result *ValidationResult) {
	check := ValidationCheck{
		Name:      "Container Health",
		Category:  "infrastructure",
		Timestamp: time.Now(),
	}

	status, err := v.orchestrator.getContainerStatus(ctx, sandbox.CollectorContainerID)
	if err != nil {
		check.Status = "failed"
		check.Severity = "critical"
		check.Message = "Failed to get container status"
		check.Details = err.Error()

		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "infrastructure",
			Severity:    "critical",
			Component:   "collector-container",
			Message:     "Collector container is not accessible",
			Description: err.Error(),
			Suggestion:  "Check if the container is running and accessible",
			Timestamp:   time.Now(),
		})
	} else if status != "running" {
		check.Status = "failed"
		check.Severity = "critical"
		check.Message = fmt.Sprintf("Container is not running (status: %s)", status)

		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "infrastructure",
			Severity:    "critical",
			Component:   "collector-container",
			Message:     "Collector container is not running",
			Description: fmt.Sprintf("Container status: %s", status),
			Suggestion:  "Check collector logs for startup errors",
			Timestamp:   time.Now(),
		})
	} else {
		check.Status = "passed"
		check.Severity = "info"
		check.Message = "Container is running"
	}

	result.Checks = append(result.Checks, check)
}

// checkCollectorConfiguration validates the collector YAML configuration
func (v *Validator) checkCollectorConfiguration(ctx context.Context, sandbox *Sandbox, result *ValidationResult) {
	check := ValidationCheck{
		Name:      "Configuration Validity",
		Category:  "configuration",
		Timestamp: time.Now(),
	}

	// Parse YAML
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(sandbox.CollectorConfig), &config); err != nil {
		check.Status = "failed"
		check.Severity = "critical"
		check.Message = "Invalid YAML configuration"
		check.Details = err.Error()

		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "configuration",
			Severity:    "critical",
			Component:   "config-yaml",
			Message:     "Configuration YAML is invalid",
			Description: err.Error(),
			Suggestion:  "Fix YAML syntax errors",
			Timestamp:   time.Now(),
		})

		result.Checks = append(result.Checks, check)
		return
	}

	// Check for required sections
	requiredSections := []string{"receivers", "processors", "exporters", "service"}
	missingSections := []string{}

	for _, section := range requiredSections {
		if _, exists := config[section]; !exists {
			missingSections = append(missingSections, section)
		}
	}

	if len(missingSections) > 0 {
		check.Status = "failed"
		check.Severity = "high"
		check.Message = fmt.Sprintf("Missing required sections: %s", strings.Join(missingSections, ", "))

		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "configuration",
			Severity:    "high",
			Component:   "config-structure",
			Message:     "Configuration is missing required sections",
			Description: fmt.Sprintf("Missing: %s", strings.Join(missingSections, ", ")),
			Suggestion:  "Add the required configuration sections",
			Timestamp:   time.Now(),
		})
	} else {
		check.Status = "passed"
		check.Severity = "info"
		check.Message = "Configuration structure is valid"
	}

	result.Checks = append(result.Checks, check)
}

// checkPipelineConfiguration validates the pipeline setup
func (v *Validator) checkPipelineConfiguration(ctx context.Context, sandbox *Sandbox, result *ValidationResult) {
	check := ValidationCheck{
		Name:      "Pipeline Configuration",
		Category:  "pipeline",
		Timestamp: time.Now(),
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(sandbox.CollectorConfig), &config); err != nil {
		// Already caught in previous check
		return
	}

	// Check service pipelines
	service, ok := config["service"].(map[interface{}]interface{})
	if !ok {
		check.Status = "warning"
		check.Severity = "medium"
		check.Message = "Service configuration not found"
		result.Checks = append(result.Checks, check)
		return
	}

	pipelines, ok := service["pipelines"].(map[interface{}]interface{})
	if !ok || len(pipelines) == 0 {
		check.Status = "failed"
		check.Severity = "high"
		check.Message = "No pipelines configured"

		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "pipeline",
			Severity:    "high",
			Component:   "service-pipelines",
			Message:     "No telemetry pipelines are configured",
			Description: "The service.pipelines section is empty or missing",
			Suggestion:  "Configure at least one pipeline (traces, metrics, or logs)",
			Timestamp:   time.Now(),
		})
	} else {
		// Check each pipeline has receivers, processors, and exporters
		issuesFound := false
		for pipelineName, pipelineConfig := range pipelines {
			pipeline, ok := pipelineConfig.(map[interface{}]interface{})
			if !ok {
				continue
			}

			pipelineNameStr := fmt.Sprintf("%v", pipelineName)

			if _, hasReceivers := pipeline["receivers"]; !hasReceivers {
				issuesFound = true
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        "pipeline",
					Severity:    "high",
					Component:   pipelineNameStr,
					Message:     fmt.Sprintf("Pipeline '%s' has no receivers", pipelineNameStr),
					Suggestion:  "Add at least one receiver to the pipeline",
					Timestamp:   time.Now(),
				})
			}

			if _, hasExporters := pipeline["exporters"]; !hasExporters {
				issuesFound = true
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        "pipeline",
					Severity:    "high",
					Component:   pipelineNameStr,
					Message:     fmt.Sprintf("Pipeline '%s' has no exporters", pipelineNameStr),
					Suggestion:  "Add at least one exporter to the pipeline",
					Timestamp:   time.Now(),
				})
			}
		}

		if issuesFound {
			check.Status = "failed"
			check.Severity = "high"
			check.Message = "Pipeline configuration has issues"
		} else {
			check.Status = "passed"
			check.Severity = "info"
			check.Message = fmt.Sprintf("%d pipeline(s) configured correctly", len(pipelines))
		}
	}

	result.Checks = append(result.Checks, check)
}

// checkTelemetryFlow validates telemetry is flowing correctly
func (v *Validator) checkTelemetryFlow(ctx context.Context, sandbox *Sandbox, result *ValidationResult) {
	check := ValidationCheck{
		Name:      "Telemetry Flow",
		Category:  "telemetry",
		Timestamp: time.Now(),
	}

	// Get metrics to check telemetry flow
	metrics, err := v.orchestrator.GetCollectorMetrics(ctx, sandbox.CollectorContainerID)
	if err != nil {
		check.Status = "warning"
		check.Severity = "medium"
		check.Message = "Unable to check telemetry flow"
		check.Details = "Metrics endpoint not accessible"
		result.Checks = append(result.Checks, check)
		return
	}

	// Check if any telemetry was received
	totalReceived := metrics.ReceiverAcceptedSpans + metrics.ReceiverAcceptedMetrics + metrics.ReceiverAcceptedLogs
	totalExported := metrics.ExporterSentSpans + metrics.ExporterSentMetrics + metrics.ExporterSentLogs

	if totalReceived == 0 {
		check.Status = "warning"
		check.Severity = "medium"
		check.Message = "No telemetry data received"
		check.Details = "Collector has not received any traces, metrics, or logs"

		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "telemetry",
			Severity:    "medium",
			Component:   "receiver",
			Message:     "No telemetry data has been received",
			Description: "The collector is running but hasn't received any data",
			Suggestion:  "Check if telemetry generators are running and sending to the correct endpoint",
			Timestamp:   time.Now(),
		})
	} else if totalExported == 0 {
		check.Status = "failed"
		check.Severity = "high"
		check.Message = "Telemetry received but not exported"
		check.Details = fmt.Sprintf("Received: %d, Exported: 0", totalReceived)

		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "telemetry",
			Severity:    "high",
			Component:   "exporter",
			Message:     "Telemetry data is not being exported",
			Description: fmt.Sprintf("Received %d data points but exported 0", totalReceived),
			Suggestion:  "Check exporter configuration and connectivity to backend",
			Timestamp:   time.Now(),
		})
	} else {
		dataLoss := float64(totalReceived-totalExported) / float64(totalReceived) * 100
		if dataLoss > 5 {
			check.Status = "warning"
			check.Severity = "medium"
			check.Message = fmt.Sprintf("Data loss detected: %.1f%%", dataLoss)

			result.Issues = append(result.Issues, ValidationIssue{
				Type:        "telemetry",
				Severity:    "medium",
				Component:   "pipeline",
				Message:     fmt.Sprintf("%.1f%% data loss detected", dataLoss),
				Description: fmt.Sprintf("Received: %d, Exported: %d", totalReceived, totalExported),
				Suggestion:  "Check for dropped data in processors or export queue overflows",
				Timestamp:   time.Now(),
			})
		} else {
			check.Status = "passed"
			check.Severity = "info"
			check.Message = fmt.Sprintf("Telemetry flowing correctly (received: %d, exported: %d)", totalReceived, totalExported)
		}
	}

	result.Checks = append(result.Checks, check)
}

// analyzeLogs analyzes collector logs for common issues
func (v *Validator) analyzeLogs(logs []LogEntry, result *ValidationResult) {
	errorCount := 0
	warnCount := 0

	for _, log := range logs {
		if log.Level == "error" {
			errorCount++

			// Detect common error patterns
			msg := strings.ToLower(log.Message)
			if strings.Contains(msg, "connection refused") {
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        "exporter",
					Severity:    "high",
					Component:   "exporter-connection",
					Message:     "Exporter connection refused",
					Description: log.Message,
					Suggestion:  "Check if the backend is accessible and the endpoint is correct",
					Timestamp:   log.Timestamp,
				})
			} else if strings.Contains(msg, "timeout") {
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        "exporter",
					Severity:    "medium",
					Component:   "exporter-timeout",
					Message:     "Exporter timeout detected",
					Description: log.Message,
					Suggestion:  "Increase timeout or check backend responsiveness",
					Timestamp:   log.Timestamp,
				})
			} else if strings.Contains(msg, "queue is full") {
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        "processor",
					Severity:    "high",
					Component:   "batch-processor",
					Message:     "Queue overflow detected",
					Description: log.Message,
					Suggestion:  "Increase queue size or adjust batch processor settings",
					Timestamp:   log.Timestamp,
				})
			}
		} else if log.Level == "warn" {
			warnCount++
		}
	}

	if errorCount > 0 || warnCount > 0 {
		result.Checks = append(result.Checks, ValidationCheck{
			Name:      "Log Analysis",
			Category:  "logs",
			Status:    "warning",
			Severity:  "medium",
			Message:   fmt.Sprintf("Found %d errors and %d warnings in logs", errorCount, warnCount),
			Timestamp: time.Now(),
		})
	}
}

// analyzeMetrics analyzes collector metrics
func (v *Validator) analyzeMetrics(metrics *CollectorMetrics, result *ValidationResult) {
	result.Summary.TracesReceived = metrics.ReceiverAcceptedSpans
	result.Summary.TracesExported = metrics.ExporterSentSpans
	result.Summary.MetricsReceived = metrics.ReceiverAcceptedMetrics
	result.Summary.MetricsExported = metrics.ExporterSentMetrics
	result.Summary.LogsReceived = metrics.ReceiverAcceptedLogs
	result.Summary.LogsExported = metrics.ExporterSentLogs

	// Calculate data loss
	totalReceived := float64(metrics.ReceiverAcceptedSpans + metrics.ReceiverAcceptedMetrics + metrics.ReceiverAcceptedLogs)
	totalExported := float64(metrics.ExporterSentSpans + metrics.ExporterSentMetrics + metrics.ExporterSentLogs)

	if totalReceived > 0 {
		result.Summary.DataLossPercent = (totalReceived - totalExported) / totalReceived * 100
	}

	// Check for resource issues
	if metrics.MemoryUsageMB > 500 {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "resource",
			Severity:    "medium",
			Component:   "memory",
			Message:     "High memory usage detected",
			Description: fmt.Sprintf("Memory usage: %.1f MB", metrics.MemoryUsageMB),
			Suggestion:  "Consider increasing memory limits or optimizing pipeline",
			Timestamp:   time.Now(),
		})
	}

	// Check queue capacity
	if metrics.QueueCapacity > 0 && metrics.QueueSize > int64(float64(metrics.QueueCapacity)*0.8) {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "processor",
			Severity:    "high",
			Component:   "queue",
			Message:     "Queue near capacity",
			Description: fmt.Sprintf("Queue: %d/%d (%.0f%% full)", metrics.QueueSize, metrics.QueueCapacity, float64(metrics.QueueSize)/float64(metrics.QueueCapacity)*100),
			Suggestion:  "Increase queue size or improve export throughput",
			Timestamp:   time.Now(),
		})
	}
}

// calculateSummary calculates the validation summary
func (v *Validator) calculateSummary(result *ValidationResult) {
	result.Summary.TotalChecks = len(result.Checks)

	for _, check := range result.Checks {
		switch check.Status {
		case "passed":
			result.Summary.Passed++
		case "failed":
			result.Summary.Failed++
		case "warning":
			result.Summary.Warnings++
		}
	}

	for _, issue := range result.Issues {
		if issue.Severity == "critical" {
			result.Summary.Critical++
		}
	}
}

// runAIAnalysis generates AI-powered recommendations
func (v *Validator) runAIAnalysis(result *ValidationResult) {
	// This would integrate with Claude API to analyze the validation results
	// For now, we'll generate basic recommendations based on issues

	if result.Summary.Failed > 0 || result.Summary.Critical > 0 {
		result.AIAnalysis = "The collector has critical issues that need immediate attention. "
	} else if result.Summary.Warnings > 0 {
		result.AIAnalysis = "The collector is operational but has some warnings. "
	} else {
		result.AIAnalysis = "The collector configuration looks good! "
	}

	// Generate recommendations based on issues
	for _, issue := range result.Issues {
		if issue.Severity == "critical" || issue.Severity == "high" {
			result.Recommendations = append(result.Recommendations, issue.Suggestion)
		}
	}

	// Add data quality recommendations
	if result.Summary.DataLossPercent > 5 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Address %.1f%% data loss by checking processor and exporter configurations", result.Summary.DataLossPercent))
	}
}
