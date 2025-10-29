package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/sandbox"
)

// HandleListSandboxes handles GET /api/sandboxes
func (s *Server) HandleListSandboxes(w http.ResponseWriter, r *http.Request) {
	response, err := s.sandboxService.ListSandboxes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetSandbox handles GET /api/sandboxes/{id}
func (s *Server) HandleGetSandbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sandboxID := vars["id"]

	sandbox, err := s.sandboxService.GetSandbox(r.Context(), sandboxID)
	if err != nil {
		if err.Error() == "sandbox ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sandbox)
}

// HandleCreateSandbox handles POST /api/sandboxes
func (s *Server) HandleCreateSandbox(w http.ResponseWriter, r *http.Request) {
	var req sandbox.CreateSandboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.sandboxService.CreateSandbox(r.Context(), req)
	if err != nil {
		if err.Error() == "name is required" || err.Error() == "collector_config is required" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// HandleStartTelemetry handles POST /api/sandboxes/{id}/telemetry
func (s *Server) HandleStartTelemetry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sandboxID := vars["id"]

	var req sandbox.StartSandboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.sandboxService.StartTelemetry(r.Context(), sandboxID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleValidateSandbox handles POST /api/sandboxes/{id}/validate
func (s *Server) HandleValidateSandbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sandboxID := vars["id"]

	var req sandbox.ValidateSandboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.sandboxService.ValidateSandbox(r.Context(), sandboxID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetSandboxLogs handles GET /api/sandboxes/{id}/logs
func (s *Server) HandleGetSandboxLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sandboxID := vars["id"]

	tailStr := r.URL.Query().Get("tail")
	tail := 100
	if tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil {
			tail = t
		}
	}

	response, err := s.sandboxService.GetSandboxLogs(r.Context(), sandboxID, tail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetSandboxMetrics handles GET /api/sandboxes/{id}/metrics
func (s *Server) HandleGetSandboxMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sandboxID := vars["id"]

	response, err := s.sandboxService.GetSandboxMetrics(r.Context(), sandboxID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleStopSandbox handles POST /api/sandboxes/{id}/stop
func (s *Server) HandleStopSandbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sandboxID := vars["id"]

	response, err := s.sandboxService.StopSandbox(r.Context(), sandboxID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDeleteSandbox handles DELETE /api/sandboxes/{id}
func (s *Server) HandleDeleteSandbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sandboxID := vars["id"]

	response, err := s.sandboxService.DeleteSandbox(r.Context(), sandboxID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
