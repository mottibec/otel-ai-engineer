package server

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// EventBridge bridges agent events to storage
type EventBridge struct {
	storage storage.Storage
	emitter events.EventEmitter
	cleanup func()
	runMap  map[string]bool // Track which runs have been created
	runMapMu sync.RWMutex   // Protect runMap from concurrent access
}

// NewEventBridge creates a new event bridge
func NewEventBridge(stor storage.Storage, emitter events.EventEmitter) *EventBridge {
	bridge := &EventBridge{
		storage: stor,
		emitter: emitter,
		runMap:  make(map[string]bool),
	}

	// Subscribe to all events
	eventChan, cleanup := emitter.SubscribeAll()
	bridge.cleanup = cleanup

	// Start processing events
	go bridge.processEvents(eventChan)

	return bridge
}

// processEvents processes events from the emitter and stores them
func (b *EventBridge) processEvents(eventChan <-chan *events.AgentEvent) {
	for event := range eventChan {
		// Handle run start - create run in storage
		if event.Type == events.EventRunStart {
			// Check if run already exists (thread-safe)
			b.runMapMu.RLock()
			exists := b.runMap[event.RunID]
			b.runMapMu.RUnlock()

			if !exists {
				var data events.RunStartData
				if err := json.Unmarshal(event.Data, &data); err == nil {
					run := &storage.Run{
						ID:              event.RunID,
						AgentID:         event.AgentID,
						AgentName:       event.AgentName,
						Status:          storage.RunStatusRunning,
						Prompt:          data.Prompt,
						Model:           data.Model,
						StartTime:       event.Timestamp,
						TotalIterations: 0,
						TotalToolCalls:  0,
						TotalTokens:     storage.TokenUsage{},
					}

					if err := b.storage.CreateRun(run); err != nil {
						log.Printf("Failed to create run: %v", err)
					} else {
						// Mark run as created (thread-safe)
						b.runMapMu.Lock()
						b.runMap[event.RunID] = true
						b.runMapMu.Unlock()
					}
				}
			}
		}

		// Handle run end - update run status
		if event.Type == events.EventRunEnd {
			var data events.RunEndData
			if err := json.Unmarshal(event.Data, &data); err == nil {
				status := storage.RunStatusSuccess
				if !data.Success {
					status = storage.RunStatusFailed
				}

				endTime := event.Timestamp
				update := &storage.RunUpdate{
					Status:          &status,
					EndTime:         &endTime,
					Duration:        &data.Duration,
					TotalIterations: &data.TotalIterations,
					TotalToolCalls:  &data.TotalToolCalls,
				}

				if data.Error != "" {
					update.Error = &data.Error
				}

				if err := b.storage.UpdateRun(event.RunID, update); err != nil {
					log.Printf("Failed to update run: %v", err)
				}
			}
		}

		// Handle API response - update token usage
		if event.Type == events.EventAPIResponse {
			var data events.APIResponseData
			if err := json.Unmarshal(event.Data, &data); err == nil {
				if data.Usage != nil {
					// Get current run to calculate cumulative tokens
					run, err := b.storage.GetRun(event.RunID)
					if err == nil {
						newTokens := storage.TokenUsage{
							InputTokens:  run.TotalTokens.InputTokens + data.Usage.InputTokens,
							OutputTokens: run.TotalTokens.OutputTokens + data.Usage.OutputTokens,
							TotalTokens:  run.TotalTokens.TotalTokens + data.Usage.InputTokens + data.Usage.OutputTokens,
						}

						update := &storage.RunUpdate{
							TotalTokens: &newTokens,
						}

						if err := b.storage.UpdateRun(event.RunID, update); err != nil {
							log.Printf("Failed to update run tokens: %v", err)
						}
					}
				}
			}
		}

		// Store the event
		if err := b.storage.AddEvent(event.RunID, event); err != nil {
			log.Printf("Failed to store event: %v", err)
		}
	}
}

// GetEmitter returns the event emitter
func (b *EventBridge) GetEmitter() events.EventEmitter {
	return b.emitter
}

// Close closes the bridge
func (b *EventBridge) Close() {
	if b.cleanup != nil {
		b.cleanup()
	}
}
