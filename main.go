package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/mottibechhofer/otel-ai-engineer/agent"
	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
	"github.com/mottibechhofer/otel-ai-engineer/config"
	"github.com/mottibechhofer/otel-ai-engineer/server"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Anthropic client with API key from config
	client := anthropic.NewClient(
		option.WithAPIKey(cfg.AnthropicAPIKey),
	)

	// Parse port from command line arguments
	port := 8080
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		}
	}

	// Create SQLite storage
	dbPath := storage.GetDBPath()
	stor, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite storage: %v", err)
	}
	log.Printf("Using SQLite storage at: %s", dbPath)

	// Create event emitter
	emitter := events.NewEmitter()

	// Create event bridge to connect emitter to storage
	bridge := server.NewEventBridge(stor, emitter)
	defer bridge.Close()

	// Create agent registry
	agentRegistry := agent.NewRegistry()

	// Create server
	srv := server.New(server.Config{
		Storage:         stor,
		Port:            port,
		AgentRegistry:   agentRegistry,
		AnthropicClient: &client,
		LogLevel:        cfg.LogLevel,
		EventBridge:     bridge,
	})

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		srv.Close()
		os.Exit(0)
	}()

	// Start server
	fmt.Printf("Starting OpenTelemetry AI Engineer UI server on port %d...\n", port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
