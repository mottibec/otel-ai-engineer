package sandbox

import (
	"fmt"
	"log"
)

// SimpleLogger is a basic logger implementation
type SimpleLogger struct {
	prefix string
}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger(prefix string) *SimpleLogger {
	return &SimpleLogger{
		prefix: prefix,
	}
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, fields map[string]interface{}) {
	log.Printf("[INFO] [%s] %s %s\n", l.prefix, msg, formatFields(fields))
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, err error, fields map[string]interface{}) {
	errMsg := ""
	if err != nil {
		errMsg = fmt.Sprintf(" error=%v", err)
	}
	log.Printf("[ERROR] [%s] %s%s %s\n", l.prefix, msg, errMsg, formatFields(fields))
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, fields map[string]interface{}) {
	log.Printf("[DEBUG] [%s] %s %s\n", l.prefix, msg, formatFields(fields))
}

func formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	result := ""
	for k, v := range fields {
		result += fmt.Sprintf("%s=%v ", k, v)
	}
	return result
}
