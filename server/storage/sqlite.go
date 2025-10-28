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
	backendsTable := `
	CREATE TABLE IF NOT EXISTS backends (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
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

