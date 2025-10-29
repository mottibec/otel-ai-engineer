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

	createdService, err := s.planService.CreateService(r.Context(), planID, &service)
	if err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "service cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error()[:26] == "failed to create service" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdService)
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

	if err := s.planService.UpdateService(r.Context(), planID, serviceID, &service); err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "service ID cannot be empty" || err.Error() == "service cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["serviceId"]

	if err := s.planService.DeleteService(r.Context(), serviceID); err != nil {
		if err.Error() == "service ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error()[:26] == "failed to delete service" {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	createdInfra, err := s.planService.CreateInfrastructure(r.Context(), planID, &infra)
	if err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "infrastructure component cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdInfra)
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

	if err := s.planService.UpdateInfrastructure(r.Context(), planID, infraID, &infra); err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "infrastructure ID cannot be empty" || err.Error() == "infrastructure component cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeleteInfrastructure(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	infraID := vars["infraId"]

	if err := s.planService.DeleteInfrastructure(r.Context(), infraID); err != nil {
		if err.Error() == "infrastructure ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error()[:31] == "failed to delete infrastructure" {
			http.Error(w, "Infrastructure not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	createdPipeline, err := s.planService.CreatePipeline(r.Context(), planID, &pipeline)
	if err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "pipeline cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPipeline)
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

	if err := s.planService.UpdatePipeline(r.Context(), planID, pipelineID, &pipeline); err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "pipeline ID cannot be empty" || err.Error() == "pipeline cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeletePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID := vars["pipelineId"]

	if err := s.planService.DeletePipeline(r.Context(), pipelineID); err != nil {
		if err.Error() == "pipeline ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error()[:27] == "failed to delete pipeline" {
			http.Error(w, "Pipeline not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// Component-level CRUD operations for Backends
func (s *Server) HandleCreatePlanBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	var backend storage.Backend
	if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdBackend, err := s.planService.CreatePlanBackend(r.Context(), planID, &backend)
	if err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "backend cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBackend)
}

func (s *Server) HandleUpdatePlanBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]
	backendID := vars["backendId"]

	var backend storage.Backend
	if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.planService.UpdatePlanBackend(r.Context(), planID, backendID, &backend); err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "backend ID cannot be empty" || err.Error() == "backend cannot be nil" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (s *Server) HandleDeletePlanBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backendID := vars["backendId"]

	if err := s.planService.DeletePlanBackend(r.Context(), backendID); err != nil {
		if err.Error() == "backend ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error()[:26] == "failed to delete backend" {
			http.Error(w, "Backend not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// HandleAttachBackendToPlan handles PUT /api/plans/{planId}/backends/{backendId}/attach
// Associates an existing backend with a plan by updating its plan_id
func (s *Server) HandleAttachBackendToPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]
	backendID := vars["backendId"]

	// Verify plan exists
	_, err := s.planService.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}

	// Get the existing backend to check current state
	backend, err := s.backendService.GetBackend(r.Context(), backendID, false)
	if err != nil {
		http.Error(w, "Backend not found", http.StatusNotFound)
		return
	}

	// Check if backend is already associated with this plan
	if backend.Backend.PlanID != nil && *backend.Backend.PlanID == planID {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(backend.Backend)
		return
	}

	// Check if backend is already associated with a different plan
	if backend.Backend.PlanID != nil && *backend.Backend.PlanID != planID {
		http.Error(w, fmt.Sprintf("Backend is already associated with plan %s", *backend.Backend.PlanID), http.StatusBadRequest)
		return
	}

	// Attach backend to plan using service
	if err := s.planService.AttachBackendToPlan(r.Context(), planID, backendID); err != nil {
		if err.Error() == "plan ID cannot be empty" || err.Error() == "backend ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get agent work for this backend
	resourceType := storage.ResourceTypeBackend
	works, _ := s.storage.GetAgentWorkByResource(resourceType, backendID)

	// Return backend with agent work info (similar to BackendResponse)
	response := map[string]interface{}{
		"id":             backend.ID,
		"plan_id":        backend.PlanID,
		"backend_type":   backend.BackendType,
		"name":           backend.Name,
		"url":            backend.URL,
		"credentials":   backend.Credentials,
		"health_status":  backend.HealthStatus,
		"last_check":     backend.LastCheck,
		"datasource_uid": backend.DatasourceUID,
		"config":         backend.Config,
		"created_at":     backend.CreatedAt,
		"updated_at":     backend.UpdatedAt,
	}
	if len(works) > 0 {
		response["agent_work"] = works
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreatePipelineFromCollectorRequest represents the request to create a pipeline from a collector
type CreatePipelineFromCollectorRequest struct {
	CollectorID string `json:"collector_id"`
	Name         string `json:"name,omitempty"`
	ConfigYAML   string `json:"config_yaml,omitempty"`
	Rules        string `json:"rules,omitempty"`
	TargetType   string `json:"target_type,omitempty"`
}

// HandleCreatePipelineFromCollector handles POST /api/plans/{planId}/pipelines/from-collector
// Creates a CollectorPipeline entry referencing an existing collector
func (s *Server) HandleCreatePipelineFromCollector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["planId"]

	// Verify plan exists
	_, err := s.planService.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}

	var req CreatePipelineFromCollectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CollectorID == "" {
		http.Error(w, "collector_id is required", http.StatusBadRequest)
		return
	}

	// Generate pipeline name if not provided
	pipelineName := req.Name
	if pipelineName == "" {
		pipelineName = fmt.Sprintf("pipeline-%s", req.CollectorID)
	}

	// Default target_type if not provided
	targetType := req.TargetType
	if targetType == "" {
		targetType = "docker" // Default, but should ideally be fetched from collector
	}

	// Default rules if not provided
	rules := req.Rules
	if rules == "" {
		rules = "{}"
	}

	// Create pipeline entry
	pipeline := &storage.CollectorPipeline{
		ID:          fmt.Sprintf("pipeline-%d", time.Now().UnixNano()),
		PlanID:      planID,
		CollectorID: req.CollectorID,
		Name:        pipelineName,
		ConfigYAML:  req.ConfigYAML,
		Rules:       rules,
		Status:      "pending",
		TargetType:  targetType,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.storage.CreatePipeline(pipeline); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pipeline)
}

