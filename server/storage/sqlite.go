package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage is a SQLite implementation of Storage
type SQLiteStorage struct {
	db      *sql.DB
	emitter events.EventEmitter
	mu      sync.RWMutex
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Initialize database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	storage := &SQLiteStorage{
		db:      db,
		emitter: events.NewEmitter(),
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the database schema
func (s *SQLiteStorage) initSchema() error {
	// Create runs table
	runsTable := `
	CREATE TABLE IF NOT EXISTS runs (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		agent_name TEXT NOT NULL,
		status TEXT NOT NULL,
		prompt TEXT NOT NULL,
		model TEXT NOT NULL,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP,
		duration TEXT,
		total_iterations INTEGER NOT NULL,
		total_tool_calls INTEGER NOT NULL,
		total_tokens TEXT NOT NULL,
		error TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create events table
	eventsTable := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		run_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		agent_name TEXT NOT NULL,
		event_type TEXT NOT NULL,
		data TEXT NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
	);`

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);",
		"CREATE INDEX IF NOT EXISTS idx_runs_start_time ON runs(start_time);",
		"CREATE INDEX IF NOT EXISTS idx_events_run_id ON events(run_id);",
		"CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);",
	}

	if _, err := s.db.Exec(runsTable); err != nil {
		return fmt.Errorf("failed to create runs table: %w", err)
	}

	if _, err := s.db.Exec(eventsTable); err != nil {
		return fmt.Errorf("failed to create events table: %w", err)
	}

	for _, index := range indexes {
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Run migrations to add handoff support
	if err := s.migrateSchema(); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}

	// Create plan tables
	if err := s.initPlanSchema(); err != nil {
		return fmt.Errorf("failed to initialize plan schema: %w", err)
	}

	// Create agent work tables
	if err := s.initAgentWorkSchema(); err != nil {
		return fmt.Errorf("failed to initialize agent work schema: %w", err)
	}

	// Create human action tables
	if err := s.initHumanActionSchema(); err != nil {
		return fmt.Errorf("failed to initialize human action schema: %w", err)
	}

	// Create custom agent tables
	if err := s.initCustomAgentSchema(); err != nil {
		return fmt.Errorf("failed to initialize custom agent schema: %w", err)
	}

	return nil
}

