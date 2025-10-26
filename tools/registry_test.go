package tools

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// TestNewRegistry verifies that a new registry is created with empty tools
func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if len(registry.tools) != 0 {
		t.Errorf("Expected empty registry, got %d tools", len(registry.tools))
	}
}

// TestRegisterTool verifies basic tool registration
func TestRegisterTool(t *testing.T) {
	registry := NewRegistry()

	// Define a simple test input type
	type TestInput struct {
		Name string `json:"name"`
	}

	// Define a handler that uses the test input
	handler := func(input TestInput) (interface{}, error) {
		return map[string]string{"greeting": "Hello, " + input.Name}, nil
	}

	// Create schema
	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The name to greet",
			},
		},
		Required: []string{"name"},
	}

	// Register the tool
	Register(registry, "greet", "Greets a person by name", schema, handler)

	// Verify tool is registered
	if !registry.HasTool("greet") {
		t.Error("Tool 'greet' was not registered")
	}

	// Verify tool is in list
	tools := registry.ListTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	if tools[0] != "greet" {
		t.Errorf("Expected tool name 'greet', got '%s'", tools[0])
	}
}

// TestExecuteTool verifies tool execution
func TestExecuteTool(t *testing.T) {
	registry := NewRegistry()

	type TestInput struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	handler := func(input TestInput) (interface{}, error) {
		return input.A + input.B, nil
	}

	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"a": map[string]interface{}{"type": "number"},
			"b": map[string]interface{}{"type": "number"},
		},
		Required: []string{"a", "b"},
	}

	Register(registry, "add", "Adds two numbers", schema, handler)

	// Create input JSON
	inputJSON := json.RawMessage(`{"a": 5, "b": 3}`)

	// Execute the tool
	result, err := registry.Execute("add", inputJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify result
	sum, ok := result.(int)
	if !ok {
		t.Fatalf("Expected int result, got %T", result)
	}
	if sum != 8 {
		t.Errorf("Expected sum of 8, got %d", sum)
	}
}

// TestExecuteUnknownTool verifies error handling for unknown tools
func TestExecuteUnknownTool(t *testing.T) {
	registry := NewRegistry()

	inputJSON := json.RawMessage(`{}`)
	_, err := registry.Execute("nonexistent", inputJSON)

	if err == nil {
		t.Error("Expected error for unknown tool, got nil")
	}

	expectedMsg := "unknown tool: nonexistent"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestExecuteWithInvalidJSON verifies error handling for invalid JSON
func TestExecuteWithInvalidJSON(t *testing.T) {
	registry := NewRegistry()

	type TestInput struct {
		Value int `json:"value"`
	}

	handler := func(input TestInput) (interface{}, error) {
		return input.Value * 2, nil
	}

	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"value": map[string]interface{}{"type": "number"},
		},
	}

	Register(registry, "double", "Doubles a number", schema, handler)

	// Try to execute with invalid JSON
	invalidJSON := json.RawMessage(`{invalid json}`)
	_, err := registry.Execute("double", invalidJSON)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestExecuteWithHandlerError verifies error propagation from handlers
func TestExecuteWithHandlerError(t *testing.T) {
	registry := NewRegistry()

	type TestInput struct {
		Value int `json:"value"`
	}

	expectedErr := errors.New("handler error")
	handler := func(input TestInput) (interface{}, error) {
		return nil, expectedErr
	}

	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"value": map[string]interface{}{"type": "number"},
		},
	}

	Register(registry, "error_tool", "A tool that returns an error", schema, handler)

	inputJSON := json.RawMessage(`{"value": 42}`)
	result, err := registry.Execute("error_tool", inputJSON)

	if err != expectedErr {
		t.Errorf("Expected handler error to be propagated, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got: %v", result)
	}
}

// TestRegisterTools verifies batch tool registration
func TestRegisterTools(t *testing.T) {
	registry := NewRegistry()

	type Input1 struct{ Value int `json:"value"` }
	type Input2 struct{ Text string `json:"text"` }

	tool1 := Tool{
		Name:        "tool1",
		Description: "First tool",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{},
		},
		Handler: func(input json.RawMessage) (interface{}, error) {
			return "result1", nil
		},
	}

	tool2 := Tool{
		Name:        "tool2",
		Description: "Second tool",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{},
		},
		Handler: func(input json.RawMessage) (interface{}, error) {
			return "result2", nil
		},
	}

	registry.RegisterTools([]Tool{tool1, tool2})

	if !registry.HasTool("tool1") {
		t.Error("tool1 not registered")
	}
	if !registry.HasTool("tool2") {
		t.Error("tool2 not registered")
	}

	tools := registry.ListTools()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

// TestGetToolParams verifies conversion to Anthropic tool parameters
func TestGetToolParams(t *testing.T) {
	registry := NewRegistry()

	type TestInput struct {
		Query string `json:"query"`
	}

	handler := func(input TestInput) (interface{}, error) {
		return "search results for: " + input.Query, nil
	}

	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
		},
		Required: []string{"query"},
	}

	Register(registry, "search", "Searches for content", schema, handler)

	params := registry.GetToolParams()

	if len(params) != 1 {
		t.Fatalf("Expected 1 tool param, got %d", len(params))
	}

	param := params[0]
	if param.Name != "search" {
		t.Errorf("Expected tool name 'search', got '%s'", param.Name)
	}

	// Verify the param has a schema
	if param.InputSchema.Properties == nil {
		t.Error("Tool schema not set correctly")
	}
}

// TestGetToolUnionParams verifies conversion to union parameters
func TestGetToolUnionParams(t *testing.T) {
	registry := NewRegistry()

	type TestInput struct{}
	handler := func(input TestInput) (interface{}, error) {
		return "pong", nil
	}

	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{},
	}

	Register(registry, "ping", "Responds with pong", schema, handler)

	unionParams := registry.GetToolUnionParams()

	if len(unionParams) != 1 {
		t.Fatalf("Expected 1 union param, got %d", len(unionParams))
	}

	if unionParams[0].OfTool == nil {
		t.Error("Union param OfTool is nil")
	}
}

// TestHasTool verifies tool existence checking
func TestHasTool(t *testing.T) {
	registry := NewRegistry()

	if registry.HasTool("nonexistent") {
		t.Error("HasTool returned true for nonexistent tool")
	}

	type TestInput struct{}
	handler := func(input TestInput) (interface{}, error) { return nil, nil }
	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{},
	}

	Register(registry, "test", "Test tool", schema, handler)

	if !registry.HasTool("test") {
		t.Error("HasTool returned false for registered tool")
	}
	if registry.HasTool("other") {
		t.Error("HasTool returned true for different tool")
	}
}

// TestRegisterFunc verifies reflection-based registration
func TestRegisterFunc(t *testing.T) {
	registry := NewRegistry()

	type MathInput struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	multiplyFunc := func(input MathInput) (interface{}, error) {
		return input.X * input.Y, nil
	}

	schema := anthropic.ToolInputSchemaParam{
		Properties: map[string]interface{}{
			"x": map[string]interface{}{"type": "number"},
			"y": map[string]interface{}{"type": "number"},
		},
	}

	err := RegisterFunc(registry, "multiply", "Multiplies two numbers", schema, multiplyFunc)
	if err != nil {
		t.Fatalf("RegisterFunc failed: %v", err)
	}

	// Test execution
	inputJSON := json.RawMessage(`{"x": 6, "y": 7}`)
	result, err := registry.Execute("multiply", inputJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	product, ok := result.(int)
	if !ok || product != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}
