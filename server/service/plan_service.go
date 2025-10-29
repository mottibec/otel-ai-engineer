package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// PlanService handles business logic for observability plan management
type PlanService struct {
	storage storage.Storage
}

// NewPlanService creates a new plan service
func NewPlanService(stor storage.Storage) *PlanService {
	return &PlanService{
		storage: stor,
	}
}

// CreatePlan creates a new observability plan
func (ps *PlanService) CreatePlan(ctx context.Context, plan *storage.ObservabilityPlan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}
	if plan.Name == "" {
		return fmt.Errorf("plan name cannot be empty")
	}

	// Generate ID if not provided
	if plan.ID == "" {
		plan.ID = fmt.Sprintf("plan-%d", time.Now().UnixNano())
	}

	// Set defaults
	if plan.Status == "" {
		plan.Status = storage.PlanStatusDraft
	}
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = time.Now()
	}
	if plan.UpdatedAt.IsZero() {
		plan.UpdatedAt = time.Now()
	}

	return ps.storage.CreatePlan(plan)
}

// GetPlan retrieves a plan with all components
func (ps *PlanService) GetPlan(ctx context.Context, planID string) (*storage.ObservabilityPlan, error) {
	plan, err := ps.storage.GetPlan(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Calculate aggregated status based on component statuses
	plan.Status = ps.calculateAggregatedStatus(plan)

	return plan, nil
}

// ListPlans retrieves all plans
func (ps *PlanService) ListPlans(ctx context.Context) ([]*storage.ObservabilityPlan, error) {
	plans, err := ps.storage.ListPlans()
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	// Calculate status for each plan
	for _, plan := range plans {
		fullPlan, err := ps.GetPlan(ctx, plan.ID)
		if err == nil {
			plan.Status = fullPlan.Status
		}
	}

	return plans, nil
}

// UpdatePlan updates a plan
func (ps *PlanService) UpdatePlan(ctx context.Context, planID string, update *storage.PlanUpdate) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	return ps.storage.UpdatePlan(planID, update)
}

// DeletePlan deletes a plan and all its components
func (ps *PlanService) DeletePlan(ctx context.Context, planID string) error {
	return ps.storage.DeletePlan(planID)
}

// calculateAggregatedStatus determines plan status based on component statuses
func (ps *PlanService) calculateAggregatedStatus(plan *storage.ObservabilityPlan) storage.PlanStatus {
	// Check if any components exist
	hasComponents := len(plan.Services) > 0 || len(plan.Infrastructure) > 0 || 
	                 len(plan.Pipelines) > 0 || len(plan.Backends) > 0

	if !hasComponents {
		return storage.PlanStatusDraft
	}

	// Count component statuses
	allHealthy := true
	hasPartials := false
	hasErrors := false

	// Check services
	for _, service := range plan.Services {
		switch service.Status {
		case "success", "healthy":
			// healthy
		case "pending", "in_progress":
			hasPartials = true
			allHealthy = false
		default:
			hasErrors = true
			allHealthy = false
		}
	}

	// Check infrastructure
	for _, infra := range plan.Infrastructure {
		switch infra.Status {
		case "success", "healthy":
			// healthy
		case "pending", "in_progress":
			hasPartials = true
			allHealthy = false
		default:
			hasErrors = true
			allHealthy = false
		}
	}

	// Check pipelines
	for _, pipeline := range plan.Pipelines {
		switch pipeline.Status {
		case "success", "healthy":
			// healthy
		case "pending", "in_progress":
			hasPartials = true
			allHealthy = false
		default:
			hasErrors = true
			allHealthy = false
		}
	}

	// Check backends
	for _, backend := range plan.Backends {
		switch backend.HealthStatus {
		case "healthy":
			// healthy
		case "unknown":
			hasPartials = true
			allHealthy = false
		case "unhealthy":
			hasErrors = true
			allHealthy = false
		}
	}

	// Determine status based on component states
	if hasErrors {
		return storage.PlanStatusFailed
	}
	if hasPartials && !hasErrors {
		return storage.PlanStatusPartial
	}
	if allHealthy {
		return storage.PlanStatusSuccess
	}

	return storage.PlanStatusPending
}

