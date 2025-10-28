package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// HandleListPlans handles GET /api/plans
func (s *Server) HandleListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := s.planService.ListPlans(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

// HandleGetPlan handles GET /api/plans/:planId
func (s *Server) HandleGetPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	plan, err := s.planService.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// HandleCreatePlan handles POST /api/plans
func (s *Server) HandleCreatePlan(w http.ResponseWriter, r *http.Request) {
	var plan storage.ObservabilityPlan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.planService.CreatePlan(r.Context(), &plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(plan)
}

// HandleUpdatePlan handles PUT /api/plans/:planId
func (s *Server) HandleUpdatePlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	var update storage.PlanUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.planService.UpdatePlan(r.Context(), planID, &update); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleDeletePlan handles DELETE /api/plans/:planId
func (s *Server) HandleDeletePlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	if err := s.planService.DeletePlan(r.Context(), planID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// HandleGetTopology handles GET /api/plans/:planId/topology
func (s *Server) HandleGetTopology(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	topology, err := s.planService.GetTopology(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topology)
}

// Component-level CRUD operations for Services
func (s *Server) HandleCreateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	var service storage.InstrumentedService
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure plan_id matches
	service.PlanID = planID
	if service.ID == "" {
		service.ID = fmt.Sprintf("service-%d", time.Now().UnixNano())
	}
	if service.CreatedAt.IsZero() {
		service.CreatedAt = time.Now()
	}
	if service.UpdatedAt.IsZero() {
		service.UpdatedAt = time.Now()
	}

	if err := s.storage.CreateService(&service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(service)
}

func (s *Server) HandleUpdateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]
	serviceID := vars["serviceId"]

	var service storage.InstrumentedService
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure IDs match
	service.ID = serviceID
	service.PlanID = planID
	service.UpdatedAt = time.Now()

	if err := s.storage.UpdateService(serviceID, &service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["serviceId"]

	if err := s.storage.DeleteService(serviceID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// Component-level CRUD operations for Infrastructure
func (s *Server) HandleCreateInfrastructure(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	var infra storage.InfrastructureComponent
	if err := json.NewDecoder(r.Body).Decode(&infra); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure plan_id matches
	infra.PlanID = planID
	if infra.ID == "" {
		infra.ID = fmt.Sprintf("infra-%d", time.Now().UnixNano())
	}
	if infra.CreatedAt.IsZero() {
		infra.CreatedAt = time.Now()
	}
	if infra.UpdatedAt.IsZero() {
		infra.UpdatedAt = time.Now()
	}

	if err := s.storage.CreateInfrastructure(&infra); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(infra)
}

func (s *Server) HandleUpdateInfrastructure(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]
	infraID := vars["infraId"]

	var infra storage.InfrastructureComponent
	if err := json.NewDecoder(r.Body).Decode(&infra); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure IDs match
	infra.ID = infraID
	infra.PlanID = planID
	infra.UpdatedAt = time.Now()

	if err := s.storage.UpdateInfrastructure(infraID, &infra); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeleteInfrastructure(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	infraID := vars["infraId"]

	if err := s.storage.DeleteInfrastructure(infraID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// Component-level CRUD operations for Pipelines
func (s *Server) HandleCreatePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	var pipeline storage.CollectorPipeline
	if err := json.NewDecoder(r.Body).Decode(&pipeline); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure plan_id matches
	pipeline.PlanID = planID
	if pipeline.ID == "" {
		pipeline.ID = fmt.Sprintf("pipeline-%d", time.Now().UnixNano())
	}
	if pipeline.CreatedAt.IsZero() {
		pipeline.CreatedAt = time.Now()
	}
	if pipeline.UpdatedAt.IsZero() {
		pipeline.UpdatedAt = time.Now()
	}

	if err := s.storage.CreatePipeline(&pipeline); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pipeline)
}

func (s *Server) HandleUpdatePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]
	pipelineID := vars["pipelineId"]

	var pipeline storage.CollectorPipeline
	if err := json.NewDecoder(r.Body).Decode(&pipeline); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure IDs match
	pipeline.ID = pipelineID
	pipeline.PlanID = planID
	pipeline.UpdatedAt = time.Now()

	if err := s.storage.UpdatePipeline(pipelineID, &pipeline); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeletePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID := vars["pipelineId"]

	if err := s.storage.DeletePipeline(pipelineID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// Component-level CRUD operations for Backends
func (s *Server) HandleCreateBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	var backend storage.Backend
	if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure plan_id matches
	backend.PlanID = planID
	if backend.ID == "" {
		backend.ID = fmt.Sprintf("backend-%d", time.Now().UnixNano())
	}
	if backend.CreatedAt.IsZero() {
		backend.CreatedAt = time.Now()
	}
	if backend.UpdatedAt.IsZero() {
		backend.UpdatedAt = time.Now()
	}

	if err := s.storage.CreateBackend(&backend); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(backend)
}

func (s *Server) HandleUpdateBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]
	backendID := vars["backendId"]

	var backend storage.Backend
	if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure IDs match
	backend.ID = backendID
	backend.PlanID = planID
	backend.UpdatedAt = time.Now()

	if err := s.storage.UpdateBackend(backendID, &backend); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeleteBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backendID := vars["backendId"]

	if err := s.storage.DeleteBackend(backendID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

