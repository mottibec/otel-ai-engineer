package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/server/service"
	collectorService "github.com/mottibechhofer/otel-ai-engineer/server/service/collector"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// DelegateRequest represents the request to delegate a task to an agent
type DelegateRequest struct {
	ResourceType    string                 `json:"resource_type"` // "collector", "backend", "service", etc.
	ResourceID      string                 `json:"resource_id"`
	AgentID         string                 `json:"agent_id"`
	TaskDescription string                 `json:"task_description"`
	AgentParams     map[string]interface{} `json:"agent_params,omitempty"` // Optional agent-specific parameters
}

// DelegateResponse represents the response from delegating a task
type DelegateResponse struct {
	WorkID    string `json:"work_id"`
	RunID     string `json:"run_id"`
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// HandleDelegate handles POST /api/resources/:resourceType/:resourceId/delegate
func (s *Server) HandleDelegate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := storage.ResourceType(vars["resourceType"])
	resourceID := vars["resourceId"]

	var req DelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate resource type matches URL
	if req.ResourceType != "" && string(resourceType) != req.ResourceType {
		http.Error(w, "Resource type mismatch", http.StatusBadRequest)
		return
	}
	req.ResourceType = string(resourceType)

	// Validate resource ID matches URL
	if req.ResourceID != "" && resourceID != req.ResourceID {
		http.Error(w, "Resource ID mismatch", http.StatusBadRequest)
		return
	}
	req.ResourceID = resourceID

	if req.AgentID == "" {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}
	if req.TaskDescription == "" {
		http.Error(w, "task_description is required", http.StatusBadRequest)
		return
	}

	// Check if agent exists
	agentInfo, exists := s.agentRegistry.Get(req.AgentID)
	if !exists {
		http.Error(w, fmt.Sprintf("Agent not found: %s", req.AgentID), http.StatusBadRequest)
		return
	}

	// Check for existing active work on this resource
	activeWork, err := s.agentWorkService.GetActiveWorkForResource(r.Context(), resourceType, resourceID)
	if err == nil && activeWork != nil {
		http.Error(w, fmt.Sprintf("Agent '%s' is already working on this resource (work_id: %s, run_id: %s). Please cancel existing work first.", activeWork.AgentName, activeWork.ID, activeWork.RunID), http.StatusConflict)
		return
	}

	// Generate work ID
	workID := fmt.Sprintf("work-%d", time.Now().UnixNano())

	// Build task prompt with resource context
	taskPrompt := s.buildTaskPrompt(req.ResourceType, resourceID, req.TaskDescription, req.AgentParams)

	// Create and start agent run
	runResponse, err := s.runService.CreateRun(r.Context(), service.CreateRunRequest{
		AgentID: req.AgentID,
		Prompt:  taskPrompt,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start agent: %v", err), http.StatusInternalServerError)
		return
	}

	// Use the actual run ID from the response
	actualRunID := runResponse.RunID
	if actualRunID == "" && runResponse.Run != nil {
		actualRunID = runResponse.Run.ID
	}

	// Create agent work entry
	work := &storage.AgentWork{
		ID:              workID,
		ResourceType:    resourceType,
		ResourceID:      resourceID,
		RunID:           actualRunID,
		AgentID:         req.AgentID,
		AgentName:       agentInfo.Name,
		TaskDescription: req.TaskDescription,
		Status:          storage.AgentWorkStatusRunning,
		StartedAt:       time.Now(),
	}

	if err := s.agentWorkService.CreateAgentWork(r.Context(), work); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create agent work entry: %v", err), http.StatusInternalServerError)
		return
	}

	// Monitor run completion to update agent work status
	go s.monitorRunCompletion(actualRunID, workID)

	response := DelegateResponse{
		WorkID:    workID,
		RunID:     actualRunID,
		AgentID:   req.AgentID,
		AgentName: agentInfo.Name,
		Status:    "running",
		Message:   fmt.Sprintf("Agent '%s' has been assigned to work on %s '%s'", agentInfo.Name, req.ResourceType, resourceID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// buildTaskPrompt creates a contextual prompt for the agent based on the resource
func (s *Server) buildTaskPrompt(resourceType string, resourceID string, taskDescription string, agentParams map[string]interface{}) string {
	prompt := fmt.Sprintf("You are working on a %s resource with ID: %s.\n\n", resourceType, resourceID)

	// Add resource-specific context
	switch storage.ResourceType(resourceType) {
	case storage.ResourceTypeCollector:
		// Try to get collector info
		collector, err := s.getCollectorInfo(resourceID)
		if err == nil && collector != nil {
			prompt += fmt.Sprintf("Collector Details:\n- Name: %s\n- Status: %s\n- Target Type: %s\n\n",
				collector.CollectorName, collector.Status, collector.TargetType)
		}

	case storage.ResourceTypeBackend:
		// Try to get backend info
		backend, err := s.backendService.GetBackend(context.Background(), resourceID, false)
		if err == nil && backend != nil && backend.Backend != nil {
			prompt += fmt.Sprintf("Backend Details:\n- Name: %s\n- Type: %s\n- URL: %s\n- Health: %s\n\n",
				backend.Backend.Name, backend.Backend.BackendType, backend.Backend.URL, backend.Backend.HealthStatus)
		}

	case storage.ResourceTypeService:
		// Try to get service info from plans
		prompt += "This is a service that needs instrumentation.\n\n"

	case storage.ResourceTypeInfrastructure:
		prompt += "This is an infrastructure component that needs monitoring.\n\n"

	case storage.ResourceTypePipeline:
		prompt += "This is a collector pipeline that needs configuration.\n\n"
	}

	prompt += fmt.Sprintf("Task: %s\n\n", taskDescription)

	prompt += "Use the available tools to complete this task. Report your progress and any issues encountered."

	return prompt
}

// getCollectorInfo gets collector information (helper for buildTaskPrompt)
func (s *Server) getCollectorInfo(collectorID string) (*collectorService.CollectorResponse, error) {
	// Use the collector service to get collector info
	collector, err := s.collectorService.GetCollector(context.Background(), collectorID, false)
	if err != nil {
		return nil, err
	}
	return collector, nil
}

// monitorRunCompletion monitors a run and updates agent work status when it completes
func (s *Server) monitorRunCompletion(runID string, workID string) {
	// Poll run status periodically
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	maxPolls := 300 // 10 minutes max
	pollCount := 0

	for range ticker.C {
		pollCount++
		if pollCount > maxPolls {
			// Timeout - mark as failed
			status := storage.AgentWorkStatusFailed
			errorMsg := "Monitoring timeout"
			s.agentWorkService.UpdateAgentWork(context.Background(), workID, &storage.AgentWorkUpdate{
				Status: &status,
				Error:  &errorMsg,
			})
			return
		}

		run, err := s.storage.GetRun(runID)
		if err != nil {
			continue
		}

		// Check if run is complete
		if run.Status == storage.RunStatusSuccess || run.Status == storage.RunStatusFailed || run.Status == storage.RunStatusCancelled {
			var workStatus storage.AgentWorkStatus
			var errorMsg string

			switch run.Status {
			case storage.RunStatusSuccess:
				workStatus = storage.AgentWorkStatusCompleted
			case storage.RunStatusFailed:
				workStatus = storage.AgentWorkStatusFailed
				errorMsg = run.Error
			case storage.RunStatusCancelled:
				workStatus = storage.AgentWorkStatusCancelled
			}

			update := &storage.AgentWorkUpdate{
				Status: &workStatus,
			}
			if errorMsg != "" {
				update.Error = &errorMsg
			}

			s.agentWorkService.UpdateAgentWork(context.Background(), workID, update)
			return
		}
	}
}