// CreateService creates a service and adds it to a plan
func (ps *PlanService) CreateService(ctx context.Context, planID string, service *storage.InstrumentedService) (*storage.InstrumentedService, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}
	if service == nil {
		return nil, fmt.Errorf("service cannot be nil")
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
	if service.Status == "" {
		service.Status = "pending"
	}

	if err := ps.storage.CreateService(service); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return service, nil
}

// UpdateService updates a service
func (ps *PlanService) UpdateService(ctx context.Context, planID, serviceID string, service *storage.InstrumentedService) error {
	if planID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if serviceID == "" {
		return fmt.Errorf("service ID cannot be empty")
	}
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	// Ensure IDs match
	service.ID = serviceID
	service.PlanID = planID
	service.UpdatedAt = time.Now()

	if err := ps.storage.UpdateService(serviceID, service); err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	return nil
}

// DeleteService deletes a service
func (ps *PlanService) DeleteService(ctx context.Context, serviceID string) error {
	if serviceID == "" {
		return fmt.Errorf("service ID cannot be empty")
	}

	if err := ps.storage.DeleteService(serviceID); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// AddServiceToPlan adds a service to a plan (legacy method, forwards to CreateService)
func (ps *PlanService) AddServiceToPlan(ctx context.Context, service *storage.InstrumentedService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}
	if service.PlanID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}

	_, err := ps.CreateService(ctx, service.PlanID, service)
	return err
}

// CreateInfrastructure creates an infrastructure component and adds it to a plan
func (ps *PlanService) CreateInfrastructure(ctx context.Context, planID string, infra *storage.InfrastructureComponent) (*storage.InfrastructureComponent, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}
	if infra == nil {
		return nil, fmt.Errorf("infrastructure component cannot be nil")
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
	if infra.Status == "" {
		infra.Status = "pending"
	}

	if err := ps.storage.CreateInfrastructure(infra); err != nil {
		return nil, fmt.Errorf("failed to create infrastructure: %w", err)
	}

	return infra, nil
}

// UpdateInfrastructure updates an infrastructure component
func (ps *PlanService) UpdateInfrastructure(ctx context.Context, planID, infraID string, infra *storage.InfrastructureComponent) error {
	if planID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if infraID == "" {
		return fmt.Errorf("infrastructure ID cannot be empty")
	}
	if infra == nil {
		return fmt.Errorf("infrastructure component cannot be nil")
	}

	// Ensure IDs match
	infra.ID = infraID
	infra.PlanID = planID
	infra.UpdatedAt = time.Now()

	if err := ps.storage.UpdateInfrastructure(infraID, infra); err != nil {
		return fmt.Errorf("failed to update infrastructure: %w", err)
	}

	return nil
}

// DeleteInfrastructure deletes an infrastructure component
func (ps *PlanService) DeleteInfrastructure(ctx context.Context, infraID string) error {
	if infraID == "" {
		return fmt.Errorf("infrastructure ID cannot be empty")
	}

	if err := ps.storage.DeleteInfrastructure(infraID); err != nil {
		return fmt.Errorf("failed to delete infrastructure: %w", err)
	}

	return nil
}

// AddInfrastructureToPlan adds infrastructure to a plan (legacy method, forwards to CreateInfrastructure)
func (ps *PlanService) AddInfrastructureToPlan(ctx context.Context, infra *storage.InfrastructureComponent) error {
	if infra == nil {
		return fmt.Errorf("infrastructure component cannot be nil")
	}
	if infra.PlanID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}

	_, err := ps.CreateInfrastructure(ctx, infra.PlanID, infra)
	return err
}

