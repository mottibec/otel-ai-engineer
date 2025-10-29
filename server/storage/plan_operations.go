package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// CreatePlan creates a new observability plan
func (s *SQLiteStorage) CreatePlan(plan *ObservabilityPlan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}
	if plan.ID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO observability_plans (id, name, description, environment, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		plan.ID, plan.Name, plan.Description, plan.Environment, plan.Status, plan.CreatedAt, plan.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert plan: %w", err)
	}

	return nil
}

// GetPlan retrieves a plan by ID with all components
func (s *SQLiteStorage) GetPlan(planID string) (*ObservabilityPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, name, description, environment, status, created_at, updated_at
		 FROM observability_plans WHERE id = ?`,
		planID)

	var plan ObservabilityPlan
	err := row.Scan(
		&plan.ID, &plan.Name, &plan.Description, &plan.Environment, &plan.Status,
		&plan.CreatedAt, &plan.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan with ID %s not found", planID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan plan: %w", err)
	}

	// Load components
	if plan.Services, err = s.GetServicesByPlan(planID); err != nil {
		return nil, fmt.Errorf("failed to load services: %w", err)
	}
	if plan.Infrastructure, err = s.GetInfrastructureByPlan(planID); err != nil {
		return nil, fmt.Errorf("failed to load infrastructure: %w", err)
	}
	if plan.Pipelines, err = s.GetPipelinesByPlan(planID); err != nil {
		return nil, fmt.Errorf("failed to load pipelines: %w", err)
	}
	if plan.Backends, err = s.GetBackendsByPlan(planID); err != nil {
		return nil, fmt.Errorf("failed to load backends: %w", err)
	}
	if plan.Dependencies, err = s.GetDependenciesByPlan(planID); err != nil {
		return nil, fmt.Errorf("failed to load dependencies: %w", err)
	}

	return &plan, nil
}

// ListPlans retrieves all plans
func (s *SQLiteStorage) ListPlans() ([]*ObservabilityPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, name, description, environment, status, created_at, updated_at
		 FROM observability_plans ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	plans := []*ObservabilityPlan{}
	for rows.Next() {
		var plan ObservabilityPlan
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.Description, &plan.Environment, &plan.Status,
			&plan.CreatedAt, &plan.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		plans = append(plans, &plan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate plans: %w", err)
	}

	return plans, nil
}

// UpdatePlan updates a plan
func (s *SQLiteStorage) UpdatePlan(planID string, update *PlanUpdate) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	updates := []string{}
	args := []interface{}{}

	if update.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *update.Name)
	}
	if update.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *update.Description)
	}
	if update.Environment != nil {
		updates = append(updates, "environment = ?")
		args = append(args, *update.Environment)
	}
	if update.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *update.Status)
	}

	if len(updates) == 0 {
		return nil
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, planID)

	query := fmt.Sprintf("UPDATE observability_plans SET %s WHERE id = ?", formatUpdateClause(updates))

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	return nil
}

// DeletePlan deletes a plan and its components
func (s *SQLiteStorage) DeletePlan(planID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM observability_plans WHERE id = ?", planID)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	return nil
}

// Service management operations
func (s *SQLiteStorage) CreateService(service *InstrumentedService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO instrumented_services 
		 (id, plan_id, service_name, language, framework, sdk_version, config_file, status, 
		  code_changes_summary, target_path, exporter_endpoint, git_repo_url, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		service.ID, service.PlanID, service.ServiceName, service.Language, service.Framework,
		service.SDKVersion, service.ConfigFile, service.Status, service.CodeChangesSummary,
		service.TargetPath, service.ExporterEndpoint, service.GitRepoURL, service.CreatedAt, service.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert service: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetService(serviceID string) (*InstrumentedService, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, plan_id, service_name, language, framework, sdk_version, config_file, status,
		 code_changes_summary, target_path, exporter_endpoint, git_repo_url, created_at, updated_at
		 FROM instrumented_services WHERE id = ?`,
		serviceID)

	var service InstrumentedService
	err := row.Scan(
		&service.ID, &service.PlanID, &service.ServiceName, &service.Language, &service.Framework,
		&service.SDKVersion, &service.ConfigFile, &service.Status, &service.CodeChangesSummary,
		&service.TargetPath, &service.ExporterEndpoint, &service.GitRepoURL, &service.CreatedAt, &service.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("service with ID %s not found", serviceID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan service: %w", err)
	}

	return &service, nil
}

