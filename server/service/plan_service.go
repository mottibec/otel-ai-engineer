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

// AddServiceToPlan adds a service to a plan
func (ps *PlanService) AddServiceToPlan(ctx context.Context, service *storage.InstrumentedService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}
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

	return ps.storage.CreateService(service)
}

// AddInfrastructureToPlan adds infrastructure to a plan
func (ps *PlanService) AddInfrastructureToPlan(ctx context.Context, infra *storage.InfrastructureComponent) error {
	if infra == nil {
		return fmt.Errorf("infrastructure component cannot be nil")
	}
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

	return ps.storage.CreateInfrastructure(infra)
}

// AddPipelineToPlan adds a pipeline to a plan
func (ps *PlanService) AddPipelineToPlan(ctx context.Context, pipeline *storage.CollectorPipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be nil")
	}
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

	return ps.storage.CreatePipeline(pipeline)
}

// AddBackendToPlan adds a backend to a plan
func (ps *PlanService) AddBackendToPlan(ctx context.Context, backend *storage.Backend) error {
	if backend == nil {
		return fmt.Errorf("backend cannot be nil")
	}
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

	return ps.storage.CreateBackend(backend)
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

