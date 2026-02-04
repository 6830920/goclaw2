package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadFile reads the content of a file
type ReadFile struct{}

func (t *ReadFile) Name() string {
	return "read_file"
}

func (t *ReadFile) Description() string {
	return "Read the content of a file at the specified path"
}

func (t *ReadFile) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute or relative path to the file",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFile) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path argument is required")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}

	// Read file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// WriteFile writes content to a file
type WriteFile struct{}

func (t *WriteFile) Name() string {
	return "write_file"
}

func (t *WriteFile) Description() string {
	return "Write content to a file at the specified path. Creates the file if it doesn't exist, overwrites if it does."
}

func (t *WriteFile) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute or relative path to the file",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFile) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path argument is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content argument is required")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// ListDir lists the contents of a directory
type ListDir struct{}

func (t *ListDir) Name() string {
	return "list_dir"
}

func (t *ListDir) Description() string {
	return "List the contents of a directory at the specified path"
}

func (t *ListDir) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute or relative path to the directory. Defaults to current directory.",
			},
		},
	}
}

func (t *ListDir) Execute(args map[string]interface{}) (string, error) {
	path := "."
	if p, ok := args["path"].(string); ok {
		path = p
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Read directory
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	result := fmt.Sprintf("Contents of %s:\n", absPath)
	for _, entry := range entries {
		if entry.IsDir() {
			result += fmt.Sprintf("  [DIR]  %s\n", entry.Name())
		} else {
			result += fmt.Sprintf("  [FILE] %s\n", entry.Name())
		}
	}

	return result, nil
}