func (s *SQLiteStorage) GetServicesByPlan(planID string) ([]*InstrumentedService, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, service_name, language, framework, sdk_version, config_file, status,
		 code_changes_summary, target_path, exporter_endpoint, git_repo_url, created_at, updated_at
		 FROM instrumented_services WHERE plan_id = ? ORDER BY created_at ASC`,
		planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	services := []*InstrumentedService{}
	for rows.Next() {
		var service InstrumentedService
		err := rows.Scan(
			&service.ID, &service.PlanID, &service.ServiceName, &service.Language, &service.Framework,
			&service.SDKVersion, &service.ConfigFile, &service.Status, &service.CodeChangesSummary,
			&service.TargetPath, &service.ExporterEndpoint, &service.GitRepoURL, &service.CreatedAt, &service.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, &service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate services: %w", err)
	}

	return services, nil
}

func (s *SQLiteStorage) UpdateService(serviceID string, service *InstrumentedService) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`UPDATE instrumented_services SET 
		 plan_id = ?, service_name = ?, language = ?, framework = ?, sdk_version = ?, 
		 config_file = ?, status = ?, code_changes_summary = ?, target_path = ?, 
		 exporter_endpoint = ?, git_repo_url = ?, updated_at = ?
		 WHERE id = ?`,
		service.PlanID, service.ServiceName, service.Language, service.Framework, service.SDKVersion,
		service.ConfigFile, service.Status, service.CodeChangesSummary, service.TargetPath,
		service.ExporterEndpoint, service.GitRepoURL, time.Now(), serviceID)

	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) DeleteService(serviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM instrumented_services WHERE id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// Infrastructure management operations
func (s *SQLiteStorage) CreateInfrastructure(infra *InfrastructureComponent) error {
	if infra == nil {
		return fmt.Errorf("infrastructure component cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO infrastructure_components 
		 (id, plan_id, component_type, name, host, receiver_type, metrics_collected, status, config, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		infra.ID, infra.PlanID, infra.ComponentType, infra.Name, infra.Host, infra.ReceiverType,
		infra.MetricsCollected, infra.Status, infra.Config, infra.CreatedAt, infra.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert infrastructure: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetInfrastructure(infraID string) (*InfrastructureComponent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, plan_id, component_type, name, host, receiver_type, metrics_collected, status, config, created_at, updated_at
		 FROM infrastructure_components WHERE id = ?`,
		infraID)

	var infra InfrastructureComponent
	err := row.Scan(
		&infra.ID, &infra.PlanID, &infra.ComponentType, &infra.Name, &infra.Host, &infra.ReceiverType,
		&infra.MetricsCollected, &infra.Status, &infra.Config, &infra.CreatedAt, &infra.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("infrastructure with ID %s not found", infraID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan infrastructure: %w", err)
	}

	return &infra, nil
}

