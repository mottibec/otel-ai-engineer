package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	collectorService "github.com/mottibechhofer/otel-ai-engineer/server/service/collector"
)

// HandleListCollectors handles GET /api/collectors
func (s *Server) HandleListCollectors(w http.ResponseWriter, r *http.Request) {
	targetTypeStr := r.URL.Query().Get("target_type")

	response, err := s.collectorService.ListCollectors(r.Context(), targetTypeStr, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleListConnectedAgents handles GET /api/collectors/connected
func (s *Server) HandleListConnectedAgents(w http.ResponseWriter, r *http.Request) {
	response, err := s.collectorService.ListConnectedAgents(r.Context(), true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetCollectorConfig handles GET /api/collectors/:id/config
func (s *Server) HandleGetCollectorConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectorID := vars["id"]

	response, err := s.collectorService.GetCollectorConfig(r.Context(), collectorID, true)
	if err != nil {
		if err.Error() == "collector ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "OTEL client not initialized" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateCollectorConfig handles PUT /api/collectors/:id/config
func (s *Server) HandleUpdateCollectorConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectorID := vars["id"]

	var req collectorService.UpdateCollectorConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.collectorService.UpdateCollectorConfig(r.Context(), collectorID, req)
	if err != nil {
		if err.Error() == "collector ID cannot be empty" || err.Error() == "yaml_config is required" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "OTEL client not initialized" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDeployCollector handles POST /api/collectors
func (s *Server) HandleDeployCollector(w http.ResponseWriter, r *http.Request) {
	var req collectorService.DeployCollectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := s.collectorService.DeployCollector(r.Context(), req)
	if err != nil {
		if err.Error() == "collector_name is required" || err.Error() == "yaml_config is required" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// HandleStopCollector handles DELETE /api/collectors/:id
func (s *Server) HandleStopCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectorID := vars["id"]

	// Get target_type from query parameter (default to docker)
	targetTypeStr := r.URL.Query().Get("target_type")
	if targetTypeStr == "" {
		targetTypeStr = "docker"
	}

	result, err := s.collectorService.StopCollector(r.Context(), collectorID, targetTypeStr)
	if err != nil {
		if err.Error() == "collector ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleGetCollector handles GET /api/collectors/:id
func (s *Server) HandleGetCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectorID := vars["id"]

	collector, err := s.collectorService.GetCollector(r.Context(), collectorID, true)
	if err != nil {
		if err.Error() == "collector ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "collector not found" {
			http.Error(w, "Collector not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(collector)
}

// HandleGetCollectorLogs handles GET /api/collectors/:id/logs
func (s *Server) HandleGetCollectorLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectorID := vars["id"]

	// Get tail parameter (default to 100)
	tail := 100
	if tailStr := r.URL.Query().Get("tail"); tailStr != "" {
		if parsedTail, err := strconv.Atoi(tailStr); err == nil && parsedTail > 0 {
			tail = parsedTail
		}
	}

	response, err := s.collectorService.GetCollectorLogs(r.Context(), collectorID, tail)
	if err != nil {
		if err.Error() == "collector ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "logs only available for Docker collectors" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error()[:25] == "failed to get collector" || err.Error()[:22] == "collector not found" {
			http.Error(w, "Collector not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
