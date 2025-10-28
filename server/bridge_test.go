package server

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// MockStorage is a minimal mock implementation of storage.Storage for testing
type MockStorage struct {
	runs   map[string]*storage.Run
	events map[string][]*events.AgentEvent
	mu     sync.RWMutex
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		runs:   make(map[string]*storage.Run),
		events: make(map[string][]*events.AgentEvent),
	}
}

func (m *MockStorage) CreateRun(run *storage.Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = run
	return nil
}

func (m *MockStorage) GetRun(runID string) (*storage.Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	run, exists := m.runs[runID]
	if !exists {
		return nil, fmt.Errorf("run not found: %s", runID)
	}
	return run, nil
}

func (m *MockStorage) UpdateRun(runID string, update *storage.RunUpdate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	run, exists := m.runs[runID]
	if !exists {
		return fmt.Errorf("run not found: %s", runID)
	}

	if update.Status != nil {
		run.Status = *update.Status
	}
	if update.EndTime != nil {
		run.EndTime = update.EndTime
	}
	if update.Duration != nil {
		run.Duration = *update.Duration
	}
	if update.TotalIterations != nil {
		run.TotalIterations = *update.TotalIterations
	}
	if update.TotalToolCalls != nil {
		run.TotalToolCalls = *update.TotalToolCalls
	}
	if update.TotalTokens != nil {
		run.TotalTokens = *update.TotalTokens
	}
	if update.Error != nil {
		run.Error = *update.Error
	}

	return nil
}

func (m *MockStorage) AddEvent(runID string, event *events.AgentEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[runID] = append(m.events[runID], event)
	return nil
}

func (m *MockStorage) GetEvents(runID string, after *time.Time) ([]*events.AgentEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.events[runID], nil
}

func (m *MockStorage) GetEventCount(runID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.events[runID]), nil
}

func (m *MockStorage) ListRuns(opts storage.RunListOptions) ([]*storage.Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	runs := make([]*storage.Run, 0, len(m.runs))
	for _, run := range m.runs {
		runs = append(runs, run)
	}
	return runs, nil
}

func (m *MockStorage) DeleteRun(runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.runs, runID)
	delete(m.events, runID)
	return nil
}

func (m *MockStorage) Subscribe(runID string) (<-chan *events.AgentEvent, func()) {
	ch := make(chan *events.AgentEvent)
	return ch, func() { close(ch) }
}

func (m *MockStorage) SubscribeAll() (<-chan *events.AgentEvent, func()) {
	ch := make(chan *events.AgentEvent)
	return ch, func() { close(ch) }
}

func (m *MockStorage) GetSubRuns(parentRunID string) ([]*storage.Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	subRuns := []*storage.Run{}
	for _, run := range m.runs {
		if run.ParentRunID != nil && *run.ParentRunID == parentRunID {
			subRuns = append(subRuns, run)
		}
	}
	return subRuns, nil
}

func (m *MockStorage) GetParentRun(subRunID string) (*storage.Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	subRun, exists := m.runs[subRunID]
	if !exists {
		return nil, fmt.Errorf("sub-run not found: %s", subRunID)
	}
	if subRun.ParentRunID == nil {
		return nil, fmt.Errorf("no parent run for sub-run %s", subRunID)
	}
	parentRun, exists := m.runs[*subRun.ParentRunID]
	if !exists {
		return nil, fmt.Errorf("parent run not found: %s", *subRun.ParentRunID)
	}
	return parentRun, nil
}

func (m *MockStorage) Close() error {
	return nil
}

// Plan management methods (stubs for testing)
func (m *MockStorage) CreatePlan(plan *storage.ObservabilityPlan) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetPlan(planID string) (*storage.ObservabilityPlan, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) ListPlans() ([]*storage.ObservabilityPlan, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) UpdatePlan(planID string, update *storage.PlanUpdate) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) DeletePlan(planID string) error {
	return fmt.Errorf("not implemented in MockStorage")
}

// Service management methods
func (m *MockStorage) CreateService(service *storage.InstrumentedService) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetService(serviceID string) (*storage.InstrumentedService, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetServicesByPlan(planID string) ([]*storage.InstrumentedService, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) UpdateService(serviceID string, service *storage.InstrumentedService) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) DeleteService(serviceID string) error {
	return fmt.Errorf("not implemented in MockStorage")
}

// Infrastructure management methods
func (m *MockStorage) CreateInfrastructure(infra *storage.InfrastructureComponent) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetInfrastructure(infraID string) (*storage.InfrastructureComponent, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetInfrastructureByPlan(planID string) ([]*storage.InfrastructureComponent, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) UpdateInfrastructure(infraID string, infra *storage.InfrastructureComponent) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) DeleteInfrastructure(infraID string) error {
	return fmt.Errorf("not implemented in MockStorage")
}

// Pipeline management methods
func (m *MockStorage) CreatePipeline(pipeline *storage.CollectorPipeline) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetPipeline(pipelineID string) (*storage.CollectorPipeline, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetPipelinesByPlan(planID string) ([]*storage.CollectorPipeline, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) UpdatePipeline(pipelineID string, pipeline *storage.CollectorPipeline) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) DeletePipeline(pipelineID string) error {
	return fmt.Errorf("not implemented in MockStorage")
}

