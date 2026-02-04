package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContextFile represents a context file with its path and content
type ContextFile struct {
	Path    string
	Content string
}

// ContextLoader handles loading context files from workspace
type ContextLoader struct {
	workspaceDir string
}

// NewContextLoader creates a new context loader
func NewContextLoader(workspaceDir string) *ContextLoader {
	// Expand ~ to home directory
	if len(workspaceDir) > 0 && workspaceDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			workspaceDir = filepath.Join(home, workspaceDir[1:])
		}
	}

	return &ContextLoader{
		workspaceDir: workspaceDir,
	}
}

// LoadContextFiles loads all context files from workspace
func (cl *ContextLoader) LoadContextFiles() []ContextFile {
	files := []ContextFile{}

	// Files to load in priority order
	contextFiles := []string{
		"IDENTITY.md",
		"SOUL.md",
		"memory/MEMORY.md",
	}

	for _, filename := range contextFiles {
		path := filepath.Join(cl.workspaceDir, filename)
		if content, err := os.ReadFile(path); err == nil {
			files = append(files, ContextFile{
				Path:    filename,
				Content: string(content),
			})
		}
		// File doesn't exist is OK, skip silently
	}

	return files
}

// LoadContextFile loads a specific context file
func (cl *ContextLoader) LoadContextFile(filename string) (ContextFile, error) {
	path := filepath.Join(cl.workspaceDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return ContextFile{}, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	return ContextFile{
		Path:    filename,
		Content: string(content),
	}, nil
}

// BuildContextPrompt builds the prompt section for context files
func BuildContextPrompt(files []ContextFile) string {
	if len(files) == 0 {
		return ""
	}

	var sections []string
	sections = append(sections, "## Workspace 上下文文件")
	sections = append(sections, "")
	sections = append(sections, "以下文件已加载，提供了我的身份和记忆：")
	sections = append(sections, "")

	for _, file := range files {
		sections = append(sections, fmt.Sprintf("### %s", file.Path))
		sections = append(sections, "")
		sections = append(sections, file.Content)
		sections = append(sections, "")
	}

	return strings.Join(sections, "\n")
}