// CreatePipeline creates a pipeline and adds it to a plan
func (ps *PlanService) CreatePipeline(ctx context.Context, planID string, pipeline *storage.CollectorPipeline) (*storage.CollectorPipeline, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}
	if pipeline == nil {
		return nil, fmt.Errorf("pipeline cannot be nil")
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
	if pipeline.Status == "" {
		pipeline.Status = "pending"
	}

	if err := ps.storage.CreatePipeline(pipeline); err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	return pipeline, nil
}

// UpdatePipeline updates a pipeline
func (ps *PlanService) UpdatePipeline(ctx context.Context, planID, pipelineID string, pipeline *storage.CollectorPipeline) error {
	if planID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if pipelineID == "" {
		return fmt.Errorf("pipeline ID cannot be empty")
	}
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be nil")
	}

	// Ensure IDs match
	pipeline.ID = pipelineID
	pipeline.PlanID = planID
	pipeline.UpdatedAt = time.Now()

	if err := ps.storage.UpdatePipeline(pipelineID, pipeline); err != nil {
		return fmt.Errorf("failed to update pipeline: %w", err)
	}

	return nil
}

// DeletePipeline deletes a pipeline
func (ps *PlanService) DeletePipeline(ctx context.Context, pipelineID string) error {
	if pipelineID == "" {
		return fmt.Errorf("pipeline ID cannot be empty")
	}

	if err := ps.storage.DeletePipeline(pipelineID); err != nil {
		return fmt.Errorf("failed to delete pipeline: %w", err)
	}

	return nil
}

// AddPipelineToPlan adds a pipeline to a plan (legacy method, forwards to CreatePipeline)
func (ps *PlanService) AddPipelineToPlan(ctx context.Context, pipeline *storage.CollectorPipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be nil")
	}
	if pipeline.PlanID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}

	_, err := ps.CreatePipeline(ctx, pipeline.PlanID, pipeline)
	return err
}

// CreatePlanBackend creates a backend and adds it to a plan
func (ps *PlanService) CreatePlanBackend(ctx context.Context, planID string, backend *storage.Backend) (*storage.Backend, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan ID cannot be empty")
	}
	if backend == nil {
		return nil, fmt.Errorf("backend cannot be nil")
	}

	// Ensure plan_id matches
	backend.PlanID = &planID
	if backend.ID == "" {
		backend.ID = fmt.Sprintf("backend-%d", time.Now().UnixNano())
	}
	if backend.CreatedAt.IsZero() {
		backend.CreatedAt = time.Now()
	}
	if backend.UpdatedAt.IsZero() {
		backend.UpdatedAt = time.Now()
	}
	if backend.HealthStatus == "" {
		backend.HealthStatus = "unknown"
	}

	if err := ps.storage.CreateBackend(backend); err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	return backend, nil
}

// UpdatePlanBackend updates a backend in a plan
func (ps *PlanService) UpdatePlanBackend(ctx context.Context, planID, backendID string, backend *storage.Backend) error {
	if planID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if backendID == "" {
		return fmt.Errorf("backend ID cannot be empty")
	}
	if backend == nil {
		return fmt.Errorf("backend cannot be nil")
	}

	// Ensure IDs match
	backend.ID = backendID
	backend.PlanID = &planID
	backend.UpdatedAt = time.Now()

	if err := ps.storage.UpdateBackend(backendID, backend); err != nil {
		return fmt.Errorf("failed to update backend: %w", err)
	}

	return nil
}

// DeletePlanBackend deletes a backend from a plan
func (ps *PlanService) DeletePlanBackend(ctx context.Context, backendID string) error {
	if backendID == "" {
		return fmt.Errorf("backend ID cannot be empty")
	}

	if err := ps.storage.DeleteBackend(backendID); err != nil {
		return fmt.Errorf("failed to delete backend: %w", err)
	}

	return nil
}

