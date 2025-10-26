import { useState, useCallback, useRef, useEffect, useMemo } from "react";
import type { AgentEvent } from "../types/events";
import { apiClient } from "../services/api";
import { useWebSocket } from "./useWebSocket";

export function useRunStream(runId: string | null) {
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const seenEventIds = useRef<Set<string>>(new Set());
  const recentSentMessageContent = useRef<string | null>(null);

  // Reset seen IDs when runId changes
  useEffect(() => {
    seenEventIds.current.clear();
    recentSentMessageContent.current = null;
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

  // Fetch events initially when runId changes
  useEffect(() => {
    if (runId) {
      setLoading(true);
      apiClient.getEvents(runId).then((data) => {
        setEvents((prev) => {
          if (!prev || prev.length === 0) {
            (data || []).forEach((e) => seenEventIds.current.add(e.id));
            return data || [];
          }
          const eventMap = new Map<string, AgentEvent>();
          prev.forEach((e) => eventMap.set(e.id, e));
          (data || []).forEach((e) => {
            if (!eventMap.has(e.id) && !seenEventIds.current.has(e.id)) {
              seenEventIds.current.add(e.id);
            }
            eventMap.set(e.id, e);
          });
          return Array.from(eventMap.values()).sort((a, b) => 
            new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
          );
        });
      }).catch((error) => {
        console.error("Failed to fetch events:", error);
      }).finally(() => {
        setLoading(false);
      });
    }
  }, [runId]); // Only depend on runId, not fetchEvents to avoid infinite loop

  // Handle new events from WebSocket
  const handleEvent = useCallback((event: AgentEvent) => {
    setEvents((prev) => {
      // Check if event already exists (deduplicate by ID)
      if (seenEventIds.current.has(event.id)) {
        return prev;
      }
      
      // Check if this is a user message that we just sent
      if (event.type === "message") {
        const data = event.data as any;
        if (data.role === "user" && recentSentMessageContent.current) {
          // Check if the content matches a recently sent message
          const eventContent = data.content?.find((c: any) => c.type === "text")?.text || "";
          if (eventContent === recentSentMessageContent.current) {
            // This is our own message echoed back, clear the tracking and skip
            recentSentMessageContent.current = null;
            return prev;
          }
        }
      }
      
      seenEventIds.current.add(event.id);
      return [...prev, event];
    });
  }, []);

  // Function to mark a message as recently sent (called before sending to server)
  const markMessageAsSent = useCallback((messageContent: string) => {
    recentSentMessageContent.current = messageContent;
    // Clear after 5 seconds to prevent stale matches
    setTimeout(() => {
      if (recentSentMessageContent.current === messageContent) {
        recentSentMessageContent.current = null;
      }
    }, 5000);
  }, []);

  // Memoize WebSocket options to prevent recreating on every render
  const wsOptions = useMemo(() => ({
    onEvent: handleEvent,
  }), [handleEvent]);

  // Connect to WebSocket
  const wsUrl = runId ? apiClient.getWebSocketUrl(runId) : null;
  const { status } = useWebSocket(wsUrl, wsOptions);

  return { events, loading, status, refetch: fetchEvents, markMessageAsSent };
}

