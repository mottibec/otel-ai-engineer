package tools

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/anthropics/anthropic-sdk-go"
)

// ToolHandler is a function that takes the raw JSON input and returns a result
type ToolHandler func(inputJSON json.RawMessage) (interface{}, error)

// ToolDefinition combines a tool's schema and handler
type ToolDefinition struct {
	Name        string
	Description string
	Schema      anthropic.ToolInputSchemaParam
	Handler     ToolHandler
}

// Tool represents a single tool definition that can be registered
type Tool struct {
	Name        string
	Description string
	Schema      anthropic.ToolInputSchemaParam
	Handler     ToolHandler
}

// RegisterTool registers a Tool into the registry
func (r *ToolRegistry) RegisterTool(tool Tool) {
	r.tools[tool.Name] = &ToolDefinition{
		Name:        tool.Name,
		Description: tool.Description,
		Schema:      tool.Schema,
		Handler:     tool.Handler,
	}
}

// RegisterTools registers multiple tools into the registry
func (r *ToolRegistry) RegisterTools(tools []Tool) {
	for _, tool := range tools {
		r.RegisterTool(tool)
	}
}

// ToolRegistry manages tool registrations and invocations
type ToolRegistry struct {
	tools map[string]*ToolDefinition
}

// NewRegistry creates a new tool registry
func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*ToolDefinition),
	}
}

// Register adds a tool with its schema and handler to the registry
// The handler function will be called with the parsed input struct
func Register[T any](r *ToolRegistry, toolName string, description string, schema anthropic.ToolInputSchemaParam, handler func(T) (interface{}, error)) {
	r.tools[toolName] = &ToolDefinition{
		Name:        toolName,
		Description: description,
		Schema:      schema,
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input T
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input for tool %s: %w", toolName, err)
			}
			return handler(input)
		},
	}
}

// RegisterFunc is a convenience method for registering tools using function reflection
// The function must accept a single struct parameter and return (interface{}, error)
func RegisterFunc(r *ToolRegistry, toolName string, description string, schema anthropic.ToolInputSchemaParam, handlerFunc interface{}) error {
	funcValue := reflect.ValueOf(handlerFunc)
	funcType := funcValue.Type()

	// Validate function signature
	if funcType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}
	if funcType.NumIn() != 1 {
		return fmt.Errorf("handler must accept exactly one parameter")
	}
	if funcType.NumOut() != 2 {
		return fmt.Errorf("handler must return exactly two values (interface{}, error)")
	}

	inputType := funcType.In(0)

	r.tools[toolName] = &ToolDefinition{
		Name:        toolName,
		Description: description,
		Schema:      schema,
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			// Create new instance of input type
			inputValue := reflect.New(inputType)

			// Unmarshal JSON into the input struct
			if err := json.Unmarshal(inputJSON, inputValue.Interface()); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input for tool %s: %w", toolName, err)
			}

			// Call the handler function
			results := funcValue.Call([]reflect.Value{inputValue.Elem()})

			// Extract return values
			var resultErr error
			if !results[1].IsNil() {
				resultErr = results[1].Interface().(error)
			}

			return results[0].Interface(), resultErr
		},
	}

	return nil
}

// Execute runs a tool by name with the given input
func (r *ToolRegistry) Execute(toolName string, inputJSON json.RawMessage) (interface{}, error) {
	tool, exists := r.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	return tool.Handler(inputJSON)
}

// ExecuteToolUseBlock processes a ToolUseBlock and returns the result
func (r *ToolRegistry) ExecuteToolUseBlock(block anthropic.ToolUseBlock) (interface{}, error) {
	return r.Execute(block.Name, block.Input)
}

// HasTool checks if a tool is registered
func (r *ToolRegistry) HasTool(toolName string) bool {
	_, exists := r.tools[toolName]
	return exists
}

// ListTools returns all registered tool names
func (r *ToolRegistry) ListTools() []string {
	toolNames := make([]string, 0, len(r.tools))
	for name := range r.tools {
		toolNames = append(toolNames, name)
	}
	return toolNames
}

// GetToolParams returns Anthropic tool parameters for all registered tools
// This is used to pass to the Messages API to inform Claude what tools are available
func (r *ToolRegistry) GetToolParams() []anthropic.ToolParam {
	params := make([]anthropic.ToolParam, 0, len(r.tools))
	for _, tool := range r.tools {
		params = append(params, anthropic.ToolParam{
			Name:        tool.Name,
			Description: anthropic.String(tool.Description),
			InputSchema: tool.Schema,
		})
	}
	return params
}

// GetToolUnionParams returns Anthropic tool union parameters for all registered tools
// This is a convenience wrapper around GetToolParams for the Messages API
func (r *ToolRegistry) GetToolUnionParams() []anthropic.ToolUnionParam {
	toolParams := r.GetToolParams()
	tools := make([]anthropic.ToolUnionParam, len(toolParams))
	for i, toolParam := range toolParams {
		tp := toolParam // Create a copy to avoid pointer issues
		tools[i] = anthropic.ToolUnionParam{OfTool: &tp}
	}
	return tools
}