// initPlanSchema creates tables for observability plan management
func (s *SQLiteStorage) initPlanSchema() error {
	// Create observability_plans table
	plansTable := `
	CREATE TABLE IF NOT EXISTS observability_plans (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		environment TEXT,
		status TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create instrumented_services table
	servicesTable := `
	CREATE TABLE IF NOT EXISTS instrumented_services (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		service_name TEXT NOT NULL,
		language TEXT,
		framework TEXT,
		sdk_version TEXT,
		config_file TEXT,
		status TEXT,
		code_changes_summary TEXT,
		target_path TEXT,
		exporter_endpoint TEXT,
		git_repo_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plan_id) REFERENCES observability_plans(id) ON DELETE CASCADE
	);`

	// Create infrastructure_components table
	infrastructureTable := `
	CREATE TABLE IF NOT EXISTS infrastructure_components (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		component_type TEXT,
		name TEXT NOT NULL,
		host TEXT,
		receiver_type TEXT,
		metrics_collected TEXT,
		status TEXT,
		config TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plan_id) REFERENCES observability_plans(id) ON DELETE CASCADE
	);`

	// Create collector_pipelines table
	pipelinesTable := `
	CREATE TABLE IF NOT EXISTS collector_pipelines (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		collector_id TEXT,
		name TEXT NOT NULL,
		config_yaml TEXT,
		rules TEXT,
		status TEXT,
		target_type TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plan_id) REFERENCES observability_plans(id) ON DELETE CASCADE
	);`

	// Create backends table
	// Note: plan_id is nullable to support standalone backends
	backendsTable := `
	CREATE TABLE IF NOT EXISTS backends (
		id TEXT PRIMARY KEY,
		plan_id TEXT,
		backend_type TEXT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		credentials TEXT,
		health_status TEXT,
		last_check TIMESTAMP,
		datasource_uid TEXT,
		config TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plan_id) REFERENCES observability_plans(id) ON DELETE CASCADE
	);`

	// Create plan_dependencies table
	dependenciesTable := `
	CREATE TABLE IF NOT EXISTS plan_dependencies (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		source_id TEXT NOT NULL,
		source_type TEXT,
		target_id TEXT NOT NULL,
		target_type TEXT,
		dependency_type TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plan_id) REFERENCES observability_plans(id) ON DELETE CASCADE
	);`

	// Create indexes for plan tables
	planIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_plans_status ON observability_plans(status);",
		"CREATE INDEX IF NOT EXISTS idx_services_plan_id ON instrumented_services(plan_id);",
		"CREATE INDEX IF NOT EXISTS idx_services_status ON instrumented_services(status);",
		"CREATE INDEX IF NOT EXISTS idx_infrastructure_plan_id ON infrastructure_components(plan_id);",
		"CREATE INDEX IF NOT EXISTS idx_pipelines_plan_id ON collector_pipelines(plan_id);",
		"CREATE INDEX IF NOT EXISTS idx_backends_plan_id ON backends(plan_id);",
		"CREATE INDEX IF NOT EXISTS idx_dependencies_plan_id ON plan_dependencies(plan_id);",
		"CREATE INDEX IF NOT EXISTS idx_dependencies_source ON plan_dependencies(source_id);",
		"CREATE INDEX IF NOT EXISTS idx_dependencies_target ON plan_dependencies(target_id);",
	}

	// Execute table creation
	tables := []string{plansTable, servicesTable, infrastructureTable, pipelinesTable, backendsTable, dependenciesTable}
	for _, table := range tables {
		if _, err := s.db.Exec(table); err != nil {
			return fmt.Errorf("failed to create plan table: %w", err)
		}
	}

	// Execute index creation
	for _, index := range planIndexes {
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create plan index: %w", err)
		}
	}

	return nil
}

// initAgentWorkSchema creates tables for agent work tracking
func (s *SQLiteStorage) initAgentWorkSchema() error {
	agentWorkTable := `
	CREATE TABLE IF NOT EXISTS agent_work (
		id TEXT PRIMARY KEY,
		resource_type TEXT NOT NULL,
		resource_id TEXT NOT NULL,
		run_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		agent_name TEXT NOT NULL,
		task_description TEXT NOT NULL,
		status TEXT NOT NULL,
		started_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		error TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
	);`

	agentWorkIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_agent_work_resource ON agent_work(resource_type, resource_id);",
		"CREATE INDEX IF NOT EXISTS idx_agent_work_status ON agent_work(status);",
		"CREATE INDEX IF NOT EXISTS idx_agent_work_run_id ON agent_work(run_id);",
		"CREATE INDEX IF NOT EXISTS idx_agent_work_started_at ON agent_work(started_at);",
	}

	if _, err := s.db.Exec(agentWorkTable); err != nil {
		return fmt.Errorf("failed to create agent_work table: %w", err)
	}

	for _, index := range agentWorkIndexes {
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create agent_work index: %w", err)
		}
	}

	return nil
}

// initHumanActionSchema creates tables for human action tracking
func (s *SQLiteStorage) initHumanActionSchema() error {
	humanActionTable := `
	CREATE TABLE IF NOT EXISTS human_actions (
		id TEXT PRIMARY KEY,
		run_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		agent_name TEXT NOT NULL,
		resource_type TEXT,
		resource_id TEXT,
		agent_work_id TEXT,
		request_type TEXT NOT NULL,
		question TEXT NOT NULL,
		context TEXT NOT NULL,
		options TEXT,
		status TEXT NOT NULL,
		response TEXT,
		responded_at TIMESTAMP,
		resumed_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
	);`

	humanActionIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_human_actions_run_id ON human_actions(run_id);",
		"CREATE INDEX IF NOT EXISTS idx_human_actions_status ON human_actions(status);",
		"CREATE INDEX IF NOT EXISTS idx_human_actions_resource ON human_actions(resource_type, resource_id);",
		"CREATE INDEX IF NOT EXISTS idx_human_actions_created_at ON human_actions(created_at);",
	}

	if _, err := s.db.Exec(humanActionTable); err != nil {
		return fmt.Errorf("failed to create human_actions table: %w", err)
	}

	for _, index := range humanActionIndexes {
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create human_actions index: %w", err)
		}
	}

	return nil
}

// migrateSchema handles schema migrations
func (s *SQLiteStorage) migrateSchema() error {
	// Check if handoff columns exist
	var columnExists bool
	err := s.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pragma_table_info('runs') WHERE name = 'parent_run_id'
		)
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if !columnExists {
		// Add handoff columns
		migrations := []string{
			"ALTER TABLE runs ADD COLUMN parent_run_id TEXT;",
			"ALTER TABLE runs ADD COLUMN sub_run_ids TEXT;",
			"ALTER TABLE runs ADD COLUMN is_handoff BOOLEAN DEFAULT 0;",
			"CREATE INDEX IF NOT EXISTS idx_runs_parent_run_id ON runs(parent_run_id);",
		}

		for _, migration := range migrations {
			if _, err := s.db.Exec(migration); err != nil {
				return fmt.Errorf("failed to run migration: %w", err)
			}
		}
	}

	// Check if backends table exists and if plan_id is NOT NULL (needs migration)
	var planIDNotNullable bool
	var backendsTableExists bool
	err = s.db.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='backends')
	`).Scan(&backendsTableExists)

	if err == nil && backendsTableExists {
		err = s.db.QueryRow(`
			SELECT "notnull" FROM pragma_table_info('backends') WHERE name = 'plan_id'
		`).Scan(&planIDNotNullable)

		if err == nil && planIDNotNullable {
			// Need to migrate: make plan_id nullable
			log.Printf("Migrating backends table to make plan_id nullable...")
			
			// Create new table with nullable plan_id
			_, err = s.db.Exec(`
				CREATE TABLE IF NOT EXISTS backends_new (
					id TEXT PRIMARY KEY,
					plan_id TEXT,
					backend_type TEXT,
					name TEXT NOT NULL,
					url TEXT NOT NULL,
					credentials TEXT,
					health_status TEXT,
					last_check TIMESTAMP,
					datasource_uid TEXT,
					config TEXT,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (plan_id) REFERENCES observability_plans(id) ON DELETE CASCADE
				);
			`)
			if err != nil {
				log.Printf("Warning: Failed to create new backends table: %v", err)
				return nil
			}

			// Copy data from old table to new table
			_, err = s.db.Exec(`
				INSERT INTO backends_new 
				SELECT * FROM backends;
			`)
			if err != nil {
				log.Printf("Warning: Failed to copy data to new backends table: %v", err)
				// Drop the new table if copy failed
				s.db.Exec("DROP TABLE IF EXISTS backends_new;")
				return nil
			}

			// Drop old table
			_, err = s.db.Exec(`DROP TABLE backends;`)
			if err != nil {
				log.Printf("Warning: Failed to drop old backends table: %v", err)
				return nil
			}

			// Rename new table
			_, err = s.db.Exec(`ALTER TABLE backends_new RENAME TO backends;`)
			if err != nil {
				log.Printf("Warning: Failed to rename new backends table: %v", err)
				return nil
			}

			// Recreate indexes
			s.db.Exec("CREATE INDEX IF NOT EXISTS idx_backends_plan_id ON backends(plan_id);")
			log.Printf("Successfully migrated backends table to support nullable plan_id")
		}
	}

	// Check if git_repo_url column exists in instrumented_services
	var gitRepoURLExists bool
	err = s.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pragma_table_info('instrumented_services') WHERE name = 'git_repo_url'
		)
	`).Scan(&gitRepoURLExists)

	if err == nil && !gitRepoURLExists {
		log.Printf("Adding git_repo_url column to instrumented_services table...")
		_, err = s.db.Exec("ALTER TABLE instrumented_services ADD COLUMN git_repo_url TEXT;")
		if err != nil {
			log.Printf("Warning: Failed to add git_repo_url column: %v", err)
		} else {
			log.Printf("Successfully added git_repo_url column to instrumented_services table")
		}
	}

	return nil
}

// CreateRun creates a new run
func (s *SQLiteStorage) CreateRun(run *Run) error {
	if run == nil {
		return fmt.Errorf("run cannot be nil")
	}
	if run.ID == "" {
		return fmt.Errorf("run ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Serialize token usage
	tokenUsageJSON, err := json.Marshal(run.TotalTokens)
	if err != nil {
		return fmt.Errorf("failed to marshal token usage: %w", err)
	}

	var duration sql.NullString
	if run.Duration != "" {
		duration.String = run.Duration
		duration.Valid = true
	}

	var endTime sql.NullTime
	if run.EndTime != nil {
		endTime.Time = *run.EndTime
		endTime.Valid = true
	}

	// Serialize handoff fields
	var parentRunID sql.NullString
	if run.ParentRunID != nil && *run.ParentRunID != "" {
		parentRunID.String = *run.ParentRunID
		parentRunID.Valid = true
	}

	var subRunIDsJSON sql.NullString
	if len(run.SubRunIDs) > 0 {
		subRunIDsJSONBytes, err := json.Marshal(run.SubRunIDs)
		if err == nil {
			subRunIDsJSON.String = string(subRunIDsJSONBytes)
			subRunIDsJSON.Valid = true
		}
	}

	_, err = s.db.Exec(
		`INSERT INTO runs (id, agent_id, agent_name, status, prompt, model, start_time, end_time, duration, 
			total_iterations, total_tool_calls, total_tokens, error, parent_run_id, sub_run_ids, is_handoff) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.AgentID, run.AgentName, run.Status, run.Prompt, run.Model, 
		run.StartTime, endTime, duration, run.TotalIterations, run.TotalToolCalls, 
		tokenUsageJSON, run.Error, parentRunID, subRunIDsJSON, run.IsHandoff)

	if err != nil {
		return fmt.Errorf("failed to insert run: %w", err)
	}

	return nil
}