func (s *SQLiteStorage) GetInfrastructureByPlan(planID string) ([]*InfrastructureComponent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, component_type, name, host, receiver_type, metrics_collected, status, config, created_at, updated_at
		 FROM infrastructure_components WHERE plan_id = ? ORDER BY created_at ASC`,
		planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query infrastructure: %w", err)
	}
	defer rows.Close()

	infrastructures := []*InfrastructureComponent{}
	for rows.Next() {
		var infra InfrastructureComponent
		err := rows.Scan(
			&infra.ID, &infra.PlanID, &infra.ComponentType, &infra.Name, &infra.Host, &infra.ReceiverType,
			&infra.MetricsCollected, &infra.Status, &infra.Config, &infra.CreatedAt, &infra.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan infrastructure: %w", err)
		}
		infrastructures = append(infrastructures, &infra)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate infrastructure: %w", err)
	}

	return infrastructures, nil
}

func (s *SQLiteStorage) UpdateInfrastructure(infraID string, infra *InfrastructureComponent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`UPDATE infrastructure_components SET 
		 plan_id = ?, component_type = ?, name = ?, host = ?, receiver_type = ?, 
		 metrics_collected = ?, status = ?, config = ?, updated_at = ?
		 WHERE id = ?`,
		infra.PlanID, infra.ComponentType, infra.Name, infra.Host, infra.ReceiverType,
		infra.MetricsCollected, infra.Status, infra.Config, time.Now(), infraID)

	if err != nil {
		return fmt.Errorf("failed to update infrastructure: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) DeleteInfrastructure(infraID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM infrastructure_components WHERE id = ?", infraID)
	if err != nil {
		return fmt.Errorf("failed to delete infrastructure: %w", err)
	}

	return nil
}

// Pipeline management operations
func (s *SQLiteStorage) CreatePipeline(pipeline *CollectorPipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO collector_pipelines 
		 (id, plan_id, collector_id, name, config_yaml, rules, status, target_type, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pipeline.ID, pipeline.PlanID, pipeline.CollectorID, pipeline.Name, pipeline.ConfigYAML,
		pipeline.Rules, pipeline.Status, pipeline.TargetType, pipeline.CreatedAt, pipeline.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert pipeline: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetPipeline(pipelineID string) (*CollectorPipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, plan_id, collector_id, name, config_yaml, rules, status, target_type, created_at, updated_at
		 FROM collector_pipelines WHERE id = ?`,
		pipelineID)

	var pipeline CollectorPipeline
	err := row.Scan(
		&pipeline.ID, &pipeline.PlanID, &pipeline.CollectorID, &pipeline.Name, &pipeline.ConfigYAML,
		&pipeline.Rules, &pipeline.Status, &pipeline.TargetType, &pipeline.CreatedAt, &pipeline.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pipeline with ID %s not found", pipelineID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan pipeline: %w", err)
	}

	return &pipeline, nil
}