// AttachBackendToPlan attaches an existing backend to a plan
func (ps *PlanService) AttachBackendToPlan(ctx context.Context, planID, backendID string) error {
	if planID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if backendID == "" {
		return fmt.Errorf("backend ID cannot be empty")
	}

	// Get the backend
	backend, err := ps.storage.GetBackend(backendID)
	if err != nil {
		return fmt.Errorf("failed to get backend: %w", err)
	}

	// Update plan_id
	backend.PlanID = &planID
	backend.UpdatedAt = time.Now()

	if err := ps.storage.UpdateBackend(backendID, backend); err != nil {
		return fmt.Errorf("failed to attach backend to plan: %w", err)
	}

	return nil
}

// AddBackendToPlan adds a backend to a plan (legacy method, forwards to CreatePlanBackend)
func (ps *PlanService) AddBackendToPlan(ctx context.Context, backend *storage.Backend) error {
	if backend == nil {
		return fmt.Errorf("backend cannot be nil")
	}
	if backend.PlanID == nil || *backend.PlanID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}

	_, err := ps.CreatePlanBackend(ctx, *backend.PlanID, backend)
	return err
}

// AddDependencyToPlan adds a dependency relationship to a plan
func (ps *PlanService) AddDependencyToPlan(ctx context.Context, dep *storage.PlanDependency) error {
	if dep == nil {
		return fmt.Errorf("dependency cannot be nil")
	}
	if dep.ID == "" {
		dep.ID = fmt.Sprintf("dep-%d", time.Now().UnixNano())
	}
	if dep.CreatedAt.IsZero() {
		dep.CreatedAt = time.Now()
	}

	return ps.storage.CreateDependency(dep)
}

// GetTopology retrieves the dependency topology for a plan
func (ps *PlanService) GetTopology(ctx context.Context, planID string) (*TopologyGraph, error) {
	plan, err := ps.storage.GetPlan(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	deps, err := ps.storage.GetDependenciesByPlan(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	topology := &TopologyGraph{
		Nodes: []TopologyNode{},
		Edges: []TopologyEdge{},
	}

	// Add service nodes
	for _, service := range plan.Services {
		topology.Nodes = append(topology.Nodes, TopologyNode{
			ID:       service.ID,
			Type:     "service",
			Label:    service.ServiceName,
			Status:   service.Status,
			Metadata: map[string]interface{}{
				"language": service.Language,
				"framework": service.Framework,
			},
		})
	}

	// Add infrastructure nodes
	for _, infra := range plan.Infrastructure {
		topology.Nodes = append(topology.Nodes, TopologyNode{
			ID:     infra.ID,
			Type:   "infrastructure",
			Label:  infra.Name,
			Status: infra.Status,
			Metadata: map[string]interface{}{
				"component_type": infra.ComponentType,
				"receiver_type":  infra.ReceiverType,
			},
		})
	}

	// Add pipeline nodes
	for _, pipeline := range plan.Pipelines {
		topology.Nodes = append(topology.Nodes, TopologyNode{
			ID:     pipeline.ID,
			Type:   "pipeline",
			Label:  pipeline.Name,
			Status: pipeline.Status,
			Metadata: map[string]interface{}{
				"target_type": pipeline.TargetType,
			},
		})
	}

	// Add backend nodes
	for _, backend := range plan.Backends {
		topology.Nodes = append(topology.Nodes, TopologyNode{
			ID:     backend.ID,
			Type:   "backend",
			Label:  backend.Name,
			Status: backend.HealthStatus,
			Metadata: map[string]interface{}{
				"backend_type": backend.BackendType,
				"url":          backend.URL,
			},
		})
	}

	// Add edges
	for _, dep := range deps {
		topology.Edges = append(topology.Edges, TopologyEdge{
			SourceID: dep.SourceID,
			TargetID: dep.TargetID,
			Type:     dep.DependencyType,
		})
	}

	return topology, nil
}

// TopologyGraph represents the dependency graph
type TopologyGraph struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

// TopologyNode represents a node in the topology graph
type TopologyNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Label    string                 `json:"label"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TopologyEdge represents an edge in the topology graph
type TopologyEdge struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Type     string `json:"type"`
}