// GetRun retrieves a run by ID
func (s *SQLiteStorage) GetRun(runID string) (*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, agent_id, agent_name, status, prompt, model, start_time, end_time, duration,
			total_iterations, total_tool_calls, total_tokens, error, parent_run_id, sub_run_ids, is_handoff
		 FROM runs WHERE id = ?`,
		runID)

	var run Run
	var endTime sql.NullTime
	var duration sql.NullString
	var tokenUsageJSON string
	var parentRunID sql.NullString
	var subRunIDsJSON sql.NullString

	err := row.Scan(
		&run.ID, &run.AgentID, &run.AgentName, &run.Status, &run.Prompt, &run.Model,
		&run.StartTime, &endTime, &duration, &run.TotalIterations, &run.TotalToolCalls,
		&tokenUsageJSON, &run.Error, &parentRunID, &subRunIDsJSON, &run.IsHandoff)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("run with ID %s not found", runID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan run: %w", err)
	}

	if endTime.Valid {
		run.EndTime = &endTime.Time
	}
	if duration.Valid {
		run.Duration = duration.String
	}

	if err := json.Unmarshal([]byte(tokenUsageJSON), &run.TotalTokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token usage: %w", err)
	}

	// Unmarshal parent_run_id
	if parentRunID.Valid && parentRunID.String != "" {
		run.ParentRunID = &parentRunID.String
	}

	// Unmarshal sub_run_ids
	if subRunIDsJSON.Valid && subRunIDsJSON.String != "" {
		if err := json.Unmarshal([]byte(subRunIDsJSON.String), &run.SubRunIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sub run IDs: %w", err)
		}
	}

	return &run, nil
}

// ListRuns retrieves runs with optional filtering
func (s *SQLiteStorage) ListRuns(opts RunListOptions) ([]*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, agent_id, agent_name, status, prompt, model, start_time, end_time, duration,
			  total_iterations, total_tool_calls, total_tokens, error
			  FROM runs WHERE 1=1`
	args := []interface{}{}

	// Apply filters
	if opts.Status != nil {
		query += " AND status = ?"
		args = append(args, *opts.Status)
	}
	if opts.Since != nil {
		query += " AND start_time >= ?"
		args = append(args, *opts.Since)
	}

	// Sort by start time (newest first)
	query += " ORDER BY start_time DESC"

	// Apply pagination
	if opts.Limit == 0 {
		opts.Limit = 100
	}
	query += " LIMIT ?"
	args = append(args, opts.Limit)

	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	runs := []*Run{}
	for rows.Next() {
		var run Run
		var endTime sql.NullTime
		var duration sql.NullString
		var tokenUsageJSON string

		err := rows.Scan(
			&run.ID, &run.AgentID, &run.AgentName, &run.Status, &run.Prompt, &run.Model,
			&run.StartTime, &endTime, &duration, &run.TotalIterations, &run.TotalToolCalls,
			&tokenUsageJSON, &run.Error)

		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		if endTime.Valid {
			run.EndTime = &endTime.Time
		}
		if duration.Valid {
			run.Duration = duration.String
		}

		if err := json.Unmarshal([]byte(tokenUsageJSON), &run.TotalTokens); err != nil {
			return nil, fmt.Errorf("failed to unmarshal token usage: %w", err)
		}

		runs = append(runs, &run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate runs: %w", err)
	}

	return runs, nil
}

// UpdateRun updates a run
func (s *SQLiteStorage) UpdateRun(runID string, update *RunUpdate) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify run exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM runs WHERE id = ?)", runID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if run exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("run with ID %s not found", runID)
	}

	// Build update query dynamically based on provided fields
	updates := []string{}
	args := []interface{}{}

	if update.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *update.Status)
	}
	if update.ClearEndTime {
		updates = append(updates, "end_time = NULL")
	} else if update.EndTime != nil {
		updates = append(updates, "end_time = ?")
		args = append(args, *update.EndTime)
	}
	if update.Duration != nil {
		updates = append(updates, "duration = ?")
		args = append(args, *update.Duration)
	}
	if update.TotalIterations != nil {
		updates = append(updates, "total_iterations = ?")
		args = append(args, *update.TotalIterations)
	}
	if update.TotalToolCalls != nil {
		updates = append(updates, "total_tool_calls = ?")
		args = append(args, *update.TotalToolCalls)
	}
	if update.TotalTokens != nil {
		tokenUsageJSON, err := json.Marshal(update.TotalTokens)
		if err != nil {
			return fmt.Errorf("failed to marshal token usage: %w", err)
		}
		updates = append(updates, "total_tokens = ?")
		args = append(args, tokenUsageJSON)
	}
	if update.Error != nil {
		updates = append(updates, "error = ?")
		args = append(args, *update.Error)
	}
	if update.ParentRunID != nil {
		updates = append(updates, "parent_run_id = ?")
		args = append(args, *update.ParentRunID)
	}
	if update.SubRunIDs != nil {
		subRunIDsJSON, err := json.Marshal(*update.SubRunIDs)
		if err == nil {
			updates = append(updates, "sub_run_ids = ?")
			args = append(args, string(subRunIDsJSON))
		}
	}
	if update.IsHandoff != nil {
		updates = append(updates, "is_handoff = ?")
		args = append(args, *update.IsHandoff)
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	args = append(args, runID)

	query := fmt.Sprintf("UPDATE runs SET %s WHERE id = ?", strings.Join(updates, ", "))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update run: %w", err)
	}

	return nil
}

