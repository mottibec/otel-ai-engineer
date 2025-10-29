package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// HandleListAgentWork handles GET /api/agent-work
func (s *Server) HandleListAgentWork(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	resourceTypeStr := r.URL.Query().Get("resource_type")
	resourceIDStr := r.URL.Query().Get("resource_id")
	statusStr := r.URL.Query().Get("status")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	opts := storage.AgentWorkListOptions{
		Limit:  limit,
		Offset: offset,
	}

	if resourceTypeStr != "" {
		rt := storage.ResourceType(resourceTypeStr)
		opts.ResourceType = &rt
	}

	if resourceIDStr != "" {
		opts.ResourceID = &resourceIDStr
	}

	if statusStr != "" {
		status := storage.AgentWorkStatus(statusStr)
		opts.Status = &status
	}

	works, err := s.agentWorkService.ListAgentWork(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(works)
}

// HandleGetAgentWork handles GET /api/agent-work/:workId
func (s *Server) HandleGetAgentWork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workID := vars["workId"]

	work, err := s.agentWorkService.GetAgentWork(r.Context(), workID)
	if err != nil {
		if err.Error() == "work ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(work)
}

// CreateAgentWorkRequest represents the request body for creating agent work
type CreateAgentWorkRequest struct {
	ResourceType   string `json:"resource_type"`
	ResourceID     string `json:"resource_id"`
	RunID          string `json:"run_id"`
	AgentID        string `json:"agent_id"`
	AgentName      string `json:"agent_name"`
	TaskDescription string `json:"task_description"`
}

// HandleCreateAgentWork handles POST /api/agent-work
func (s *Server) HandleCreateAgentWork(w http.ResponseWriter, r *http.Request) {
	var req CreateAgentWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	work := &storage.AgentWork{
		ResourceType:   storage.ResourceType(req.ResourceType),
		ResourceID:     req.ResourceID,
		RunID:          req.RunID,
		AgentID:        req.AgentID,
		AgentName:      req.AgentName,
		TaskDescription: req.TaskDescription,
		Status:         storage.AgentWorkStatusRunning,
	}

	if err := s.agentWorkService.CreateAgentWork(r.Context(), work); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(work)
}

// UpdateAgentWorkRequest represents the request body for updating agent work
type UpdateAgentWorkRequest struct {
	Status      *string `json:"status,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	Error       *string `json:"error,omitempty"`
}

// HandleUpdateAgentWork handles PUT /api/agent-work/:workId
func (s *Server) HandleUpdateAgentWork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workID := vars["workId"]

	var req UpdateAgentWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	update := &storage.AgentWorkUpdate{}
	if req.Status != nil {
		status := storage.AgentWorkStatus(*req.Status)
		update.Status = &status
	}
	if req.Error != nil {
		update.Error = req.Error
	}

	if err := s.agentWorkService.UpdateAgentWork(r.Context(), workID, update); err != nil {
		if err.Error() == "work ID cannot be empty" || err.Error() == "update cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	work, err := s.agentWorkService.GetAgentWork(r.Context(), workID)
	if err != nil {
		if err.Error() == "work ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(work)
}

// HandleDeleteAgentWork handles DELETE /api/agent-work/:workId
func (s *Server) HandleDeleteAgentWork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workID := vars["workId"]

	if err := s.agentWorkService.DeleteAgentWork(r.Context(), workID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "deleted",
		"work_id": workID,
	})
}

// HandleGetAgentWorkByResource handles GET /api/agent-work/resource/:resourceType/:resourceId
func (s *Server) HandleGetAgentWorkByResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := storage.ResourceType(vars["resourceType"])
	resourceID := vars["resourceId"]

	works, err := s.agentWorkService.GetAgentWorkByResource(r.Context(), resourceType, resourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(works)
}

// HandleCancelAgentWork handles POST /api/agent-work/:workId/cancel
func (s *Server) HandleCancelAgentWork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workID := vars["workId"]

	// Get the agent work entry
	work, err := s.agentWorkService.GetAgentWork(r.Context(), workID)
	if err != nil {
		if err.Error() == "work ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if work is already completed or cancelled
	if work.Status != storage.AgentWorkStatusRunning {
		http.Error(w, "Agent work is not running", http.StatusBadRequest)
		return
	}

	// Stop the associated run if it's still running
	if work.RunID != "" {
		if err := s.runService.StopRun(r.Context(), work.RunID); err != nil {
			// Log error but continue to update agent work status
			// The run might already be stopped
			_ = err
		}
	}

	// Update agent work status to cancelled
	status := storage.AgentWorkStatusCancelled
	update := &storage.AgentWorkUpdate{
		Status: &status,
	}

	if err := s.agentWorkService.UpdateAgentWork(r.Context(), workID, update); err != nil {
		if err.Error() == "work ID cannot be empty" || err.Error() == "update cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated work
	updatedWork, err := s.agentWorkService.GetAgentWork(r.Context(), workID)
	if err != nil {
		if err.Error() == "work ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedWork)
}

