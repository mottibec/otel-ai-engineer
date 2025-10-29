package storage

import (
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
)

// Storage is the interface for storing and retrieving agent runs and events
type Storage interface {
	// Run management
	CreateRun(run *Run) error
	GetRun(runID string) (*Run, error)
	ListRuns(opts RunListOptions) ([]*Run, error)
	UpdateRun(runID string, update *RunUpdate) error
	DeleteRun(runID string) error

	// Handoff queries
	GetSubRuns(parentRunID string) ([]*Run, error)
	GetParentRun(subRunID string) (*Run, error)

	// Event management
	AddEvent(runID string, event *events.AgentEvent) error
	GetEvents(runID string, after *time.Time) ([]*events.AgentEvent, error)
	GetEventCount(runID string) (int, error)

	// Stream support (for real-time updates)
	Subscribe(runID string) (<-chan *events.AgentEvent, func())
	SubscribeAll() (<-chan *events.AgentEvent, func())

	// Cleanup
	Close() error

	// Plan management
	CreatePlan(plan *ObservabilityPlan) error
	GetPlan(planID string) (*ObservabilityPlan, error)
	ListPlans() ([]*ObservabilityPlan, error)
	UpdatePlan(planID string, update *PlanUpdate) error
	DeletePlan(planID string) error

	// Service management
	CreateService(service *InstrumentedService) error
	GetService(serviceID string) (*InstrumentedService, error)
	GetServicesByPlan(planID string) ([]*InstrumentedService, error)
	UpdateService(serviceID string, service *InstrumentedService) error
	DeleteService(serviceID string) error

	// Infrastructure management
	CreateInfrastructure(infra *InfrastructureComponent) error
	GetInfrastructure(infraID string) (*InfrastructureComponent, error)
	GetInfrastructureByPlan(planID string) ([]*InfrastructureComponent, error)
	UpdateInfrastructure(infraID string, infra *InfrastructureComponent) error
	DeleteInfrastructure(infraID string) error

	// Pipeline management
	CreatePipeline(pipeline *CollectorPipeline) error
	GetPipeline(pipelineID string) (*CollectorPipeline, error)
	GetPipelinesByPlan(planID string) ([]*CollectorPipeline, error)
	UpdatePipeline(pipelineID string, pipeline *CollectorPipeline) error
	DeletePipeline(pipelineID string) error

	// Backend management
	CreateBackend(backend *Backend) error
	GetBackend(backendID string) (*Backend, error)
	GetBackendsByPlan(planID string) ([]*Backend, error)
	ListAllBackends() ([]*Backend, error)
	UpdateBackend(backendID string, backend *Backend) error
	DeleteBackend(backendID string) error

	// Dependency management
	CreateDependency(dep *PlanDependency) error
	GetDependenciesByPlan(planID string) ([]*PlanDependency, error)
	GetDependenciesBySource(sourceID string) ([]*PlanDependency, error)
	GetDependenciesByTarget(targetID string) ([]*PlanDependency, error)
	DeleteDependency(depID string) error

	// Agent work management
	CreateAgentWork(work *AgentWork) error
	GetAgentWork(workID string) (*AgentWork, error)
	GetAgentWorkByResource(resourceType ResourceType, resourceID string) ([]*AgentWork, error)
	ListAgentWork(opts AgentWorkListOptions) ([]*AgentWork, error)
	UpdateAgentWork(workID string, update *AgentWorkUpdate) error
	DeleteAgentWork(workID string) error

	// Human action management
	CreateHumanAction(action *HumanAction) error
	GetHumanAction(actionID string) (*HumanAction, error)
	ListHumanActions(opts HumanActionListOptions) ([]*HumanAction, error)
	UpdateHumanAction(actionID string, update *HumanActionUpdate) error
	DeleteHumanAction(actionID string) error
	GetHumanActionsByRun(runID string) ([]*HumanAction, error)
	GetPendingHumanActions() ([]*HumanAction, error)

	// Custom agent management
	CreateCustomAgent(agent *CustomAgent) error
	GetCustomAgent(agentID string) (*CustomAgent, error)
	ListCustomAgents() ([]*CustomAgent, error)
	UpdateCustomAgent(agentID string, update *CustomAgentUpdate) error
	DeleteCustomAgent(agentID string) error
}
