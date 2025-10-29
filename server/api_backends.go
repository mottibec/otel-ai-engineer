package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	backendService "github.com/mottibechhofer/otel-ai-engineer/server/service/backend"
)

// HandleListBackends handles GET /api/backends
func (s *Server) HandleListBackends(w http.ResponseWriter, r *http.Request) {
	responses, err := s.backendService.ListBackends(r.Context(), true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// HandleGetBackend handles GET /api/backends/:id
func (s *Server) HandleGetBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backendID := vars["id"]

	response, err := s.backendService.GetBackend(r.Context(), backendID, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCreateBackend handles POST /api/backends
func (s *Server) HandleCreateBackend(w http.ResponseWriter, r *http.Request) {
	var req backendService.CreateBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.backendService.CreateBackend(r.Context(), req)
	if err != nil {
		if err.Error() == "name is required" || err.Error() == "url is required" || err.Error() == "backend_type is required" {
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

// HandleUpdateBackend handles PUT /api/backends/:id
func (s *Server) HandleUpdateBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backendID := vars["id"]

	var req backendService.UpdateBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.backendService.UpdateBackend(r.Context(), backendID, req)
	if err != nil {
		if err.Error() == "failed to get backend: backend not found" {
			http.Error(w, "backend not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDeleteBackend handles DELETE /api/backends/:id
func (s *Server) HandleDeleteBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backendID := vars["id"]

	if err := s.backendService.DeleteBackend(r.Context(), backendID); err != nil {
		if err.Error() == "failed to delete backend: backend not found" {
			http.Error(w, "backend not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "deleted",
		"backend_id": backendID,
	})
}

// HandleTestConnection handles POST /api/backends/:id/test-connection
func (s *Server) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backendID := vars["id"]

	var req backendService.TestConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Request body is optional - use backend's stored credentials
		req = backendService.TestConnectionRequest{}
	}

	result, err := s.backendService.TestConnection(r.Context(), backendID, req)
	if err != nil {
		if err.Error() == "failed to get backend: backend not found" {
			http.Error(w, "backend not found", http.StatusNotFound)
			return
		}
		if err.Error() == "username and password required for Grafana" || err.Error()[:35] == "unsupported backend type:" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleConfigureGrafanaDatasource handles POST /api/backends/:id/configure-datasource
func (s *Server) HandleConfigureGrafanaDatasource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backendID := vars["id"]

	var req backendService.ConfigureGrafanaDatasourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := s.backendService.ConfigureGrafanaDatasource(r.Context(), backendID, req)
	if err != nil {
		if err.Error() == "failed to get backend: backend not found" {
			http.Error(w, "backend not found", http.StatusNotFound)
			return
		}
		if err.Error() == "only Grafana backends support datasource configuration" ||
			err.Error() == "datasource_name, datasource_type, and url are required" ||
			err.Error()[:25] == "Grafana credentials not" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

