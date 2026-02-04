package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveConversation saves the current conversation to a markdown file
type SaveConversation struct {
	workspaceDir string
}

func (t *SaveConversation) Name() string {
	return "save_conversation"
}

func (t *SaveConversation) Description() string {
	return "保存当前对话到 markdown 文件。使用 LLM 生成描述性文件名。"
}

func (t *SaveConversation) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "对话的简短描述，用于生成文件名（可选）",
			},
		},
	}
}

func (t *SaveConversation) Execute(args map[string]interface{}) (string, error) {
	// 从 memory store 获取对话历史
	// 这里需要通过依赖注入传入，暂时返回说明
	return "save_conversation 工具需要在 agent 中集成完整功能", nil
}

// MemorySearch searches for information in memory files
type MemorySearch struct {
	WorkspaceDir string
}

func (t *MemorySearch) Name() string {
	return "memory_search"
}

func (t *MemorySearch) Description() string {
	return "在记忆文件中搜索相关信息。搜索 MEMORY.md 和 memory/*.md"
}

func (t *MemorySearch) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "搜索关键词或问题",
			},
		},
		"required": []string{"query"},
	}
}

func (t *MemorySearch) Execute(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query 参数是必需的")
	}

	results := t.searchInMemory(query)
	if len(results) == 0 {
		return fmt.Sprintf("未找到关于 '%s' 的记忆", query), nil
	}

	return fmt.Sprintf("找到 %d 条相关记忆：\n%s", len(results), results), nil
}

// searchInMemory performs a simple keyword search in memory files
func (t *MemorySearch) searchInMemory(query string) string {
	memoryDir := filepath.Join(t.WorkspaceDir, "memory")
	results := []string{}

	// Search MEMORY.md
	memoryPath := filepath.Join(memoryDir, "MEMORY.md")
	content, err := os.ReadFile(memoryPath)
	if err == nil {
		// Debug: print file content
		// fmt.Fprintf(os.Stderr, "Searching in %s (length: %d)\n", memoryPath, len(content))
		if containsKeywords(string(content), query) {
			results = append(results, fmt.Sprintf("- MEMORY.md"))
		}
	} else {
		// fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", memoryPath, err)
	}

	// Search conversation files
	conversationsDir := filepath.Join(memoryDir, "conversations")
	if entries, err := os.ReadDir(conversationsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				path := filepath.Join(conversationsDir, entry.Name())
				if content, err := os.ReadFile(path); err == nil {
					if containsKeywords(string(content), query) {
						results = append(results, fmt.Sprintf("- %s", entry.Name()))
					}
				}
			}
		}
	}

	return joinStrings(results, "\n")
}

// containsKeywords checks if text contains any of the query keywords
func containsKeywords(text, query string) bool {
	// Simple case-insensitive substring search
	// In production, you might want more sophisticated matching
	return containsIgnoreCase(text, query)
}

// MemoryGet reads the content of a specific memory file
type MemoryGet struct {
	WorkspaceDir string
}

func (t *MemoryGet) Name() string {
	return "memory_get"
}

func (t *MemoryGet) Description() string {
	return "读取指定记忆文件的完整内容"
}

func (t *MemoryGet) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"filename": map[string]interface{}{
				"type":        "string",
				"description": "文件名，例如 MEMORY.md 或 conversations/xxx.md",
			},
		},
		"required": []string{"filename"},
	}
}

func (t *MemoryGet) Execute(args map[string]interface{}) (string, error) {
	filename, ok := args["filename"].(string)
	if !ok {
		return "", fmt.Errorf("filename 参数是必需的")
	}

	path := filepath.Join(t.WorkspaceDir, "memory", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("无法读取文件 %s: %w", filename, err)
	}

	return string(content), nil
}

// UpdateMemory updates MEMORY.md with new information
type UpdateMemory struct {
	WorkspaceDir string
}

func (t *UpdateMemory) Name() string {
	return "update_memory"
}

func (t *UpdateMemory) Description() string {
	return "更新长期记忆文件 MEMORY.md，添加新的重要信息"
}

func (t *UpdateMemory) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "要添加的内容",
			},
			"section": map[string]interface{}{
				"type":        "string",
				"description": "目标章节（可选），例如：用户偏好、重要事项",
			},
		},
		"required": []string{"content"},
	}
}

func (t *UpdateMemory) Execute(args map[string]interface{}) (string, error) {
	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content 参数是必需的")
	}

	section, _ := args["section"].(string)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	memoryPath := filepath.Join(t.WorkspaceDir, "memory", "MEMORY.md")
	existing, _ := os.ReadFile(memoryPath)

	// Build new entry
	newEntry := fmt.Sprintf("\n## %s\n\n**时间**: %s\n\n%s\n",
		sectionOrDefault(section, "其他"), timestamp, content)

	updatedContent := string(existing) + newEntry
	if err := os.WriteFile(memoryPath, []byte(updatedContent), 0644); err != nil {
		return "", fmt.Errorf("写入 MEMORY.md 失败: %w", err)
	}

	return fmt.Sprintf("✓ 已更新 MEMORY.md [%s]", sectionOrDefault(section, "其他")), nil
}

func sectionOrDefault(section, defaultVal string) string {
	if section == "" {
		return defaultVal
	}
	return section
}

// Helper functions
func containsIgnoreCase(text, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(text) < len(substr) {
		return false
	}

	// Simple case-insensitive substring search
	textLower := strings.ToLower(text)
	substrLower := strings.ToLower(substr)
	return strings.Contains(textLower, substrLower)
}

func contains(text, substr string) bool {
	return len(text) >= len(substr) && indexOf(text, substr) >= 0
}

func indexOf(text, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if i+j >= len(text) || toLower(text[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
