package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gorilla/mux"
	"github.com/mottibechhofer/otel-ai-engineer/agent"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/otelclient"
	"github.com/mottibechhofer/otel-ai-engineer/server/service"
	backendService "github.com/mottibechhofer/otel-ai-engineer/server/service/backend"
	collectorService "github.com/mottibechhofer/otel-ai-engineer/server/service/collector"
	humanActionService "github.com/mottibechhofer/otel-ai-engineer/server/service/humanaction"
	sandboxService "github.com/mottibechhofer/otel-ai-engineer/server/service/sandbox"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
	agentService "github.com/mottibechhofer/otel-ai-engineer/server/service/agent"
	toolService "github.com/mottibechhofer/otel-ai-engineer/server/service/tools"
	sandboxTools "github.com/mottibechhofer/otel-ai-engineer/tools/sandbox"
	dc "github.com/mottibechhofer/otel-ai-engineer/tools/dockerclient"
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
	storage            storage.Storage
	hub                *WebSocketHub
	router             *mux.Router
	port               int
	storageCleanup     func()
	agentRegistry      *agent.Registry
	anthropicClient    *anthropic.Client
	logLevel           config.LogLevel
	eventBridge        *EventBridge
	activeRuns         map[string]*ActiveRun
	activeRunsMu       sync.RWMutex
	runService         *service.RunService                    // Service layer for business logic
	activeRunManager   *service.ActiveRunManagerImpl          // Manager for active runs
	traceService       *service.TraceService                  // Service for trace computation
	planService        *service.PlanService                   // Service for plan management
	agentWorkService   *service.AgentWorkService              // Service for agent work tracking
	backendService     *backendService.BackendService         // Service for backend management
	collectorService   *collectorService.CollectorService     // Service for collector management
	sandboxService       *sandboxService.SandboxService         // Service for sandbox management
	humanActionService   *humanActionService.HumanActionService   // Service for human action management
	toolDiscoveryService *toolService.ToolDiscoveryService       // Service for tool discovery
	agentService         *agentService.AgentService              // Service for agent management
	otelClient           *otelclient.OtelClient                  // OTEL client for collector management
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

	// Create trace service
	traceService := service.NewTraceService(cfg.Storage)

	// Create plan service
	planService := service.NewPlanService(cfg.Storage)

	// Create agent work service
	agentWorkService := service.NewAgentWorkService(cfg.Storage)

	// Create backend service
	backendService := backendService.NewBackendService(cfg.Storage, agentWorkService)

	// Create OTEL client
	lawrenceURL := os.Getenv("LAWRENCE_API_URL")
	if lawrenceURL == "" {
		lawrenceURL = "http://lawrence:8080"
	}
	otelClient := otelclient.NewOtelClient(lawrenceURL, &http.Client{
		Timeout: 30 * time.Second,
	})

	// Create collector service
	collectorService := collectorService.NewCollectorService(cfg.Storage, agentWorkService, otelClient)

	// Create sandbox service
	sandboxService := sandboxService.NewSandboxService()

	// Create human action service
	humanActionService := humanActionService.NewHumanActionService(cfg.Storage, runService)

	// Create Docker client for tool discovery (Grafana tools need it)
	dockerClient, err := dc.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to create Docker client for tool discovery: %v", err)
		dockerClient = nil
	}

	// Create tool discovery service
	toolDiscoveryService := toolService.NewToolDiscoveryService(dockerClient, otelClient)

	// Create agent service
	agentService := agentService.NewAgentService(cfg.Storage, cfg.AgentRegistry, toolDiscoveryService)

	s := &Server{
		storage:            cfg.Storage,
		hub:                NewWebSocketHub(),
		router:             mux.NewRouter(),
		port:               cfg.Port,
		agentRegistry:      cfg.AgentRegistry,
		anthropicClient:    cfg.AnthropicClient,
		logLevel:           cfg.LogLevel,
		eventBridge:        cfg.EventBridge,
		activeRuns:         make(map[string]*ActiveRun),
		runService:         runService,
		activeRunManager:   activeRunManager,
		traceService:       traceService,
		planService:        planService,
		agentWorkService:   agentWorkService,
		backendService:     backendService,
		collectorService:   collectorService,
		sandboxService:       sandboxService,
		humanActionService:   humanActionService,
		toolDiscoveryService: toolDiscoveryService,
		agentService:         agentService,
		otelClient:           otelClient,
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
	api.HandleFunc("/agents/{agentId}", s.HandleGetAgent).Methods("GET")
	api.HandleFunc("/agents/{agentId}/tools", s.HandleGetAgentTools).Methods("GET")
	api.HandleFunc("/agents/custom", s.HandleCreateCustomAgent).Methods("POST")
	api.HandleFunc("/agents/custom/{agentId}", s.HandleUpdateCustomAgent).Methods("PUT")
	api.HandleFunc("/agents/custom/{agentId}", s.HandleDeleteCustomAgent).Methods("DELETE")
	api.HandleFunc("/agents/meta", s.HandleCreateMetaAgent).Methods("POST")

	// Tool endpoints
	api.HandleFunc("/tools", s.HandleListTools).Methods("GET")

	// Run endpoints
	api.HandleFunc("/runs", s.HandleListRuns).Methods("GET")
	api.HandleFunc("/runs", s.HandleCreateRun).Methods("POST")
	api.HandleFunc("/runs/{runId}", s.HandleGetRun).Methods("GET")
	api.HandleFunc("/runs/{runId}/events", s.HandleGetEvents).Methods("GET")
	api.HandleFunc("/runs/{runId}/events/count", s.HandleGetEventCount).Methods("GET")
	api.HandleFunc("/runs/{runId}/trace", s.HandleGetTrace).Methods("GET")

	// Run control endpoints
	api.HandleFunc("/runs/{runId}/stop", s.HandleStopRun).Methods("POST")
	api.HandleFunc("/runs/{runId}/pause", s.HandlePauseRun).Methods("POST")
	api.HandleFunc("/runs/{runId}/resume", s.HandleResumeRun).Methods("POST")
	api.HandleFunc("/runs/{runId}/instruction", s.HandleAddInstruction).Methods("POST")

	// WebSocket endpoints
	api.HandleFunc("/runs/{runId}/stream", s.HandleWebSocket)
	api.HandleFunc("/stream", s.HandleWebSocketAll)

	// Plan endpoints
	api.HandleFunc("/plans", s.HandleListPlans).Methods("GET")
	api.HandleFunc("/plans", s.HandleCreatePlan).Methods("POST")
	api.HandleFunc("/plans/{planId}", s.HandleGetPlan).Methods("GET")
	api.HandleFunc("/plans/{planId}", s.HandleUpdatePlan).Methods("PUT")
	api.HandleFunc("/plans/{planId}", s.HandleDeletePlan).Methods("DELETE")
	api.HandleFunc("/plans/{planId}/topology", s.HandleGetTopology).Methods("GET")
	api.HandleFunc("/plans/{planId}/execute", s.HandleExecutePlan).Methods("POST")

	// Service component endpoints
	api.HandleFunc("/plans/{planId}/services", s.HandleCreateService).Methods("POST")
	api.HandleFunc("/plans/{planId}/services/{serviceId}", s.HandleUpdateService).Methods("PUT")
	api.HandleFunc("/plans/{planId}/services/{serviceId}", s.HandleDeleteService).Methods("DELETE")

	// Infrastructure component endpoints
	api.HandleFunc("/plans/{planId}/infrastructure", s.HandleCreateInfrastructure).Methods("POST")
	api.HandleFunc("/plans/{planId}/infrastructure/{infraId}", s.HandleUpdateInfrastructure).Methods("PUT")
	api.HandleFunc("/plans/{planId}/infrastructure/{infraId}", s.HandleDeleteInfrastructure).Methods("DELETE")

	// Pipeline component endpoints
	api.HandleFunc("/plans/{planId}/pipelines", s.HandleCreatePipeline).Methods("POST")
	api.HandleFunc("/plans/{planId}/pipelines/from-collector", s.HandleCreatePipelineFromCollector).Methods("POST")
	api.HandleFunc("/plans/{planId}/pipelines/{pipelineId}", s.HandleUpdatePipeline).Methods("PUT")
	api.HandleFunc("/plans/{planId}/pipelines/{pipelineId}", s.HandleDeletePipeline).Methods("DELETE")

	// Backend component endpoints (plan-scoped)
	api.HandleFunc("/plans/{planId}/backends", s.HandleCreatePlanBackend).Methods("POST")
	api.HandleFunc("/plans/{planId}/backends/{backendId}", s.HandleUpdatePlanBackend).Methods("PUT")
	api.HandleFunc("/plans/{planId}/backends/{backendId}", s.HandleDeletePlanBackend).Methods("DELETE")
	api.HandleFunc("/plans/{planId}/backends/{backendId}/attach", s.HandleAttachBackendToPlan).Methods("PUT")

	// Sandbox endpoints
	api.HandleFunc("/sandboxes", s.HandleListSandboxes).Methods("GET")
	api.HandleFunc("/sandboxes", s.HandleCreateSandbox).Methods("POST")
	api.HandleFunc("/sandboxes/{id}", s.HandleGetSandbox).Methods("GET")
	api.HandleFunc("/sandboxes/{id}", s.HandleDeleteSandbox).Methods("DELETE")
	api.HandleFunc("/sandboxes/{id}/telemetry", s.HandleStartTelemetry).Methods("POST")
	api.HandleFunc("/sandboxes/{id}/validate", s.HandleValidateSandbox).Methods("POST")
	api.HandleFunc("/sandboxes/{id}/logs", s.HandleGetSandboxLogs).Methods("GET")
	api.HandleFunc("/sandboxes/{id}/metrics", s.HandleGetSandboxMetrics).Methods("GET")
	api.HandleFunc("/sandboxes/{id}/stop", s.HandleStopSandbox).Methods("POST")

	// Agent work endpoints
	api.HandleFunc("/agent-work", s.HandleListAgentWork).Methods("GET")
	api.HandleFunc("/agent-work", s.HandleCreateAgentWork).Methods("POST")
	api.HandleFunc("/agent-work/{workId}", s.HandleGetAgentWork).Methods("GET")
	api.HandleFunc("/agent-work/{workId}", s.HandleUpdateAgentWork).Methods("PUT")
	api.HandleFunc("/agent-work/{workId}", s.HandleDeleteAgentWork).Methods("DELETE")
	api.HandleFunc("/agent-work/{workId}/cancel", s.HandleCancelAgentWork).Methods("POST")
	api.HandleFunc("/agent-work/resource/{resourceType}/{resourceId}", s.HandleGetAgentWorkByResource).Methods("GET")

	// Collector endpoints
	api.HandleFunc("/collectors", s.HandleListCollectors).Methods("GET")
	api.HandleFunc("/collectors", s.HandleDeployCollector).Methods("POST")
	api.HandleFunc("/collectors/{id}", s.HandleGetCollector).Methods("GET")
	api.HandleFunc("/collectors/{id}", s.HandleStopCollector).Methods("DELETE")
	api.HandleFunc("/collectors/{id}/config", s.HandleGetCollectorConfig).Methods("GET")
	api.HandleFunc("/collectors/{id}/config", s.HandleUpdateCollectorConfig).Methods("PUT")
	api.HandleFunc("/collectors/{id}/logs", s.HandleGetCollectorLogs).Methods("GET")
	api.HandleFunc("/collectors/connected", s.HandleListConnectedAgents).Methods("GET")

	// Backend endpoints
	api.HandleFunc("/backends", s.HandleListBackends).Methods("GET")
	api.HandleFunc("/backends", s.HandleCreateBackend).Methods("POST")
	api.HandleFunc("/backends/{id}", s.HandleGetBackend).Methods("GET")
	api.HandleFunc("/backends/{id}", s.HandleUpdateBackend).Methods("PUT")
	api.HandleFunc("/backends/{id}", s.HandleDeleteBackend).Methods("DELETE")
	api.HandleFunc("/backends/{id}/test-connection", s.HandleTestConnection).Methods("POST")
	api.HandleFunc("/backends/{id}/configure-datasource", s.HandleConfigureGrafanaDatasource).Methods("POST")

	// Resource delegation endpoint
	api.HandleFunc("/resources/{resourceType}/{resourceId}/delegate", s.HandleDelegate).Methods("POST")

	// Human action endpoints
	api.HandleFunc("/human-actions", s.HandleListHumanActions).Methods("GET")
	api.HandleFunc("/human-actions/pending", s.HandleGetPendingHumanActions).Methods("GET")
	api.HandleFunc("/human-actions/{actionId}", s.HandleGetHumanAction).Methods("GET")
	api.HandleFunc("/human-actions/{actionId}/respond", s.HandleRespondToHumanAction).Methods("POST")
	api.HandleFunc("/human-actions/{actionId}/resume", s.HandleResumeFromHumanAction).Methods("POST")
	api.HandleFunc("/human-actions/{actionId}", s.HandleDeleteHumanAction).Methods("DELETE")

	// Error logging middleware (should be before CORS so we can log errors)
	s.router.Use(errorLoggingMiddleware)

	// CORS middleware
	s.router.Use(corsMiddleware)
}

