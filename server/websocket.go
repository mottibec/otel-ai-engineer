package server

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, you should restrict this
		return true
	},
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	conn   *websocket.Conn
	send   chan *events.AgentEvent
	runID  string
	mu     sync.Mutex
	closed bool
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan *events.AgentEvent
	mu         sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan *events.AgentEvent, 256),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.mu.Lock()
				if !client.closed {
					close(client.send)
					client.closed = true
				}
				client.mu.Unlock()
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Only send to clients subscribed to this run or to all runs
				if client.runID == "" || client.runID == event.RunID {
					select {
					case client.send <- event:
					default:
						// Client's send buffer is full, close the connection
						go func(c *WebSocketClient) {
							h.unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent broadcasts an event to all connected clients
func (h *WebSocketHub) BroadcastEvent(event *events.AgentEvent) {
	select {
	case h.broadcast <- event:
	default:
		log.Printf("Warning: broadcast channel full, dropping event")
	}
}

// HandleWebSocket handles WebSocket connections for a specific run
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runId"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &WebSocketClient{
		conn:  conn,
		send:  make(chan *events.AgentEvent, 256),
		runID: runID,
	}

	s.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(s.hub)

	// Send existing events for this run
	go s.sendExistingEvents(client, runID)
}

// HandleWebSocketAll handles WebSocket connections for all runs
func (s *Server) HandleWebSocketAll(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &WebSocketClient{
		conn:  conn,
		send:  make(chan *events.AgentEvent, 256),
		runID: "", // Empty means all runs
	}

	s.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(s.hub)
}

// sendExistingEvents sends all existing events for a run to a new client
func (s *Server) sendExistingEvents(client *WebSocketClient, runID string) {
	// Check if storage is available
	if s.storage == nil {
		log.Printf("Storage is nil, skipping existing events for run %s", runID)
		return
	}

	events, err := s.storage.GetEvents(runID, nil)
	if err != nil {
		log.Printf("Failed to get events for run %s: %v", runID, err)
		return
	}

	log.Printf("Sending %d existing events for run %s to new client", len(events), runID)

	for _, event := range events {
		if event == nil {
			log.Printf("Skipping nil event")
			continue
		}

		// Check if client is still connected before sending
		client.mu.Lock()
		isClosed := client.closed
		client.mu.Unlock()

		if isClosed {
			log.Printf("Client disconnected, stopping event replay")
			return
		}

		// Send with timeout and panic recovery
		shouldStop := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic while sending event to client: %v", r)
					shouldStop = true
				}
			}()

			select {
			case client.send <- event:
			case <-time.After(5 * time.Second):
				log.Printf("Timeout sending existing event to client")
				shouldStop = true
			}
		}()

		if shouldStop {
			return
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send the event as JSON
			if err := c.conn.WriteJSON(event); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *WebSocketClient) readPump(hub *WebSocketHub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		// We don't expect clients to send messages, but we need to read to detect disconnects
	}
}

// SubscribeToStorage subscribes the hub to storage events
func (s *Server) SubscribeToStorage() {
	eventChan, cleanup := s.storage.SubscribeAll()

	// Store cleanup function
	s.storageCleanup = cleanup

	// Start goroutine to forward storage events to WebSocket hub
	go func() {
		for event := range eventChan {
			s.hub.BroadcastEvent(event)
		}
	}()
}
