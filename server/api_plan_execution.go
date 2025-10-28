package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/agent"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// HandleExecutePlan handles POST /api/plans/:planId/execute
func (s *Server) HandleExecutePlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	// Get the plan from storage
	plan, err := s.storage.GetPlan(planID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Plan not found: %v", err), http.StatusNotFound)
		return
	}

	// Update plan status to executing
	err = s.storage.UpdatePlan(planID, &storage.PlanUpdate{
		Status: func() *storage.PlanStatus {
			status := storage.PlanStatusExecuting
			return &status
		}(),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update plan status: %v", err), http.StatusInternalServerError)
		return
	}

	// Create execution context
	ctx := r.Context()

	// Create a run with the observability agent to execute the plan
	runID := fmt.Sprintf("plan-exec-%s-%d", planID, time.Now().UnixNano())

	// Start the plan execution asynchronously
	go func() {
		s.executePlanAsync(ctx, plan, runID)
	}()

	// Return the run ID immediately
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"run_id":  runID,
		"plan_id": planID,
		"status":  "executing",
		"message": "Plan execution started. Check run details for progress.",
	})
}

// executePlanAsync executes the plan in the background
func (s *Server) executePlanAsync(ctx context.Context, plan *storage.ObservabilityPlan, runID string) {
	// Create ObservabilityAgent using the eventBridge's emitter
	obsAgent, err := s.agentRegistry.Create("observability", s.anthropicClient, s.logLevel, s.eventBridge.emitter)
	if err != nil {
		log.Printf("Failed to create observability agent: %v", err)
		return
	}

	// Register handoff tool to enable agent-to-agent delegation
	handoffCtx := &agent.HandoffContext{
		ParentRunID:   runID,
		ParentAgentID: "observability",
		Registry:      s.agentRegistry,
		Client:        s.anthropicClient,
		LogLevel:      s.logLevel,
		EventEmitter:  s.eventBridge.emitter,
		Storage:       s.storage,
	}
	handoffTool := agent.CreateHandoffTool(handoffCtx)
	obsAgent.RegisterTool("handoff_task", "Delegate a task to another specialized agent", handoffTool.Schema, handoffTool.Handler)

	// Create a prompt that includes the plan details
	prompt := s.createPlanExecutionPrompt(plan)

	// Run the agent
	result := obsAgent.Run(ctx, prompt)

	// Update plan status based on result
	status := storage.PlanStatusFailed
	if result.Success {
		status = storage.PlanStatusSuccess
	}

	s.storage.UpdatePlan(plan.ID, &storage.PlanUpdate{
		Status: &status,
	})
}

// createPlanExecutionPrompt creates a detailed prompt for executing the plan
func (s *Server) createPlanExecutionPrompt(plan *storage.ObservabilityPlan) string {
	prompt := fmt.Sprintf(`Execute the observability plan "%s" (ID: %s).

Plan Details:
- Environment: %s
- Status: %s
- Created: %s

Components to deploy:

`, plan.Name, plan.ID, plan.Environment, plan.Status, plan.CreatedAt)

	// Add services to instrument
	if len(plan.Services) > 0 {
		prompt += "Services to instrument:\n"
		for _, service := range plan.Services {
			prompt += fmt.Sprintf("- %s (Language: %s, Framework: %s, Path: %s)\n",
				service.ServiceName, service.Language, service.Framework, service.TargetPath)
		}
		prompt += "\n"
	}

	// Add infrastructure components
	if len(plan.Infrastructure) > 0 {
		prompt += "Infrastructure to monitor:\n"
		for _, infra := range plan.Infrastructure {
			prompt += fmt.Sprintf("- %s (Type: %s, Host: %s, Receiver: %s)\n",
				infra.Name, infra.ComponentType, infra.Host, infra.ReceiverType)
		}
		prompt += "\n"
	}

	// Add pipelines
	if len(plan.Pipelines) > 0 {
		prompt += "Collector pipelines to configure:\n"
		for _, pipeline := range plan.Pipelines {
			prompt += fmt.Sprintf("- %s (Collector: %s, Target: %s)\n",
				pipeline.Name, pipeline.CollectorID, pipeline.TargetType)
		}
		prompt += "\n"
	}

	// Add backends
	if len(plan.Backends) > 0 {
		prompt += "Backends to connect:\n"
		for _, backend := range plan.Backends {
			prompt += fmt.Sprintf("- %s (Type: %s, URL: %s)\n",
				backend.Name, backend.BackendType, backend.URL)
		}
		prompt += "\n"
	}

	prompt += `Execution Strategy:
For each component in the plan, use the handoff_task tool to delegate to the appropriate specialized agent:
- For services: Call handoff_task with to_agent_id='instrumentation' and describe the instrumentation task
- For infrastructure: Call handoff_task with to_agent_id='infrastructure' and describe the monitoring setup
- For pipelines: Call handoff_task with to_agent_id='pipeline' and describe the pipeline configuration
- For backends: Call handoff_task with to_agent_id='backend' and describe the backend connection

Example for service instrumentation:
{
  "to_agent_id": "instrumentation",
  "task_description": "Instrument service 'user-service' located at '/path/to/service'. Language: Go, Framework: Gin. Install OpenTelemetry SDK, configure OTLP exporter, and add instrumentation to HTTP handlers."
}

Each handoff will create a sub-run that you can track. Report progress and update component statuses as you proceed.`

	return prompt
}