// responseWriter is a wrapper around http.ResponseWriter that captures status code and body
type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	body          []byte
	headerWritten bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // default status code
		headerWritten:  false,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.headerWritten {
		rw.statusCode = code
		rw.headerWritten = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// If Write is called before WriteHeader, WriteHeader is automatically called with 200
	if !rw.headerWritten {
		rw.WriteHeader(http.StatusOK)
	}

	// Capture the body for error logging (only for errors, and limit size to avoid huge logs)
	if rw.statusCode >= 400 && len(rw.body) < 1024 {
		rw.body = append(rw.body, b...)
	}
	return rw.ResponseWriter.Write(b)
}

// errorLoggingMiddleware logs errors returned by API handlers
func errorLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		// Log errors (4xx and 5xx status codes)
		if rw.statusCode >= 400 {
			errorBody := string(rw.body)
			if errorBody == "" {
				errorBody = http.StatusText(rw.statusCode)
			}
			// Truncate very long error messages for readability
			if len(errorBody) > 500 {
				errorBody = errorBody[:500] + "..."
			}
			log.Printf("[API ERROR] %s %s - Status: %d - Error: %s", r.Method, r.URL.Path, rw.statusCode, errorBody)
		}
	})
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
	// Initialize sandbox tools
	if err := sandboxTools.InitializeSandboxTools(); err != nil {
		log.Printf("Warning: Failed to initialize sandbox tools: %v", err)
	} else {
		log.Printf("Sandbox tools initialized successfully")
	}

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
