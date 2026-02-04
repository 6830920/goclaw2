package tools

import (
	"encoding/json"
	"fmt"
)

// Tool represents a tool that can be executed
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(args map[string]interface{}) (string, error)
}

// Registry manages available tools
type Registry struct {
	tools map[string]Tool
}

// New creates a new tool registry
func New() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

// ToProviderFormat converts tools to the format expected by AI providers
func (r *Registry) ToProviderFormat() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name(),
				"description": tool.Description(),
				"parameters":  tool.Parameters(),
			},
		})
	}
	return result
}

// ExecuteToolCall executes a tool call with the given arguments
func (r *Registry) ExecuteToolCall(name string, argsJSON string) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := tool.Execute(args)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}
