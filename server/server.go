package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/agent"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/server/service"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// ActiveRun tracks an active agent run
type ActiveRun struct {
	RunID          string
	CancelFunc     context.CancelFunc
	Context        context.Context
	PendingMessage chan string
}

// Server represents the HTTP/WebSocket server
type Server struct {
	storage          storage.Storage
	hub              *WebSocketHub
	router           *mux.Router
	port             int
	storageCleanup   func()
	agentRegistry    *agent.Registry
	anthropicClient  *anthropic.Client
	logLevel         config.LogLevel
	eventBridge      *EventBridge
	activeRuns       map[string]*ActiveRun
	activeRunsMu     sync.RWMutex
	runService       *service.RunService       // Service layer for business logic
	activeRunManager *service.ActiveRunManagerImpl // Manager for active runs
}

// Config holds server configuration
type Config struct {
	Storage         storage.Storage
	Port            int
	AgentRegistry   *agent.Registry
	AnthropicClient *anthropic.Client
	LogLevel        config.LogLevel
	EventBridge     *EventBridge
}

// New creates a new server
func New(cfg Config) *Server {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	// Create active run manager
	activeRunManager := service.NewActiveRunManager()

	// Create event bridge adapter
	eventBridgeAdapter := service.NewEventBridgeAdapter(cfg.EventBridge)

	// Create run service
	runService := service.NewRunService(service.Config{
		Storage:         cfg.Storage,
		AgentRegistry:   cfg.AgentRegistry,
		AnthropicClient: cfg.AnthropicClient,
		LogLevel:        cfg.LogLevel,
		EventBridge:     eventBridgeAdapter,
		ActiveRuns:      activeRunManager,
	})

	s := &Server{
		storage:          cfg.Storage,
		hub:              NewWebSocketHub(),
		router:           mux.NewRouter(),
		port:             cfg.Port,
		agentRegistry:    cfg.AgentRegistry,
		anthropicClient:  cfg.AnthropicClient,
		logLevel:         cfg.LogLevel,
		eventBridge:      cfg.EventBridge,
		activeRuns:       make(map[string]*ActiveRun),
		runService:       runService,
		activeRunManager: activeRunManager,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()

	// Health check
	api.HandleFunc("/health", s.HandleHealth).Methods("GET")

	// Agent endpoints
	api.HandleFunc("/agents", s.HandleListAgents).Methods("GET")

	// Run endpoints
	api.HandleFunc("/runs", s.HandleListRuns).Methods("GET")
	api.HandleFunc("/runs", s.HandleCreateRun).Methods("POST")
	api.HandleFunc("/runs/{runId}", s.HandleGetRun).Methods("GET")
	api.HandleFunc("/runs/{runId}/events", s.HandleGetEvents).Methods("GET")
	api.HandleFunc("/runs/{runId}/events/count", s.HandleGetEventCount).Methods("GET")

	// Run control endpoints
	api.HandleFunc("/runs/{runId}/stop", s.HandleStopRun).Methods("POST")
	api.HandleFunc("/runs/{runId}/pause", s.HandlePauseRun).Methods("POST")
	api.HandleFunc("/runs/{runId}/resume", s.HandleResumeRun).Methods("POST")
	api.HandleFunc("/runs/{runId}/instruction", s.HandleAddInstruction).Methods("POST")

	// WebSocket endpoints
	api.HandleFunc("/runs/{runId}/stream", s.HandleWebSocket)
	api.HandleFunc("/stream", s.HandleWebSocketAll)

	// CORS middleware
	s.router.Use(corsMiddleware)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins for development
		// In production, you should restrict this
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start starts the server
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.hub.Run()

	// Subscribe to storage events
	s.SubscribeToStorage()

	// Listen on all interfaces (0.0.0.0) for Docker compatibility
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	log.Printf("Server starting on http://%s", addr)
	log.Printf("API available at http://%s/api", addr)
	log.Printf("WebSocket endpoint: ws://%s/api/stream", addr)

	return http.ListenAndServe(addr, s.router)
}

// Close gracefully shuts down the server
func (s *Server) Close() error {
	// Cancel all active runs using the manager
	s.activeRunManager.CancelAll()

	if s.storageCleanup != nil {
		s.storageCleanup()
	}
	return s.storage.Close()
}

// addActiveRun adds a new active run
func (s *Server) addActiveRun(runID string, ctx context.Context, cancel context.CancelFunc) {
	s.activeRunsMu.Lock()
	defer s.activeRunsMu.Unlock()
	s.activeRuns[runID] = &ActiveRun{
		RunID:          runID,
		Context:        ctx,
		CancelFunc:     cancel,
		PendingMessage: make(chan string, 10), // Buffered channel for pending messages
	}
}

// getActiveRun retrieves an active run
func (s *Server) getActiveRun(runID string) (*ActiveRun, bool) {
	s.activeRunsMu.RLock()
	defer s.activeRunsMu.RUnlock()
	activeRun, exists := s.activeRuns[runID]
	return activeRun, exists
}

// removeActiveRun removes an active run
func (s *Server) removeActiveRun(runID string) {
	s.activeRunsMu.Lock()
	defer s.activeRunsMu.Unlock()
	delete(s.activeRuns, runID)
}
