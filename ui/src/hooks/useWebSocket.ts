import { useEffect, useRef, useState, useCallback } from "react";
import type { AgentEvent } from "../types/events";

export const WebSocketStatus = {
  CONNECTING: "connecting",
  CONNECTED: "connected",
  DISCONNECTED: "disconnected",
  ERROR: "error",
} as const;

export type WebSocketStatus =
  (typeof WebSocketStatus)[keyof typeof WebSocketStatus];

interface UseWebSocketOptions {
  onEvent?: (event: AgentEvent) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export function useWebSocket(
  url: string | null,
  options: UseWebSocketOptions = {},
) {
  const [status, setStatus] = useState<WebSocketStatus>(
    WebSocketStatus.DISCONNECTED,
  );
  const [lastEvent, setLastEvent] = useState<AgentEvent | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const reconnectTimer = useRef<number | null>(null);

  const {
    onEvent,
    onConnect,
    onDisconnect,
    onError,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options;

  const connect = useCallback(() => {
    if (!url) return;

    try {
      setStatus(WebSocketStatus.CONNECTING);
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setStatus(WebSocketStatus.CONNECTED);
        reconnectAttempts.current = 0;
        onConnect?.();
      };

      ws.onmessage = (event) => {
        try {
          const agentEvent: AgentEvent = JSON.parse(event.data);
          setLastEvent(agentEvent);
          onEvent?.(agentEvent);
        } catch (error) {
          console.error("Failed to parse WebSocket message:", error);
        }
      };

      ws.onerror = (error) => {
        setStatus(WebSocketStatus.ERROR);
        onError?.(error);
      };

      ws.onclose = () => {
        setStatus(WebSocketStatus.DISCONNECTED);
        onDisconnect?.();

        // Auto-reconnect logic
        if (reconnectAttempts.current < maxReconnectAttempts) {
          reconnectAttempts.current += 1;
          reconnectTimer.current = window.setTimeout(() => {
            connect();
          }, reconnectInterval);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error("Failed to create WebSocket:", error);
      setStatus(WebSocketStatus.ERROR);
    }
  }, [
    url,
    onEvent,
    onConnect,
    onDisconnect,
    onError,
    reconnectInterval,
    maxReconnectAttempts,
  ]);

  const disconnect = useCallback(() => {
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setStatus(WebSocketStatus.DISCONNECTED);
  }, []);

  useEffect(() => {
    if (url) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [url, connect, disconnect]);

  return {
    status,
    lastEvent,
    connect,
    disconnect,
  };
}