// DeleteRun deletes a run and its events
func (s *SQLiteStorage) DeleteRun(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM runs WHERE id = ?", runID)
	if err != nil {
		return fmt.Errorf("failed to delete run: %w", err)
	}

	// Events are automatically deleted due to CASCADE

	return nil
}

// AddEvent adds an event to a run
func (s *SQLiteStorage) AddEvent(runID string, event *events.AgentEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify run exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM runs WHERE id = ?)", runID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if run exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("run with ID %s not found", runID)
	}

	// Serialize event data
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO events (id, run_id, agent_id, agent_name, event_type, data, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID, runID, event.AgentID, event.AgentName, event.Type, dataJSON, event.Timestamp)

	if err != nil {
		// Check if this is a duplicate key error
		// In SQLite, duplicate inserts return "UNIQUE constraint failed" error
		if strings.Contains(err.Error(), "UNIQUE constraint") || strings.Contains(err.Error(), "unique constraint") {
			// Event already exists, don't emit it again
			return nil
		}
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Emit event to subscribers (only if this was a successful new insert)
	s.emitter.Emit(event)

	return nil
}

// GetEvents retrieves events for a run
func (s *SQLiteStorage) GetEvents(runID string, after *time.Time) ([]*events.AgentEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// If we can't get a read lock or db is nil, return empty slice instead of error
	if s.db == nil {
		return []*events.AgentEvent{}, nil
	}

	query := `SELECT id, agent_id, agent_name, run_id, event_type, data, timestamp
			  FROM events WHERE run_id = ?`
	args := []interface{}{runID}

	if after != nil {
		query += " AND timestamp > ?"
		args = append(args, *after)
	}

	query += " ORDER BY timestamp ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		// Return empty slice instead of error for non-existent runs
		return []*events.AgentEvent{}, nil
	}
	defer rows.Close()

	agentEvents := []*events.AgentEvent{}
	for rows.Next() {
		var event events.AgentEvent
		var dataJSON string

		err := rows.Scan(
			&event.ID, &event.AgentID, &event.AgentName, &event.RunID, &event.Type, &dataJSON, &event.Timestamp)

		if err != nil {
			log.Printf("Failed to scan event: %v", err)
			continue // Skip invalid events instead of returning error
		}

		event.Data = json.RawMessage(dataJSON)
		agentEvents = append(agentEvents, &event)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating events: %v", err)
		// Return what we have instead of erroring
	}

	return agentEvents, nil
}

