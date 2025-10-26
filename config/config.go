package config

import (
	"fmt"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogLevelSilent:
		return "silent"
	case LogLevelError:
		return "error"
	case LogLevelInfo:
		return "info"
	case LogLevelDebug:
		return "debug"
	default:
		return "unknown"
	}
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "silent":
		return LogLevelSilent
	case "error":
		return LogLevelError
	case "info":
		return LogLevelInfo
	case "debug":
		return LogLevelDebug
	default:
		return LogLevelInfo // default
	}
}

// Config holds the application configuration
type Config struct {
	// Anthropic API key
	AnthropicAPIKey string

	// Model to use (default: claude-3-5-sonnet-20241022)
	Model string

	// Max tokens for responses
	MaxTokens int

	// Log level for debugging
	LogLevel LogLevel
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Default to Claude 3.5 Sonnet
	}

	maxTokens := 4096 // Default
	if mt := os.Getenv("MAX_TOKENS"); mt != "" {
		fmt.Sscanf(mt, "%d", &maxTokens)
	}

	logLevel := ParseLogLevel(os.Getenv("LOG_LEVEL"))

	return &Config{
		AnthropicAPIKey: apiKey,
		Model:           model,
		MaxTokens:       maxTokens,
		LogLevel:        logLevel,
	}, nil
}