// Backend management methods
func (m *MockStorage) CreateBackend(backend *storage.Backend) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetBackend(backendID string) (*storage.Backend, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetBackendsByPlan(planID string) ([]*storage.Backend, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) UpdateBackend(backendID string, backend *storage.Backend) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) DeleteBackend(backendID string) error {
	return fmt.Errorf("not implemented in MockStorage")
}

// Dependency management methods
func (m *MockStorage) CreateDependency(dep *storage.PlanDependency) error {
	return fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetDependenciesByPlan(planID string) ([]*storage.PlanDependency, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetDependenciesBySource(sourceID string) ([]*storage.PlanDependency, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) GetDependenciesByTarget(targetID string) ([]*storage.PlanDependency, error) {
	return nil, fmt.Errorf("not implemented in MockStorage")
}

func (m *MockStorage) DeleteDependency(depID string) error {
	return fmt.Errorf("not implemented in MockStorage")
}

// TestEventBridgeCreation verifies EventBridge is created correctly
func TestEventBridgeCreation(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()

	bridge := NewEventBridge(stor, emitter)

	if bridge == nil {
		t.Fatal("NewEventBridge returned nil")
	}

	if bridge.storage != stor {
		t.Error("EventBridge storage not set correctly")
	}

	if bridge.emitter != emitter {
		t.Error("EventBridge emitter not set correctly")
	}

	if bridge.runMap == nil {
		t.Error("EventBridge runMap not initialized")
	}

	// Cleanup
	bridge.Close()
}

// TestEventBridgeRunStartHandling verifies run start event handling
func TestEventBridgeRunStartHandling(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)
	defer bridge.Close()

	// Create run start event
	runID := "test-run-123"
	data := events.RunStartData{
		Prompt:       "Test prompt",
		Model:        "claude-3-5-sonnet",
		MaxTokens:    4096,
		SystemPrompt: "Test system prompt",
	}

	evt, err := events.NewRunStartEvent(runID, "test-agent", "Test Agent", data)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	// Emit the event
	emitter.Emit(evt)

	// Wait for event to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify run was created in storage
	run, err := stor.GetRun(runID)
	if err != nil {
		t.Fatalf("Run not created in storage: %v", err)
	}

	if run.ID != runID {
		t.Errorf("Expected run ID %s, got %s", runID, run.ID)
	}
	if run.Status != storage.RunStatusRunning {
		t.Errorf("Expected status running, got %s", run.Status)
	}
	if run.Prompt != data.Prompt {
		t.Errorf("Expected prompt %q, got %q", data.Prompt, run.Prompt)
	}
}

// TestEventBridgeConcurrentRunStarts verifies thread-safety of concurrent run starts
func TestEventBridgeConcurrentRunStarts(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)
	defer bridge.Close()

	runID := "concurrent-run-123"
	numGoroutines := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Emit the same run start event from multiple goroutines concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			data := events.RunStartData{
				Prompt: "Concurrent test prompt",
				Model:  "claude-3-5-sonnet",
			}

			evt, err := events.NewRunStartEvent(runID, "test-agent", "Test Agent", data)
			if err != nil {
				t.Errorf("Goroutine %d: Failed to create event: %v", i, err)
				return
			}

			emitter.Emit(evt)
		}(i)
	}

	wg.Wait()

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify only one run was created despite multiple concurrent events
	runs, _ := stor.ListRuns(storage.RunListOptions{})
	count := 0
	for _, run := range runs {
		if run.ID == runID {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected exactly 1 run to be created, got %d", count)
	}
}

// TestEventBridgeRunEndHandling verifies run end event handling
func TestEventBridgeRunEndHandling(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)
	defer bridge.Close()

	runID := "test-run-end"

	// First create a run
	startData := events.RunStartData{Prompt: "Test", Model: "claude-3-5-sonnet"}
	startEvt, _ := events.NewRunStartEvent(runID, "test-agent", "Test Agent", startData)
	emitter.Emit(startEvt)
	time.Sleep(50 * time.Millisecond)

	// Now send run end event
	endData := events.RunEndData{
		Success:         true,
		TotalToolCalls:  5,
		TotalIterations: 3,
		Duration:        "1.5s",
	}

	endEvt, err := events.NewRunEndEvent(runID, "test-agent", "Test Agent", endData)
	if err != nil {
		t.Fatalf("Failed to create end event: %v", err)
	}

	emitter.Emit(endEvt)
	time.Sleep(50 * time.Millisecond)

	// Verify run status was updated
	run, err := stor.GetRun(runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.Status != storage.RunStatusSuccess {
		t.Errorf("Expected status success, got %s", run.Status)
	}
	if run.TotalToolCalls != 5 {
		t.Errorf("Expected 5 tool calls, got %d", run.TotalToolCalls)
	}
	if run.TotalIterations != 3 {
		t.Errorf("Expected 3 iterations, got %d", run.TotalIterations)
	}
}