func (s *SQLiteStorage) GetPipelinesByPlan(planID string) ([]*CollectorPipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, collector_id, name, config_yaml, rules, status, target_type, created_at, updated_at
		 FROM collector_pipelines WHERE plan_id = ? ORDER BY created_at ASC`,
		planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pipelines: %w", err)
	}
	defer rows.Close()

	pipelines := []*CollectorPipeline{}
	for rows.Next() {
		var pipeline CollectorPipeline
		err := rows.Scan(
			&pipeline.ID, &pipeline.PlanID, &pipeline.CollectorID, &pipeline.Name, &pipeline.ConfigYAML,
			&pipeline.Rules, &pipeline.Status, &pipeline.TargetType, &pipeline.CreatedAt, &pipeline.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan pipeline: %w", err)
		}
		pipelines = append(pipelines, &pipeline)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate pipelines: %w", err)
	}

	return pipelines, nil
}

func (s *SQLiteStorage) UpdatePipeline(pipelineID string, pipeline *CollectorPipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`UPDATE collector_pipelines SET 
		 plan_id = ?, collector_id = ?, name = ?, config_yaml = ?, rules = ?, 
		 status = ?, target_type = ?, updated_at = ?
		 WHERE id = ?`,
		pipeline.PlanID, pipeline.CollectorID, pipeline.Name, pipeline.ConfigYAML, pipeline.Rules,
		pipeline.Status, pipeline.TargetType, time.Now(), pipelineID)

	if err != nil {
		return fmt.Errorf("failed to update pipeline: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) DeletePipeline(pipelineID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM collector_pipelines WHERE id = ?", pipelineID)
	if err != nil {
		return fmt.Errorf("failed to delete pipeline: %w", err)
	}

	return nil
}

// Backend management operations
func (s *SQLiteStorage) CreateBackend(backend *Backend) error {
	if backend == nil {
		return fmt.Errorf("backend cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var planID sql.NullString
	if backend.PlanID != nil && *backend.PlanID != "" {
		planID.String = *backend.PlanID
		planID.Valid = true
	}

	_, err := s.db.Exec(
		`INSERT INTO backends 
		 (id, plan_id, backend_type, name, url, credentials, health_status, last_check, datasource_uid, config, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		backend.ID, planID, backend.BackendType, backend.Name, backend.URL, backend.Credentials,
		backend.HealthStatus, backend.LastCheck, backend.DatasourceUID, backend.Config, backend.CreatedAt, backend.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert backend: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetBackend(backendID string) (*Backend, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, plan_id, backend_type, name, url, credentials, health_status, last_check, datasource_uid, config, created_at, updated_at
		 FROM backends WHERE id = ?`,
		backendID)

	var backend Backend
	var planID sql.NullString
	var lastCheck sql.NullTime
	err := row.Scan(
		&backend.ID, &planID, &backend.BackendType, &backend.Name, &backend.URL, &backend.Credentials,
		&backend.HealthStatus, &lastCheck, &backend.DatasourceUID, &backend.Config, &backend.CreatedAt, &backend.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("backend with ID %s not found", backendID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan backend: %w", err)
	}

	if planID.Valid && planID.String != "" {
		backend.PlanID = &planID.String
	}

	if lastCheck.Valid {
		backend.LastCheck = &lastCheck.Time
	}

	return &backend, nil
}

func (s *SQLiteStorage) GetBackendsByPlan(planID string) ([]*Backend, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, backend_type, name, url, credentials, health_status, last_check, datasource_uid, config, created_at, updated_at
		 FROM backends WHERE plan_id = ? ORDER BY created_at ASC`,
		planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query backends: %w", err)
	}
	defer rows.Close()

	backends := []*Backend{}
	for rows.Next() {
		var backend Backend
		var planIDVal sql.NullString
		var lastCheck sql.NullTime
		err := rows.Scan(
			&backend.ID, &planIDVal, &backend.BackendType, &backend.Name, &backend.URL, &backend.Credentials,
			&backend.HealthStatus, &lastCheck, &backend.DatasourceUID, &backend.Config, &backend.CreatedAt, &backend.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan backend: %w", err)
		}

		if planIDVal.Valid && planIDVal.String != "" {
			backend.PlanID = &planIDVal.String
		}

		if lastCheck.Valid {
			backend.LastCheck = &lastCheck.Time
		}

		backends = append(backends, &backend)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate backends: %w", err)
	}

	return backends, nil
}

func (s *SQLiteStorage) ListAllBackends() ([]*Backend, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, backend_type, name, url, credentials, health_status, last_check, datasource_uid, config, created_at, updated_at
		 FROM backends ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query backends: %w", err)
	}
	defer rows.Close()

	backends := []*Backend{}
	for rows.Next() {
		var backend Backend
		var planIDVal sql.NullString
		var lastCheck sql.NullTime
		err := rows.Scan(
			&backend.ID, &planIDVal, &backend.BackendType, &backend.Name, &backend.URL, &backend.Credentials,
			&backend.HealthStatus, &lastCheck, &backend.DatasourceUID, &backend.Config, &backend.CreatedAt, &backend.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan backend: %w", err)
		}

		if planIDVal.Valid && planIDVal.String != "" {
			backend.PlanID = &planIDVal.String
		}

		if lastCheck.Valid {
			backend.LastCheck = &lastCheck.Time
		}

		backends = append(backends, &backend)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate backends: %w", err)
	}

	return backends, nil
}