// GetEventCount returns the number of events for a run
func (s *SQLiteStorage) GetEventCount(runID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM events WHERE run_id = ?", runID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// GetSubRuns returns all sub-runs for a parent run
func (s *SQLiteStorage) GetSubRuns(parentRunID string) ([]*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, agent_id, agent_name, status, prompt, model, start_time, end_time, duration,
			  total_iterations, total_tool_calls, total_tokens, error, parent_run_id, sub_run_ids, is_handoff
			  FROM runs WHERE parent_run_id = ? ORDER BY start_time ASC`

	rows, err := s.db.Query(query, parentRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sub-runs: %w", err)
	}
	defer rows.Close()

	runs := []*Run{}
	for rows.Next() {
		var run Run
		var endTime sql.NullTime
		var duration sql.NullString
		var tokenUsageJSON string
		var parentRunID sql.NullString
		var subRunIDsJSON sql.NullString

		err := rows.Scan(
			&run.ID, &run.AgentID, &run.AgentName, &run.Status, &run.Prompt, &run.Model,
			&run.StartTime, &endTime, &duration, &run.TotalIterations, &run.TotalToolCalls,
			&tokenUsageJSON, &run.Error, &parentRunID, &subRunIDsJSON, &run.IsHandoff)

		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		if endTime.Valid {
			run.EndTime = &endTime.Time
		}
		if duration.Valid {
			run.Duration = duration.String
		}

		if err := json.Unmarshal([]byte(tokenUsageJSON), &run.TotalTokens); err != nil {
			return nil, fmt.Errorf("failed to unmarshal token usage: %w", err)
		}

		// Unmarshal parent_run_id
		if parentRunID.Valid && parentRunID.String != "" {
			run.ParentRunID = &parentRunID.String
		}

		// Unmarshal sub_run_ids
		if subRunIDsJSON.Valid && subRunIDsJSON.String != "" {
			if err := json.Unmarshal([]byte(subRunIDsJSON.String), &run.SubRunIDs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal sub run IDs: %w", err)
			}
		}

		runs = append(runs, &run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate runs: %w", err)
	}

	return runs, nil
}

// GetParentRun returns the parent run for a sub-run
func (s *SQLiteStorage) GetParentRun(subRunID string) (*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, agent_id, agent_name, status, prompt, model, start_time, end_time, duration,
			total_iterations, total_tool_calls, total_tokens, error, parent_run_id, sub_run_ids, is_handoff
		 FROM runs WHERE id = (
			 SELECT parent_run_id FROM runs WHERE id = ?
		 )`,
		subRunID)

	var run Run
	var endTime sql.NullTime
	var duration sql.NullString
	var tokenUsageJSON string
	var parentRunID sql.NullString
	var subRunIDsJSON sql.NullString

	err := row.Scan(
		&run.ID, &run.AgentID, &run.AgentName, &run.Status, &run.Prompt, &run.Model,
		&run.StartTime, &endTime, &duration, &run.TotalIterations, &run.TotalToolCalls,
		&tokenUsageJSON, &run.Error, &parentRunID, &subRunIDsJSON, &run.IsHandoff)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("parent run not found for sub-run %s", subRunID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan run: %w", err)
	}

	if endTime.Valid {
		run.EndTime = &endTime.Time
	}
	if duration.Valid {
		run.Duration = duration.String
	}

	if err := json.Unmarshal([]byte(tokenUsageJSON), &run.TotalTokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token usage: %w", err)
	}

	// Unmarshal parent_run_id
	if parentRunID.Valid && parentRunID.String != "" {
		run.ParentRunID = &parentRunID.String
	}

	// Unmarshal sub_run_ids
	if subRunIDsJSON.Valid && subRunIDsJSON.String != "" {
		if err := json.Unmarshal([]byte(subRunIDsJSON.String), &run.SubRunIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sub run IDs: %w", err)
		}
	}

	return &run, nil
}

// Subscribe creates a subscription to events for a specific run
func (s *SQLiteStorage) Subscribe(runID string) (<-chan *events.AgentEvent, func()) {
	return s.emitter.Subscribe(runID)
}

// SubscribeAll creates a subscription to all events
func (s *SQLiteStorage) SubscribeAll() (<-chan *events.AgentEvent, func()) {
	return s.emitter.SubscribeAll()
}

// Close closes the storage
func (s *SQLiteStorage) Close() error {
	s.emitter.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// CreateAgentWork creates a new agent work entry
func (s *SQLiteStorage) CreateAgentWork(work *AgentWork) error {
	if work == nil {
		return fmt.Errorf("agent work cannot be nil")
	}
	if work.ID == "" {
		return fmt.Errorf("agent work ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var completedAt sql.NullTime
	if work.CompletedAt != nil {
		completedAt.Time = *work.CompletedAt
		completedAt.Valid = true
	}

	var errorMsg sql.NullString
	if work.Error != "" {
		errorMsg.String = work.Error
		errorMsg.Valid = true
	}

	now := time.Now()
	work.CreatedAt = now
	work.UpdatedAt = now

	_, err := s.db.Exec(
		`INSERT INTO agent_work (id, resource_type, resource_id, run_id, agent_id, agent_name, 
		 task_description, status, started_at, completed_at, error, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		work.ID, work.ResourceType, work.ResourceID, work.RunID, work.AgentID, work.AgentName,
		work.TaskDescription, work.Status, work.StartedAt, completedAt, errorMsg,
		work.CreatedAt, work.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert agent work: %w", err)
	}

	return nil
}

// GetAgentWork retrieves agent work by ID
func (s *SQLiteStorage) GetAgentWork(workID string) (*AgentWork, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, resource_type, resource_id, run_id, agent_id, agent_name, task_description,
		 status, started_at, completed_at, error, created_at, updated_at
		 FROM agent_work WHERE id = ?`,
		workID)

	var work AgentWork
	var completedAt sql.NullTime
	var errorMsg sql.NullString

	err := row.Scan(
		&work.ID, &work.ResourceType, &work.ResourceID, &work.RunID, &work.AgentID,
		&work.AgentName, &work.TaskDescription, &work.Status, &work.StartedAt,
		&completedAt, &errorMsg, &work.CreatedAt, &work.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent work with ID %s not found", workID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan agent work: %w", err)
	}

	if completedAt.Valid {
		work.CompletedAt = &completedAt.Time
	}
	if errorMsg.Valid {
		work.Error = errorMsg.String
	}

	return &work, nil
}

// GetAgentWorkByResource retrieves agent work for a specific resource
func (s *SQLiteStorage) GetAgentWorkByResource(resourceType ResourceType, resourceID string) ([]*AgentWork, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, resource_type, resource_id, run_id, agent_id, agent_name, task_description,
		 status, started_at, completed_at, error, created_at, updated_at
		 FROM agent_work WHERE resource_type = ? AND resource_id = ?
		 ORDER BY started_at DESC`,
		resourceType, resourceID)

	if err != nil {
		return nil, fmt.Errorf("failed to query agent work: %w", err)
	}
	defer rows.Close()

	var works []*AgentWork
	for rows.Next() {
		var work AgentWork
		var completedAt sql.NullTime
		var errorMsg sql.NullString

		err := rows.Scan(
			&work.ID, &work.ResourceType, &work.ResourceID, &work.RunID, &work.AgentID,
			&work.AgentName, &work.TaskDescription, &work.Status, &work.StartedAt,
			&completedAt, &errorMsg, &work.CreatedAt, &work.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan agent work: %w", err)
		}

		if completedAt.Valid {
			work.CompletedAt = &completedAt.Time
		}
		if errorMsg.Valid {
			work.Error = errorMsg.String
		}

		works = append(works, &work)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate agent work: %w", err)
	}

	return works, nil
}

// ListAgentWork lists agent work with optional filtering
func (s *SQLiteStorage) ListAgentWork(opts AgentWorkListOptions) ([]*AgentWork, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, resource_type, resource_id, run_id, agent_id, agent_name, task_description,
			  status, started_at, completed_at, error, created_at, updated_at
			  FROM agent_work WHERE 1=1`
	args := []interface{}{}

	if opts.ResourceType != nil {
		query += " AND resource_type = ?"
		args = append(args, *opts.ResourceType)
	}
	if opts.ResourceID != nil {
		query += " AND resource_id = ?"
		args = append(args, *opts.ResourceID)
	}
	if opts.Status != nil {
		query += " AND status = ?"
		args = append(args, *opts.Status)
	}

	query += " ORDER BY started_at DESC"

	if opts.Limit == 0 {
		opts.Limit = 100
	}
	query += " LIMIT ?"
	args = append(args, opts.Limit)

	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent work: %w", err)
	}
	defer rows.Close()

	var works []*AgentWork
	for rows.Next() {
		var work AgentWork
		var completedAt sql.NullTime
		var errorMsg sql.NullString

		err := rows.Scan(
			&work.ID, &work.ResourceType, &work.ResourceID, &work.RunID, &work.AgentID,
			&work.AgentName, &work.TaskDescription, &work.Status, &work.StartedAt,
			&completedAt, &errorMsg, &work.CreatedAt, &work.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan agent work: %w", err)
		}

		if completedAt.Valid {
			work.CompletedAt = &completedAt.Time
		}
		if errorMsg.Valid {
			work.Error = errorMsg.String
		}

		works = append(works, &work)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate agent work: %w", err)
	}

	return works, nil
}

// UpdateAgentWork updates agent work
func (s *SQLiteStorage) UpdateAgentWork(workID string, update *AgentWorkUpdate) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify work exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM agent_work WHERE id = ?)", workID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if agent work exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("agent work with ID %s not found", workID)
	}

	updates := []string{}
	args := []interface{}{}

	if update.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *update.Status)
	}
	if update.CompletedAt != nil {
		updates = append(updates, "completed_at = ?")
		args = append(args, *update.CompletedAt)
	} else if update.Status != nil && (*update.Status == AgentWorkStatusCompleted || *update.Status == AgentWorkStatusFailed || *update.Status == AgentWorkStatusCancelled) {
		// Auto-set completed_at if status is terminal
		now := time.Now()
		updates = append(updates, "completed_at = ?")
		args = append(args, now)
	}
	if update.Error != nil {
		updates = append(updates, "error = ?")
		args = append(args, *update.Error)
	}

	if len(updates) == 0 {
		return nil
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, workID)

	query := fmt.Sprintf("UPDATE agent_work SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update agent work: %w", err)
	}

	return nil
}