// TestEventBridgeTokenUsageUpdate verifies token usage tracking
func TestEventBridgeTokenUsageUpdate(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)
	defer bridge.Close()

	runID := "test-run-tokens"

	// Create a run first
	startData := events.RunStartData{Prompt: "Test", Model: "claude-3-5-sonnet"}
	startEvt, _ := events.NewRunStartEvent(runID, "test-agent", "Test Agent", startData)
	emitter.Emit(startEvt)
	time.Sleep(50 * time.Millisecond)

	// Send API response event with token usage
	apiData := events.APIResponseData{
		StopReason: "end_turn",
		Model:      "claude-3-5-sonnet",
		Usage: &events.UsageInfo{
			InputTokens:  100,
			OutputTokens: 50,
		},
		ContentCount: 1,
	}

	apiEvt, err := events.NewAPIResponseEvent(runID, "test-agent", "Test Agent", apiData)
	if err != nil {
		t.Fatalf("Failed to create API response event: %v", err)
	}

	emitter.Emit(apiEvt)
	time.Sleep(50 * time.Millisecond)

	// Verify tokens were updated
	run, err := stor.GetRun(runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.TotalTokens.InputTokens != 100 {
		t.Errorf("Expected 100 input tokens, got %d", run.TotalTokens.InputTokens)
	}
	if run.TotalTokens.OutputTokens != 50 {
		t.Errorf("Expected 50 output tokens, got %d", run.TotalTokens.OutputTokens)
	}
	if run.TotalTokens.TotalTokens != 150 {
		t.Errorf("Expected 150 total tokens, got %d", run.TotalTokens.TotalTokens)
	}
}

// TestEventBridgeGetEmitter verifies GetEmitter method
func TestEventBridgeGetEmitter(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)
	defer bridge.Close()

	retrievedEmitter := bridge.GetEmitter()
	if retrievedEmitter != emitter {
		t.Error("GetEmitter returned different emitter")
	}
}

// TestEventBridgeClose verifies cleanup
func TestEventBridgeClose(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)

	// Close should not panic
	bridge.Close()

	// Multiple closes should be safe
	bridge.Close()
}

// TestEventBridgeFailedRunHandling verifies failed run handling
func TestEventBridgeFailedRunHandling(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)
	defer bridge.Close()

	runID := "test-run-failed"

	// Create a run
	startData := events.RunStartData{Prompt: "Test", Model: "claude-3-5-sonnet"}
	startEvt, _ := events.NewRunStartEvent(runID, "test-agent", "Test Agent", startData)
	emitter.Emit(startEvt)
	time.Sleep(50 * time.Millisecond)

	// Send failed run end event
	endData := events.RunEndData{
		Success:         false,
		Error:           "API call failed",
		TotalToolCalls:  2,
		TotalIterations: 1,
		Duration:        "0.5s",
	}

	endEvt, _ := events.NewRunEndEvent(runID, "test-agent", "Test Agent", endData)
	emitter.Emit(endEvt)
	time.Sleep(50 * time.Millisecond)

	// Verify run status is failed
	run, err := stor.GetRun(runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.Status != storage.RunStatusFailed {
		t.Errorf("Expected status failed, got %s", run.Status)
	}
	if run.Error != "API call failed" {
		t.Errorf("Expected error message 'API call failed', got %q", run.Error)
	}
}

// TestEventBridgeEventStorage verifies events are stored
func TestEventBridgeEventStorage(t *testing.T) {
	stor := NewMockStorage()
	emitter := events.NewEmitter()
	bridge := NewEventBridge(stor, emitter)
	defer bridge.Close()

	runID := "test-run-events"

	// Create various events
	startData := events.RunStartData{Prompt: "Test", Model: "claude-3-5-sonnet"}
	startEvt, _ := events.NewRunStartEvent(runID, "test-agent", "Test Agent", startData)
	emitter.Emit(startEvt)

	messageData := events.MessageData{
		Role: "user",
		Content: []events.ContentBlock{
			{Type: "text", Text: "Hello"},
		},
	}
	msgEvt, _ := events.NewMessageEvent(runID, "test-agent", "Test Agent", messageData)
	emitter.Emit(msgEvt)

	time.Sleep(50 * time.Millisecond)

	// Verify events were stored
	storedEvents, err := stor.GetEvents(runID, nil)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(storedEvents) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(storedEvents))
	}

	// Verify event types
	foundStart := false
	foundMessage := false
	for _, evt := range storedEvents {
		if evt.Type == events.EventRunStart {
			foundStart = true
		}
		if evt.Type == events.EventMessage {
			foundMessage = true

			// Verify message data
			var msgData events.MessageData
			if err := json.Unmarshal(evt.Data, &msgData); err == nil {
				if msgData.Role != "user" {
					t.Errorf("Expected role 'user', got %s", msgData.Role)
				}
			}
		}
	}

	if !foundStart {
		t.Error("Run start event not found in storage")
	}
	if !foundMessage {
		t.Error("Message event not found in storage")
	}
}
