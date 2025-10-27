import { useState, useCallback, useRef, useEffect, useMemo } from "react";
import type { AgentEvent } from "../types/events";
import type { RunStatus } from "../types/models";
import { apiClient } from "../services/api";
import { useWebSocket } from "./useWebSocket";

export function useRunStream(runId: string | null, runStatus?: RunStatus) {
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const seenEventIds = useRef<Set<string>>(new Set());

  // Determine if this run is active (should use WebSocket) or completed (should use REST API)
  const isActiveRun = runStatus === "running" || runStatus === "paused";
  const isCompletedRun = runStatus === "success" || runStatus === "failed" || runStatus === "cancelled";

  // Reset seen IDs when runId changes
  useEffect(() => {
    seenEventIds.current.clear();
    setEvents([]);
  }, [runId]);

  // Fetch existing events
  const fetchEvents = useCallback(async () => {
    if (!runId) return;

    console.log('[useRunStream] Fetching events for runId:', runId);
    try {
      setLoading(true);
      const data = await apiClient.getEvents(runId);
      console.log('[useRunStream] Fetched events:', data?.length || 0, 'events');
      // Deduplicate events by ID when setting them
      setEvents((prev) => {
        if (!prev || prev.length === 0) {
          // Track all event IDs
          (data || []).forEach((e) => seenEventIds.current.add(e.id));
          return data || [];
        }
        // Merge fetched events with existing ones, deduplicating by ID
        const eventMap = new Map<string, AgentEvent>();
        // Add existing events
        prev.forEach((e) => eventMap.set(e.id, e));
        // Add fetched events (will overwrite duplicates)
        (data || []).forEach((e) => {
          if (!eventMap.has(e.id) && !seenEventIds.current.has(e.id)) {
            seenEventIds.current.add(e.id);
          }
          eventMap.set(e.id, e);
        });
        // Sort by timestamp
        return Array.from(eventMap.values()).sort((a, b) => 
          new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
        );
      });
    } catch (error) {
      console.error("Failed to fetch events:", error);
    } finally {
      setLoading(false);
    }
  }, [runId]);

  // Fetch events from REST API for completed runs
  useEffect(() => {
    if (!runId) {
      setLoading(false);
      return;
    }

    // For completed runs, fetch all events via REST API (no WebSocket)
    if (isCompletedRun) {
      console.log('[useRunStream] Fetching events via REST API for completed run:', runId);
      setLoading(true);
      apiClient.getEvents(runId).then((data) => {
        (data || []).forEach((e) => seenEventIds.current.add(e.id));
        setEvents(data || []);
      }).catch((error) => {
        console.error("Failed to fetch events:", error);
      }).finally(() => {
        setLoading(false);
      });
    } else {
      // For active runs, rely on WebSocket (sendExistingEvents will send initial events)
      console.log('[useRunStream] Using WebSocket for active run:', runId);
      setLoading(false);
    }
  }, [runId, isCompletedRun]);

  // Handle new events from WebSocket
  const handleEvent = useCallback((event: AgentEvent) => {
    console.log('[useRunStream] handleEvent called:', event.type, event.id);

    // Check if event already exists BEFORE calling setEvents (deduplicate by ID)
    if (seenEventIds.current.has(event.id)) {
      console.log('[useRunStream] Event already seen, skipping:', event.id);
      return;
    }

    // Mark as seen BEFORE calling setEvents to prevent race conditions
    console.log('[useRunStream] Adding new event to state:', event.type, event.id);
    seenEventIds.current.add(event.id);

    // Update state
    setEvents((prev) => {
      const newEvents = [...prev, event];
      console.log('[useRunStream] New events array length:', newEvents.length);
      return newEvents;
    });
  }, []);

  // Memoize WebSocket options to prevent recreating on every render
  const wsOptions = useMemo(() => ({
    onEvent: handleEvent,
  }), [handleEvent]);

  // Connect to WebSocket ONLY for active runs (running/paused)
  // For completed runs, use REST API exclusively
  const wsUrl = (runId && isActiveRun) ? apiClient.getWebSocketUrl(runId) : null;
  const { status } = useWebSocket(wsUrl, wsOptions);

  return { events, loading, status, refetch: fetchEvents };
}