func (s *SQLiteStorage) UpdateBackend(backendID string, backend *Backend) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var planID sql.NullString
	if backend.PlanID != nil && *backend.PlanID != "" {
		planID.String = *backend.PlanID
		planID.Valid = true
	}

	_, err := s.db.Exec(
		`UPDATE backends SET 
		 plan_id = ?, backend_type = ?, name = ?, url = ?, credentials = ?, health_status = ?, 
		 last_check = ?, datasource_uid = ?, config = ?, updated_at = ?
		 WHERE id = ?`,
		planID, backend.BackendType, backend.Name, backend.URL, backend.Credentials,
		backend.HealthStatus, backend.LastCheck, backend.DatasourceUID, backend.Config, time.Now(), backendID)

	if err != nil {
		return fmt.Errorf("failed to update backend: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) DeleteBackend(backendID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM backends WHERE id = ?", backendID)
	if err != nil {
		return fmt.Errorf("failed to delete backend: %w", err)
	}

	return nil
}

// Dependency management operations
func (s *SQLiteStorage) CreateDependency(dep *PlanDependency) error {
	if dep == nil {
		return fmt.Errorf("dependency cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO plan_dependencies 
		 (id, plan_id, source_id, source_type, target_id, target_type, dependency_type, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		dep.ID, dep.PlanID, dep.SourceID, dep.SourceType, dep.TargetID, dep.TargetType, dep.DependencyType, dep.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert dependency: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetDependenciesByPlan(planID string) ([]*PlanDependency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, source_id, source_type, target_id, target_type, dependency_type, created_at
		 FROM plan_dependencies WHERE plan_id = ? ORDER BY created_at ASC`,
		planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	dependencies := []*PlanDependency{}
	for rows.Next() {
		var dep PlanDependency
		err := rows.Scan(
			&dep.ID, &dep.PlanID, &dep.SourceID, &dep.SourceType, &dep.TargetID, &dep.TargetType, &dep.DependencyType, &dep.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, &dep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dependencies: %w", err)
	}

	return dependencies, nil
}

func (s *SQLiteStorage) GetDependenciesBySource(sourceID string) ([]*PlanDependency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, source_id, source_type, target_id, target_type, dependency_type, created_at
		 FROM plan_dependencies WHERE source_id = ? ORDER BY created_at ASC`,
		sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	dependencies := []*PlanDependency{}
	for rows.Next() {
		var dep PlanDependency
		err := rows.Scan(
			&dep.ID, &dep.PlanID, &dep.SourceID, &dep.SourceType, &dep.TargetID, &dep.TargetType, &dep.DependencyType, &dep.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, &dep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dependencies: %w", err)
	}

	return dependencies, nil
}

func (s *SQLiteStorage) GetDependenciesByTarget(targetID string) ([]*PlanDependency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, plan_id, source_id, source_type, target_id, target_type, dependency_type, created_at
		 FROM plan_dependencies WHERE target_id = ? ORDER BY created_at ASC`,
		targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	dependencies := []*PlanDependency{}
	for rows.Next() {
		var dep PlanDependency
		err := rows.Scan(
			&dep.ID, &dep.PlanID, &dep.SourceID, &dep.SourceType, &dep.TargetID, &dep.TargetType, &dep.DependencyType, &dep.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, &dep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dependencies: %w", err)
	}

	return dependencies, nil
}

func (s *SQLiteStorage) DeleteDependency(depID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM plan_dependencies WHERE id = ?", depID)
	if err != nil {
		return fmt.Errorf("failed to delete dependency: %w", err)
	}

	return nil
}

// Helper function
func formatUpdateClause(updates []string) string {
	if len(updates) == 0 {
		return ""
	}
	clause := updates[0]
	for i := 1; i < len(updates); i++ {
		clause += ", " + updates[i]
	}
	return clause
}