// DeleteAgentWork deletes agent work
func (s *SQLiteStorage) DeleteAgentWork(workID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM agent_work WHERE id = ?", workID)
	if err != nil {
		return fmt.Errorf("failed to delete agent work: %w", err)
	}

	return nil
}

// CreateHumanAction creates a new human action entry
func (s *SQLiteStorage) CreateHumanAction(action *HumanAction) error {
	if action == nil {
		return fmt.Errorf("human action cannot be nil")
	}
	if action.ID == "" {
		return fmt.Errorf("human action ID cannot be empty")
	}
	if action.RunID == "" {
		return fmt.Errorf("run ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var resourceType sql.NullString
	if action.ResourceType != nil {
		resourceType.String = string(*action.ResourceType)
		resourceType.Valid = true
	}

	var resourceID sql.NullString
	if action.ResourceID != nil {
		resourceID.String = *action.ResourceID
		resourceID.Valid = true
	}

	var agentWorkID sql.NullString
	if action.AgentWorkID != nil {
		agentWorkID.String = *action.AgentWorkID
		agentWorkID.Valid = true
	}

	var optionsJSON sql.NullString
	if len(action.Options) > 0 {
		optionsBytes, err := json.Marshal(action.Options)
		if err == nil {
			optionsJSON.String = string(optionsBytes)
			optionsJSON.Valid = true
		}
	}

	var response sql.NullString
	if action.Response != nil {
		response.String = *action.Response
		response.Valid = true
	}

	var respondedAt sql.NullTime
	if action.RespondedAt != nil {
		respondedAt.Time = *action.RespondedAt
		respondedAt.Valid = true
	}

	var resumedAt sql.NullTime
	if action.ResumedAt != nil {
		resumedAt.Time = *action.ResumedAt
		resumedAt.Valid = true
	}

	now := time.Now()
	action.CreatedAt = now
	action.UpdatedAt = now

	_, err := s.db.Exec(
		`INSERT INTO human_actions (id, run_id, agent_id, agent_name, resource_type, resource_id, 
		 agent_work_id, request_type, question, context, options, status, response, responded_at, 
		 resumed_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		action.ID, action.RunID, action.AgentID, action.AgentName, resourceType, resourceID,
		agentWorkID, action.RequestType, action.Question, action.Context, optionsJSON,
		action.Status, response, respondedAt, resumedAt, action.CreatedAt, action.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert human action: %w", err)
	}

	return nil
}

// GetHumanAction retrieves a human action by ID
func (s *SQLiteStorage) GetHumanAction(actionID string) (*HumanAction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		`SELECT id, run_id, agent_id, agent_name, resource_type, resource_id, agent_work_id,
		 request_type, question, context, options, status, response, responded_at, resumed_at,
		 created_at, updated_at
		 FROM human_actions WHERE id = ?`,
		actionID)

	var action HumanAction
	var resourceType sql.NullString
	var resourceID sql.NullString
	var agentWorkID sql.NullString
	var optionsJSON sql.NullString
	var response sql.NullString
	var respondedAt sql.NullTime
	var resumedAt sql.NullTime

	err := row.Scan(
		&action.ID, &action.RunID, &action.AgentID, &action.AgentName,
		&resourceType, &resourceID, &agentWorkID,
		&action.RequestType, &action.Question, &action.Context, &optionsJSON,
		&action.Status, &response, &respondedAt, &resumedAt,
		&action.CreatedAt, &action.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("human action with ID %s not found", actionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan human action: %w", err)
	}

	if resourceType.Valid && resourceType.String != "" {
		rt := ResourceType(resourceType.String)
		action.ResourceType = &rt
	}
	if resourceID.Valid {
		action.ResourceID = &resourceID.String
	}
	if agentWorkID.Valid {
		action.AgentWorkID = &agentWorkID.String
	}
	if optionsJSON.Valid && optionsJSON.String != "" {
		if err := json.Unmarshal([]byte(optionsJSON.String), &action.Options); err != nil {
			return nil, fmt.Errorf("failed to unmarshal options: %w", err)
		}
	}
	if response.Valid {
		action.Response = &response.String
	}
	if respondedAt.Valid {
		action.RespondedAt = &respondedAt.Time
	}
	if resumedAt.Valid {
		action.ResumedAt = &resumedAt.Time
	}

	return &action, nil
}

// ListHumanActions lists human actions with optional filtering
func (s *SQLiteStorage) ListHumanActions(opts HumanActionListOptions) ([]*HumanAction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, run_id, agent_id, agent_name, resource_type, resource_id, agent_work_id,
			  request_type, question, context, options, status, response, responded_at, resumed_at,
			  created_at, updated_at
			  FROM human_actions WHERE 1=1`
	args := []interface{}{}

	if opts.RunID != nil {
		query += " AND run_id = ?"
		args = append(args, *opts.RunID)
	}
	if opts.Status != nil {
		query += " AND status = ?"
		args = append(args, *opts.Status)
	}
	if opts.RequestType != nil {
		query += " AND request_type = ?"
		args = append(args, *opts.RequestType)
	}

	query += " ORDER BY created_at DESC"

	if opts.Limit == 0 {
		opts.Limit = 100
	}
	query += " LIMIT ?"
	args = append(args, opts.Limit)

	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query human actions: %w", err)
	}
	defer rows.Close()

	actions := []*HumanAction{}
	for rows.Next() {
		var action HumanAction
		var resourceType sql.NullString
		var resourceID sql.NullString
		var agentWorkID sql.NullString
		var optionsJSON sql.NullString
		var response sql.NullString
		var respondedAt sql.NullTime
		var resumedAt sql.NullTime

		err := rows.Scan(
			&action.ID, &action.RunID, &action.AgentID, &action.AgentName,
			&resourceType, &resourceID, &agentWorkID,
			&action.RequestType, &action.Question, &action.Context, &optionsJSON,
			&action.Status, &response, &respondedAt, &resumedAt,
			&action.CreatedAt, &action.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan human action: %w", err)
		}

		if resourceType.Valid && resourceType.String != "" {
			rt := ResourceType(resourceType.String)
			action.ResourceType = &rt
		}
		if resourceID.Valid {
			action.ResourceID = &resourceID.String
		}
		if agentWorkID.Valid {
			action.AgentWorkID = &agentWorkID.String
		}
		if optionsJSON.Valid && optionsJSON.String != "" {
			if err := json.Unmarshal([]byte(optionsJSON.String), &action.Options); err == nil {
				// Ignore unmarshal errors for options
			}
		}
		if response.Valid {
			action.Response = &response.String
		}
		if respondedAt.Valid {
			action.RespondedAt = &respondedAt.Time
		}
		if resumedAt.Valid {
			action.ResumedAt = &resumedAt.Time
		}

		actions = append(actions, &action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate human actions: %w", err)
	}

	return actions, nil
}

// UpdateHumanAction updates a human action
func (s *SQLiteStorage) UpdateHumanAction(actionID string, update *HumanActionUpdate) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	updates := []string{}
	args := []interface{}{}

	if update.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *update.Status)
	}
	if update.Response != nil {
		updates = append(updates, "response = ?")
		args = append(args, *update.Response)
	}
	if update.RespondedAt != nil {
		updates = append(updates, "responded_at = ?")
		args = append(args, *update.RespondedAt)
	} else if update.Status != nil && *update.Status == HumanActionStatusResponded {
		// Auto-set responded_at if status is responded
		now := time.Now()
		updates = append(updates, "responded_at = ?")
		args = append(args, now)
	}
	if update.ResumedAt != nil {
		updates = append(updates, "resumed_at = ?")
		args = append(args, *update.ResumedAt)
	} else if update.Status != nil && *update.Status == HumanActionStatusResumed {
		// Auto-set resumed_at if status is resumed
		now := time.Now()
		updates = append(updates, "resumed_at = ?")
		args = append(args, now)
	}

	if len(updates) == 0 {
		return nil
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, actionID)

	query := fmt.Sprintf("UPDATE human_actions SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update human action: %w", err)
	}

	return nil
}

