package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/config"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorWhite  = "\033[97m"

	colorBold      = "\033[1m"
	colorDim       = "\033[2m"
	colorUnderline = "\033[4m"
)

// Logger handles debug logging for the agent
type Logger struct {
	level config.LogLevel
}

// NewLogger creates a new logger with the specified level
func NewLogger(level config.LogLevel) *Logger {
	return &Logger{level: level}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= config.LogLevelDebug {
		fmt.Printf(colorGray+"[DEBUG] "+format+colorReset+"\n", args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= config.LogLevelInfo {
		fmt.Printf(colorCyan+"[INFO] "+format+colorReset+"\n", args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level >= config.LogLevelError {
		fmt.Printf(colorRed+"[ERROR] "+format+colorReset+"\n", args...)
	}
}

// LogIteration logs the start of a new iteration
func (l *Logger) LogIteration(iteration int, totalMessages int) {
	if l.level >= config.LogLevelDebug {
		fmt.Printf("\n%s%s=== Iteration %d ===%s (Messages in history: %d)\n",
			colorBold, colorBlue, iteration, colorReset, totalMessages)
	}
}

// LogAPIRequest logs an API request
func (l *Logger) LogAPIRequest(model anthropic.Model, maxTokens int64, toolCount int) {
	if l.level >= config.LogLevelDebug {
		fmt.Printf("%s📤 API Request:%s\n", colorYellow, colorReset)
		fmt.Printf("  Model: %s\n", model)
		fmt.Printf("  Max Tokens: %d\n", maxTokens)
		fmt.Printf("  Tools Available: %d\n", toolCount)
	}
}

// LogAPIResponse logs an API response
func (l *Logger) LogAPIResponse(message *anthropic.Message) {
	if l.level < config.LogLevelDebug {
		return
	}

	fmt.Printf("\n%s📥 API Response:%s\n", colorGreen, colorReset)
	fmt.Printf("  Stop Reason: %s%s%s\n", colorBold, message.StopReason, colorReset)
	fmt.Printf("  Model: %s\n", message.Model)
	fmt.Printf("  Usage: Input=%d tokens, Output=%d tokens\n",
		message.Usage.InputTokens, message.Usage.OutputTokens)

	fmt.Printf("\n%s  Content Blocks:%s\n", colorCyan, colorReset)
	for i, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Printf("    %s[%d] Text:%s\n", colorWhite, i, colorReset)
			// Truncate long text for readability
			text := v.Text
			if len(text) > 200 {
				text = text[:200] + "..."
			}
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				fmt.Printf("      %s\n", line)
			}
		case anthropic.ToolUseBlock:
			fmt.Printf("    %s[%d] Tool Use:%s\n", colorPurple, i, colorReset)
			fmt.Printf("      ID: %s\n", v.ID)
			fmt.Printf("      Name: %s%s%s\n", colorBold, v.Name, colorReset)
			fmt.Printf("      Input: %s\n", formatJSON(v.Input))
		}
	}
}

// LogToolExecution logs tool execution
func (l *Logger) LogToolExecution(toolName string, input json.RawMessage) {
	if l.level >= config.LogLevelDebug {
		fmt.Printf("\n%s🔧 Executing Tool:%s %s%s%s\n",
			colorYellow, colorReset, colorBold, toolName, colorReset)
		fmt.Printf("  Input: %s\n", formatJSON(input))
	}
}

// LogToolResult logs tool result
func (l *Logger) LogToolResult(toolName string, result interface{}, err error) {
	if l.level < config.LogLevelDebug {
		return
	}

	if err != nil {
		fmt.Printf("%s❌ Tool Failed:%s %s\n", colorRed, colorReset, toolName)
		fmt.Printf("  Error: %s%s%s\n", colorRed, err.Error(), colorReset)
	} else {
		fmt.Printf("%s✅ Tool Success:%s %s\n", colorGreen, colorReset, toolName)
		resultJSON, _ := json.MarshalIndent(result, "  ", "  ")
		fmt.Printf("  Result: %s\n", string(resultJSON))
	}
}

// LogTaskComplete logs task completion
func (l *Logger) LogTaskComplete(iterations int, toolCalls int) {
	if l.level >= config.LogLevelInfo {
		fmt.Printf("\n%s%s✨ Task Completed!%s\n", colorBold, colorGreen, colorReset)
		fmt.Printf("  Iterations: %d\n", iterations)
		fmt.Printf("  Tool Calls: %d\n", toolCalls)
	}
}

// LogTaskFailed logs task failure
func (l *Logger) LogTaskFailed(err error, iterations int) {
	if l.level >= config.LogLevelError {
		fmt.Printf("\n%s%s❌ Task Failed%s\n", colorBold, colorRed, colorReset)
		fmt.Printf("  Error: %s\n", err.Error())
		fmt.Printf("  Iterations completed: %d\n", iterations)
	}
}

// LogStopReason logs the stop reason
func (l *Logger) LogStopReason(reason string) {
	if l.level >= config.LogLevelDebug {
		color := colorGreen
		icon := "✓"

		switch reason {
		case "end_turn":
			color = colorGreen
			icon = "✓"
		case "tool_use":
			color = colorYellow
			icon = "🔧"
		case "max_tokens":
			color = colorRed
			icon = "⚠"
		default:
			color = colorRed
			icon = "?"
		}

		fmt.Printf("%s%s Stop Reason: %s%s\n", color, icon, reason, colorReset)
	}
}

// LogMessageHistory logs the current message history
func (l *Logger) LogMessageHistory(messages []anthropic.MessageParam) {
	if l.level < config.LogLevelDebug {
		return
	}

	fmt.Printf("\n%s📜 Message History:%s (%d messages)\n", colorGray, colorReset, len(messages))
	for i, msg := range messages {
		role := "user"
		if msg.Role != "" {
			role = string(msg.Role)
		}

		contentCount := 0
		if msg.Content != nil {
			// Count content blocks (this is a rough estimate)
			contentJSON, _ := json.Marshal(msg.Content)
			contentCount = strings.Count(string(contentJSON), `"type"`)
		}

		fmt.Printf("  [%d] %s%s%s - %d content block(s)\n",
			i, colorCyan, role, colorReset, contentCount)
	}
}

// formatJSON formats JSON for display
func formatJSON(data json.RawMessage) string {
	var pretty interface{}
	if err := json.Unmarshal(data, &pretty); err != nil {
		return string(data)
	}
	formatted, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return string(data)
	}
	return string(formatted)
}
