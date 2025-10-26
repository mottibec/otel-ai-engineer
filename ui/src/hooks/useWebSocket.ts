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
  const currentUrlRef = useRef<string | null>(null);

  const {
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options;

  // Use refs to store callbacks so they don't trigger re-renders
  const callbacksRef = useRef(options);
  useEffect(() => {
    callbacksRef.current = options;
  }, [options]);

  useEffect(() => {
    if (!url) {
      console.log('[WebSocket] URL is null, disconnecting');
      // Disconnect if URL is cleared
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
      currentUrlRef.current = null;
      return;
    }

    console.log('[WebSocket] Attempting to connect to:', url);

    // Don't reconnect if already connected to the same URL
    if (currentUrlRef.current === url && wsRef.current) {
      console.log('[WebSocket] Already connected to this URL, skipping');
      return;
    }

    // Clean up previous connection to different URL
    if (wsRef.current && currentUrlRef.current !== url) {
      wsRef.current.close();
      wsRef.current = null;
    }

    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }

    const thisUrl = url; // Capture URL for this effect run
    currentUrlRef.current = thisUrl;

    let ws: WebSocket | null = null;

    try {
      console.log('[WebSocket] Creating new WebSocket connection to:', thisUrl);
      setStatus(WebSocketStatus.CONNECTING);
      ws = new WebSocket(thisUrl);

      ws.onopen = () => {
        console.log('[WebSocket] Connection opened successfully');
        // Only update status if this is still the current connection
        if (wsRef.current === ws) {
          setStatus(WebSocketStatus.CONNECTED);
          reconnectAttempts.current = 0;
          callbacksRef.current.onConnect?.();
        }
      };

      ws.onmessage = (event) => {
        try {
          const agentEvent: AgentEvent = JSON.parse(event.data);
          // Only process event if this is still the current connection
          if (wsRef.current === ws) {
            setLastEvent(agentEvent);
            callbacksRef.current.onEvent?.(agentEvent);
          }
        } catch (error) {
          console.error("Failed to parse WebSocket message:", error);
        }
      };

      ws.onerror = (error) => {
        console.error('[WebSocket] Connection error:', error);
        if (wsRef.current === ws) {
          setStatus(WebSocketStatus.ERROR);
          callbacksRef.current.onError?.(error);
        }
      };

      ws.onclose = () => {
        // Only update status if this is still the current connection
        if (wsRef.current === ws) {
          setStatus(WebSocketStatus.DISCONNECTED);
          callbacksRef.current.onDisconnect?.();

          // Auto-reconnect logic
          const shouldReconnect = 
            reconnectAttempts.current < maxReconnectAttempts && 
            currentUrlRef.current === thisUrl;

          if (shouldReconnect) {
            reconnectAttempts.current += 1;
            reconnectTimer.current = window.setTimeout(() => {
              // Only reconnect if the URL hasn't changed
              if (currentUrlRef.current === thisUrl) {
                wsRef.current = null;
                // Re-trigger effect by calling setStatus
                setStatus(WebSocketStatus.DISCONNECTED);
              }
            }, reconnectInterval);
          } else {
            wsRef.current = null;
          }
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error("Failed to create WebSocket:", error);
      setStatus(WebSocketStatus.ERROR);
    }

    return () => {
      // Only cleanup if:
      // 1. This is the current socket (not replaced by a different URL change)
      // 2. And the current URL matches this one
      // 3. And the socket is NOT connecting (prevents premature cleanup in StrictMode)
      const currentWs = wsRef.current;
      const currentUrl = currentUrlRef.current;
      
      // Only close if this is still the active connection and the socket has opened
      // This prevents closing during React StrictMode double-invoke when the socket is still connecting
      if (currentWs && ws && currentWs === ws) {
        if (currentUrl === thisUrl && currentWs.readyState === WebSocket.OPEN) {
          // Same URL and socket is open - normal cleanup
          currentWs.close();
          wsRef.current = null;
          currentUrlRef.current = null;
        } else if (currentUrl !== thisUrl) {
          // URL changed, always clean up regardless of state
          currentWs.close();
          wsRef.current = null;
        }
        // If same URL but still connecting, don't close (StrictMode protection)
      }
      
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
    };
  }, [url, maxReconnectAttempts, reconnectInterval]);

  const connect = useCallback(() => {
    // Connection is handled by the effect
  }, []);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
    currentUrlRef.current = null;
    setStatus(WebSocketStatus.DISCONNECTED);
  }, []);

  return {
    status,
    lastEvent,
    connect,
    disconnect,
  };
}
