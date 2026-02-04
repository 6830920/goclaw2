package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecCommand executes a shell command
type ExecCommand struct{}

func (t *ExecCommand) Name() string {
	return "exec_command"
}

func (t *ExecCommand) Description() string {
	return "Execute a shell command and return its output. Use with caution."
}

func (t *ExecCommand) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 30)",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ExecCommand) Execute(args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command argument is required")
	}

	timeout := 30
	if to, ok := args["timeout"].(float64); ok {
		timeout = int(to)
	}

	// Split command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Create command with timeout using context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
