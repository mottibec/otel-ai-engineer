package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// TraceService handles trace computation from events
type TraceService struct {
	storage storage.Storage
}

// NewTraceService creates a new trace service
func NewTraceService(storage storage.Storage) *TraceService {
	return &TraceService{
		storage: storage,
	}
}

// ComputeTrace computes a hierarchical trace from run events
func (s *TraceService) ComputeTrace(runID string) (*storage.Trace, error) {
	// Get all events for the run
	allEvents, err := s.storage.GetEvents(runID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(allEvents) == 0 {
		return nil, fmt.Errorf("no events found for run %s", runID)
	}

	// Find run start event
	var runStart *events.AgentEvent
	var runEnd *events.AgentEvent
	for _, event := range allEvents {
		if event.Type == events.EventRunStart {
			runStart = event
		}
		if event.Type == events.EventRunEnd {
			runEnd = event
		}
	}

	if runStart == nil {
		return nil, fmt.Errorf("no run_start event found")
	}

	trace := &storage.Trace{
		TraceID:   runID,
		StartTime: runStart.Timestamp,
	}

	if runEnd != nil {
		trace.EndTime = &runEnd.Timestamp
		duration := runEnd.Timestamp.Sub(runStart.Timestamp)
		trace.Duration = duration.String()
		trace.DurationMs = duration.Milliseconds()
	}

	// Create root span from run boundaries
	rootSpan := &storage.Span{
		ID:        fmt.Sprintf("root-%s", runID),
		Type:      storage.SpanTypeTrace,
		Name:      "Agent Run",
		StartTime: runStart.Timestamp,
		Tags:      make(map[string]interface{}),
	}

	if runEnd != nil {
		rootSpan.EndTime = &runEnd.Timestamp
		duration := runEnd.Timestamp.Sub(runStart.Timestamp)
		rootSpan.Duration = duration.String()
		rootSpan.DurationMs = duration.Milliseconds()

		// Parse run end data for error info
		var runEndData events.RunEndData
		if err := json.Unmarshal(runEnd.Data, &runEndData); err == nil {
			rootSpan.Error = !runEndData.Success
			if runEndData.Error != "" {
				rootSpan.ErrorMsg = runEndData.Error
			}
		}
	}

	// Build span tree from events
	spans := s.buildSpanTree(allEvents)
	rootSpan.Children = spans

	trace.RootSpan = rootSpan

	return trace, nil
}

// buildSpanTree builds a hierarchical tree of spans from flat events
func (s *TraceService) buildSpanTree(eventList []*events.AgentEvent) []*storage.Span {
	spans := []*storage.Span{}

	// Track pending spans by their IDs
	pendingSpans := make(map[string]*storage.Span)
	spanMap := make(map[string]*storage.Span) // Map for fast lookup

	// Track iteration context
	var currentIteration *storage.Span
	iterationNum := 0

	for _, event := range eventList {
		switch event.Type {
		case events.EventIteration:
			// Start new iteration span
			iterationNum++
			newIteration := &storage.Span{
				ID:        fmt.Sprintf("iteration-%d", iterationNum),
				Type:      storage.SpanTypeIteration,
				Name:      fmt.Sprintf("Iteration %d", iterationNum),
				StartTime: event.Timestamp,
				Tags:      make(map[string]interface{}),
			}

			var iterData events.IterationData
			if err := json.Unmarshal(event.Data, &iterData); err == nil {
				newIteration.Tags["iteration_number"] = iterData.Iteration
				newIteration.Tags["total_messages"] = iterData.TotalMessages
			}

			currentIteration = newIteration
			spans = append(spans, newIteration)

		case events.EventToolCall:
			// Start tool span
			var toolCallData events.ToolCallData
			if err := json.Unmarshal(event.Data, &toolCallData); err != nil {
				continue
			}

			toolSpan := &storage.Span{
				ID:        fmt.Sprintf("tool-%s", toolCallData.ToolUseID),
				Type:      storage.SpanTypeTool,
				Name:      toolCallData.ToolName,
				StartTime: event.Timestamp,
				Tags:      make(map[string]interface{}),
			}

			toolSpan.Tags["tool_use_id"] = toolCallData.ToolUseID
			toolSpan.Tags["tool_input"] = toolCallData.Input

			if currentIteration != nil {
				toolSpan.ParentSpanID = &currentIteration.ID
				currentIteration.Children = append(currentIteration.Children, toolSpan)
			} else {
				spans = append(spans, toolSpan)
			}

			pendingSpans[toolCallData.ToolUseID] = toolSpan
			spanMap[toolSpan.ID] = toolSpan

		case events.EventToolResult:
			// Complete tool span
			var toolResultData events.ToolResultData
			if err := json.Unmarshal(event.Data, &toolResultData); err != nil {
				continue
			}

			if toolSpan, exists := pendingSpans[toolResultData.ToolUseID]; exists {
				toolSpan.EndTime = &event.Timestamp
				duration := event.Timestamp.Sub(toolSpan.StartTime)
				toolSpan.Duration = duration.String()
				toolSpan.DurationMs = duration.Milliseconds()

				toolSpan.Error = toolResultData.IsError
				if toolResultData.Error != "" {
					toolSpan.ErrorMsg = toolResultData.Error
				}

				if toolResultData.Result != nil {
					toolSpan.Tags["tool_result"] = toolResultData.Result
				}

				delete(pendingSpans, toolResultData.ToolUseID)
			}

		case events.EventAPIRequest:
			// Start API call span
			var apiReqData events.APIRequestData
			if err := json.Unmarshal(event.Data, &apiReqData); err != nil {
				continue
			}

			apiSpan := &storage.Span{
				ID:        fmt.Sprintf("api-%d", time.Now().UnixNano()),
				Type:      storage.SpanTypeAPICall,
				Name:      fmt.Sprintf("API Call (%s)", apiReqData.Model),
				StartTime: event.Timestamp,
				Tags:      make(map[string]interface{}),
			}

			apiSpan.Tags["model"] = apiReqData.Model
			apiSpan.Tags["max_tokens"] = apiReqData.MaxTokens
			apiSpan.Tags["tool_count"] = apiReqData.ToolCount

			if currentIteration != nil {
				apiSpan.ParentSpanID = &currentIteration.ID
				currentIteration.Children = append(currentIteration.Children, apiSpan)
			} else {
				spans = append(spans, apiSpan)
			}

			spanMap[apiSpan.ID] = apiSpan
			pendingSpans[apiSpan.ID] = apiSpan

		case events.EventAPIResponse:
			// Complete API call span (find the most recent pending API span)
			var apiRespData events.APIResponseData
			if err := json.Unmarshal(event.Data, &apiRespData); err != nil {
				continue
			}

			// Find pending API span (assume last one is the matching request)
			for id, apiSpan := range pendingSpans {
				if apiSpan.Type == storage.SpanTypeAPICall {
					apiSpan.EndTime = &event.Timestamp
					duration := event.Timestamp.Sub(apiSpan.StartTime)
					apiSpan.Duration = duration.String()
					apiSpan.DurationMs = duration.Milliseconds()

					apiSpan.Tags["stop_reason"] = apiRespData.StopReason
					if apiRespData.Usage != nil {
						apiSpan.Tags["input_tokens"] = apiRespData.Usage.InputTokens
						apiSpan.Tags["output_tokens"] = apiRespData.Usage.OutputTokens
					}
					apiSpan.Tags["content_count"] = apiRespData.ContentCount

					delete(pendingSpans, id)
					break
				}
			}

		case events.EventAgentHandoff:
			// Start handoff span
			var handoffData events.HandoffData
			if err := json.Unmarshal(event.Data, &handoffData); err != nil {
				continue
			}

			handoffSpan := &storage.Span{
				ID:        fmt.Sprintf("handoff-%s", handoffData.SubRunID),
				Type:      storage.SpanTypeAgentHandoff,
				Name:      fmt.Sprintf("Handoff to %s", handoffData.ToAgentID),
				StartTime: event.Timestamp,
				Tags:      make(map[string]interface{}),
			}

			handoffSpan.Tags["task_description"] = handoffData.TaskDescription
			handoffSpan.Tags["from_agent"] = handoffData.FromAgentID
			handoffSpan.Tags["to_agent"] = handoffData.ToAgentID
			handoffSpan.Tags["sub_run_id"] = handoffData.SubRunID

			if currentIteration != nil {
				handoffSpan.ParentSpanID = &currentIteration.ID
				currentIteration.Children = append(currentIteration.Children, handoffSpan)
			} else {
				spans = append(spans, handoffSpan)
			}

			pendingSpans[handoffData.SubRunID] = handoffSpan
			spanMap[handoffSpan.ID] = handoffSpan

		case events.EventAgentHandoffComplete:
			// Complete handoff span
			var handoffCompleteData events.HandoffCompleteData
			if err := json.Unmarshal(event.Data, &handoffCompleteData); err != nil {
				continue
			}

			if handoffSpan, exists := pendingSpans[handoffCompleteData.SubRunID]; exists {
				handoffSpan.EndTime = &event.Timestamp
				duration := event.Timestamp.Sub(handoffSpan.StartTime)
				handoffSpan.Duration = duration.String()
				handoffSpan.DurationMs = duration.Milliseconds()

				handoffSpan.Error = !handoffCompleteData.Success
				if handoffCompleteData.Error != "" {
					handoffSpan.ErrorMsg = handoffCompleteData.Error
				}

				handoffSpan.Tags["summary"] = handoffCompleteData.Summary

				delete(pendingSpans, handoffCompleteData.SubRunID)
			}
		}
	}

	// Close any still-open spans at the end
	// Leave them without end time to indicate they're incomplete
	for range pendingSpans {
		// Spans are left without end time to indicate they're incomplete
	}

	return spans
}