// DeleteHumanAction deletes a human action
func (s *SQLiteStorage) DeleteHumanAction(actionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM human_actions WHERE id = ?", actionID)
	if err != nil {
		return fmt.Errorf("failed to delete human action: %w", err)
	}

	return nil
}

// GetHumanActionsByRun retrieves all human actions for a specific run
func (s *SQLiteStorage) GetHumanActionsByRun(runID string) ([]*HumanAction, error) {
	opts := HumanActionListOptions{
		RunID:  &runID,
		Limit:  1000,
		Offset: 0,
	}
	return s.ListHumanActions(opts)
}

// GetPendingHumanActions retrieves all pending human actions
func (s *SQLiteStorage) GetPendingHumanActions() ([]*HumanAction, error) {
	status := HumanActionStatusPending
	opts := HumanActionListOptions{
		Status: &status,
		Limit:  1000,
		Offset: 0,
	}
	return s.ListHumanActions(opts)
}

// initCustomAgentSchema creates tables for custom agent management
func (s *SQLiteStorage) initCustomAgentSchema() error {
	customAgentTable := `
	CREATE TABLE IF NOT EXISTS custom_agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		system_prompt TEXT,
		model TEXT,
		max_tokens INTEGER,
		tool_names TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	customAgentIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_custom_agents_name ON custom_agents(name);",
		"CREATE INDEX IF NOT EXISTS idx_custom_agents_created_at ON custom_agents(created_at);",
	}

	if _, err := s.db.Exec(customAgentTable); err != nil {
		return fmt.Errorf("failed to create custom_agents table: %w", err)
	}

	for _, index := range customAgentIndexes {
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// CreateCustomAgent creates a new custom agent
func (s *SQLiteStorage) CreateCustomAgent(agent *CustomAgent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Serialize tool_names to JSON
	toolNamesJSON, err := json.Marshal(agent.ToolNames)
	if err != nil {
		return fmt.Errorf("failed to marshal tool names: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO custom_agents (id, name, description, system_prompt, model, max_tokens, tool_names, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		agent.ID, agent.Name, agent.Description, agent.SystemPrompt, agent.Model, agent.MaxTokens,
		string(toolNamesJSON), agent.CreatedAt, agent.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create custom agent: %w", err)
	}

	return nil
}

// GetCustomAgent retrieves a custom agent by ID
func (s *SQLiteStorage) GetCustomAgent(agentID string) (*CustomAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var agent CustomAgent
	var toolNamesJSON string
	var systemPrompt, model sql.NullString
	var maxTokens sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, name, description, system_prompt, model, max_tokens, tool_names, created_at, updated_at
		FROM custom_agents WHERE id = ?`, agentID).
		Scan(&agent.ID, &agent.Name, &agent.Description, &systemPrompt, &model, &maxTokens,
			&toolNamesJSON, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("custom agent not found")
		}
		return nil, fmt.Errorf("failed to get custom agent: %w", err)
	}

	if systemPrompt.Valid {
		agent.SystemPrompt = systemPrompt.String
	}
	if model.Valid {
		agent.Model = model.String
	}
	if maxTokens.Valid {
		agent.MaxTokens = maxTokens.Int64
	}

	// Deserialize tool_names from JSON
	if err := json.Unmarshal([]byte(toolNamesJSON), &agent.ToolNames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool names: %w", err)
	}

	return &agent, nil
}

