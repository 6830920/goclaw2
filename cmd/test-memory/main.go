package main

import (
	"fmt"
	"os"

	"github.com/user/goclaw2/internal/config"
	"github.com/user/goclaw2/internal/tools"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Test memory_search tool
	tool := &tools.MemorySearch{WorkspaceDir: cfg.Memory.Workspace}

	fmt.Printf("Testing memory_search tool...\n")
	fmt.Printf("Workspace: %s\n\n", cfg.Memory.Workspace)

	// Execute with search query "Go"
	result, err := tool.Execute(map[string]interface{}{
		"query": "Go",
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Result:\n%s\n", result)
}
