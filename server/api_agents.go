package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	agentService "github.com/mottibechhofer/otel-ai-engineer/server/service/agent"
)

// HandleListAgents handles GET /api/agents
func (s *Server) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	if s.agentService == nil {
		http.Error(w, "Agent service not initialized", http.StatusInternalServerError)
		return
	}

	agents, err := s.agentService.ListAgents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// HandleGetAgent handles GET /api/agents/:agentId
func (s *Server) HandleGetAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentService == nil {
		http.Error(w, "Agent service not initialized", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	agentID := vars["agentId"]

	agent, err := s.agentService.GetAgent(r.Context(), agentID)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// HandleGetAgentTools handles GET /api/agents/:agentId/tools
func (s *Server) HandleGetAgentTools(w http.ResponseWriter, r *http.Request) {
	if s.agentService == nil {
		http.Error(w, "Agent service not initialized", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	agentID := vars["agentId"]

	tools, err := s.agentService.GetAgentTools(r.Context(), agentID)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

// HandleListTools handles GET /api/tools
func (s *Server) HandleListTools(w http.ResponseWriter, r *http.Request) {
	if s.toolDiscoveryService == nil {
		http.Error(w, "Tool discovery service not initialized", http.StatusInternalServerError)
		return
	}

	response, err := s.toolDiscoveryService.GetAllTools(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCreateCustomAgent handles POST /api/agents/custom
func (s *Server) HandleCreateCustomAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentService == nil {
		http.Error(w, "Agent service not initialized", http.StatusInternalServerError)
		return
	}

	var req agentService.CreateCustomAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.agentService.CreateCustomAgent(r.Context(), req)
	if err != nil {
		if err.Error() == "name is required" || err.Error() == "description is required" ||
			err.Error() == "at least one tool is required" {
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

// HandleUpdateCustomAgent handles PUT /api/agents/custom/:agentId
func (s *Server) HandleUpdateCustomAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentService == nil {
		http.Error(w, "Agent service not initialized", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	agentID := vars["agentId"]

	var req agentService.UpdateCustomAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.agentService.UpdateCustomAgent(r.Context(), agentID, req)
	if err != nil {
		if err.Error() == "agent not found: "+agentID || err.Error() == "custom agent not found" {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDeleteCustomAgent handles DELETE /api/agents/custom/:agentId
func (s *Server) HandleDeleteCustomAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentService == nil {
		http.Error(w, "Agent service not initialized", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	agentID := vars["agentId"]

	if err := s.agentService.DeleteCustomAgent(r.Context(), agentID); err != nil {
		if err.Error() == "custom agent not found" {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "deleted",
		"agent_id": agentID,
	})
}

// HandleCreateMetaAgent handles POST /api/agents/meta
func (s *Server) HandleCreateMetaAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentService == nil {
		http.Error(w, "Agent service not initialized", http.StatusInternalServerError)
		return
	}

	var req agentService.CreateMetaAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.agentService.CreateMetaAgent(r.Context(), req)
	if err != nil {
		if err.Error() == "name is required" || err.Error() == "description is required" ||
			err.Error() == "at least one tool is required" {
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

