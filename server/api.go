package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/server/service"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
	"github.com/mottibechhofer/otel-ai-engineer/server/validation"
)

// HandleListRuns handles GET /api/runs
func (s *Server) HandleListRuns(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
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

	// Validate parameters
	if err := validation.ValidateListLimit(limit); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidateListOffset(offset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := storage.RunListOptions{
		Limit:  limit,
		Offset: offset,
	}

	if statusStr != "" {
		status := storage.RunStatus(statusStr)
		opts.Status = &status
	}

	runs, err := s.storage.ListRuns(opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

// HandleGetRun handles GET /api/runs/:runId
func (s *Server) HandleGetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	// Validate runID
	if err := validation.ValidateRunID("run_id", runID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	run, err := s.storage.GetRun(runID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

// HandleGetEvents handles GET /api/runs/:runId/events
func (s *Server) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	// Parse query parameters
	afterStr := r.URL.Query().Get("after")
	var after *time.Time
	if afterStr != "" {
		if t, err := time.Parse(time.RFC3339, afterStr); err == nil {
			after = &t
		}
	}

	events, err := s.storage.GetEvents(runID, after)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// HandleGetEventCount handles GET /api/runs/:runId/events/count
func (s *Server) HandleGetEventCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	count, err := s.storage.GetEventCount(runID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// HandleGetTrace handles GET /api/runs/:runId/trace
func (s *Server) HandleGetTrace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	// Compute trace from events
	trace, err := s.traceService.ComputeTrace(runID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trace)
}

// HandleHealth handles GET /api/health
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// HandleListAgents is now in api_agents.go

// CreateRunRequest represents the request body for creating a new run
type CreateRunRequest struct {
	AgentID         string `json:"agent_id"`
	Prompt          string `json:"prompt"`
	ResumeFromRunID string `json:"resume_from_run_id,omitempty"`
}

// AddInstructionRequest represents a request to add a custom instruction
type AddInstructionRequest struct {
	Instruction string `json:"instruction"`
}

// HandleCreateRun handles POST /api/runs
func (s *Server) HandleCreateRun(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req CreateRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize inputs
	req.AgentID = validation.SanitizeString(req.AgentID)
	req.Prompt = validation.SanitizeString(req.Prompt)
	req.ResumeFromRunID = validation.SanitizeString(req.ResumeFromRunID)

	// Validate input using validation package
	validationReq := validation.CreateRunRequest{
		AgentID:         req.AgentID,
		Prompt:          req.Prompt,
		ResumeFromRunID: req.ResumeFromRunID,
	}
	if err := validationReq.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Delegate to service layer
	response, err := s.runService.CreateRun(r.Context(), service.CreateRunRequest{
		AgentID:         req.AgentID,
		Prompt:          req.Prompt,
		ResumeFromRunID: req.ResumeFromRunID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.Run)
}

// HandleStopRun handles POST /api/runs/{runId}/stop
func (s *Server) HandleStopRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	// Delegate to service layer
	if err := s.runService.StopRun(r.Context(), runID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "stopped",
		"run_id": runID,
	})
}

// HandlePauseRun handles POST /api/runs/{runId}/pause
func (s *Server) HandlePauseRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	// Delegate to service layer
	if err := s.runService.PauseRun(r.Context(), runID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "paused",
		"run_id": runID,
	})
}

// HandleResumeRun handles POST /api/runs/{runId}/resume
func (s *Server) HandleResumeRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	// Validate runID
	if err := validation.ValidateRunID("run_id", runID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse request
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize and validate message
	req.Message = validation.SanitizeString(req.Message)
	validationReq := validation.ResumeRunRequest{Message: req.Message}
	if err := validationReq.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Delegate to service layer
	run, err := s.runService.ResumeRun(r.Context(), service.ResumeRunRequest{
		RunID:   runID,
		Message: req.Message,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(run)
}

func (s *Server) HandleAddInstruction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	// Validate runID
	if err := validation.ValidateRunID("run_id", runID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse request
	var req AddInstructionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize and validate instruction
	req.Instruction = validation.SanitizeString(req.Instruction)
	validationReq := validation.AddInstructionRequest{Instruction: req.Instruction}
	if err := validationReq.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Delegate to service layer
	if err := s.runService.AddInstruction(r.Context(), runID, req.Instruction); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "instruction_added",
		"run_id":      runID,
		"instruction": req.Instruction,
	})
}