// ListCustomAgents retrieves all custom agents
func (s *SQLiteStorage) ListCustomAgents() ([]*CustomAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, name, description, system_prompt, model, max_tokens, tool_names, created_at, updated_at
		FROM custom_agents ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom agents: %w", err)
	}
	defer rows.Close()

	var agents []*CustomAgent
	for rows.Next() {
		var agent CustomAgent
		var toolNamesJSON string
		var systemPrompt, model sql.NullString
		var maxTokens sql.NullInt64

		err := rows.Scan(&agent.ID, &agent.Name, &agent.Description, &systemPrompt, &model, &maxTokens,
			&toolNamesJSON, &agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan custom agent: %w", err)
		}

		if systemPrompt.Valid {
			agent.SystemPrompt = systemPrompt.String
		}
		if model.Valid {
			agent.Model = model.String
		}
		if maxTokens.Valid {
			agent.MaxTokens = maxTokens.Int64
		}

		// Deserialize tool_names from JSON
		if err := json.Unmarshal([]byte(toolNamesJSON), &agent.ToolNames); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool names: %w", err)
		}

		agents = append(agents, &agent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate custom agents: %w", err)
	}

	return agents, nil
}

// UpdateCustomAgent updates a custom agent
func (s *SQLiteStorage) UpdateCustomAgent(agentID string, update *CustomAgentUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Build dynamic update query
	updates := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}

	if update.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *update.Name)
	}
	if update.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *update.Description)
	}
	if update.SystemPrompt != nil {
		if *update.SystemPrompt == "" {
			updates = append(updates, "system_prompt = NULL")
		} else {
			updates = append(updates, "system_prompt = ?")
			args = append(args, *update.SystemPrompt)
		}
	}
	if update.Model != nil {
		if *update.Model == "" {
			updates = append(updates, "model = NULL")
		} else {
			updates = append(updates, "model = ?")
			args = append(args, *update.Model)
		}
	}
	if update.MaxTokens != nil {
		updates = append(updates, "max_tokens = ?")
		args = append(args, *update.MaxTokens)
	}
	if update.ToolNames != nil {
		toolNamesJSON, err := json.Marshal(*update.ToolNames)
		if err != nil {
			return fmt.Errorf("failed to marshal tool names: %w", err)
		}
		updates = append(updates, "tool_names = ?")
		args = append(args, string(toolNamesJSON))
	}

	if len(updates) == 1 {
		// Only updated_at, nothing to update
		return nil
	}

	query := fmt.Sprintf("UPDATE custom_agents SET %s WHERE id = ?", strings.Join(updates, ", "))
	args = append(args, agentID)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update custom agent: %w", err)
	}

	return nil
}

// DeleteCustomAgent deletes a custom agent
func (s *SQLiteStorage) DeleteCustomAgent(agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM custom_agents WHERE id = ?", agentID)
	if err != nil {
		return fmt.Errorf("failed to delete custom agent: %w", err)
	}

	return nil
}

// GetDBPath returns the default database path
func GetDBPath() string {
	// Try to get path from environment variable
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Default to current directory
		dbPath = "./otel-ai-engineer.db"
	}
	return dbPath
}

