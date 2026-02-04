package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/goclaw2/internal/config"
	"github.com/user/goclaw2/internal/memory"
	"github.com/user/goclaw2/internal/provider/zhipu"
	"github.com/user/goclaw2/internal/tools"
)

// Agent represents the AI agent
type Agent struct {
	cfg           *config.Config
	client        *zhipu.Client
	memory        *memory.Store
	tools         *tools.Registry
	maxHistory    int
	contextLoader *ContextLoader
}

// New creates a new agent
func New(cfg *config.Config, mem *memory.Store, toolRegistry *tools.Registry) *Agent {
	return &Agent{
		cfg:           cfg,
		client:        zhipu.New(cfg),
		memory:        mem,
		tools:         toolRegistry,
		maxHistory:    cfg.Agent.MaxHistory,
		contextLoader: NewContextLoader(cfg.Memory.Workspace),
	}
}

// Chat processes a user message and returns the response
func (a *Agent) Chat(userMessage string) (string, error) {
	// Store user message
	if err := a.memory.Add("user", userMessage); err != nil {
		return "", fmt.Errorf("failed to store user message: %w", err)
	}

	// Get conversation history
	messages, err := a.memory.ToProviderFormat(a.maxHistory)
	if err != nil {
		return "", fmt.Errorf("failed to get history: %w", err)
	}

	// Convert to provider message format
	var providerMessages []zhipu.Message

	// Load context files
	contextFiles := a.contextLoader.LoadContextFiles()
	contextPrompt := BuildContextPrompt(contextFiles)

	// Build system prompt
	baseSystemPrompt := `你是一个 AI 助手，拥有文件操作和命令执行能力。

可用工具：
- read_file: 读取文件内容
- write_file: 写入文件
- list_dir: 列出目录内容
- exec_command: 执行 shell 命令

重要规则：
1. 当用户请求读取、写入、列出文件或执行命令时，必须调用相应的工具
2. 不要猜测文件内容，使用 read_file 工具读取
3. 在总结文件操作结果时要准确详细
4. 执行命令前要确认命令的安全性`

	systemContent := baseSystemPrompt
	if contextPrompt != "" {
		systemContent += "\n\n" + contextPrompt
	}

	// Add system message if this is the start of conversation
	if len(messages) == 1 {
		systemMsg := zhipu.Message{
			Role:    "system",
			Content: systemContent,
		}
		providerMessages = []zhipu.Message{systemMsg}

		// Add existing messages
		for _, msg := range messages {
			providerMessages = append(providerMessages, zhipu.Message{
				Role:    msg["role"],
				Content: msg["content"],
			})
		}
	} else {
		// Convert to provider message format (include system prompt at start)
		systemMsg := zhipu.Message{
			Role:    "system",
			Content: systemContent,
		}
		providerMessages = []zhipu.Message{systemMsg}

		// Add existing messages
		for _, msg := range messages {
			providerMessages = append(providerMessages, zhipu.Message{
				Role:    msg["role"],
				Content: msg["content"],
			})
		}
	}

	// Get available tools
	toolList := a.tools.List()
	providerTools := make([]zhipu.Tool, len(toolList))
	for i, tool := range toolList {
		providerTools[i] = zhipu.Tool{
			Type: "function",
			Function: zhipu.ToolFunction{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}
	}

	// Make API call with tools
	resp, err := a.client.ChatWithTools(providerMessages, providerTools)
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	// Handle tool calls
	for resp.HasToolCalls() {
		toolCalls := resp.GetToolCalls()

		// Add assistant response with tool calls to history
		assistantMsg := zhipu.Message{
			Role:      "assistant",
			Content:   resp.GetContent(),
			ToolCalls: toolCalls,
		}
		providerMessages = append(providerMessages, assistantMsg)

		// Execute each tool call
		for _, toolCall := range toolCalls {
			toolName := toolCall.Function.Name
			toolArgs := toolCall.Function.Arguments

			// Execute tool
			result, err := a.tools.ExecuteToolCall(toolName, toolArgs)
			if err != nil {
				result = fmt.Sprintf("Error: %s", err)
			}

			// Add tool result to history
			toolMsg := zhipu.Message{
				Role:    "tool",
				Content: result,
				ToolID:  toolCall.ID,
			}
			providerMessages = append(providerMessages, toolMsg)
		}

		// Make another API call with tool results
		resp, err = a.client.ChatWithTools(providerMessages, providerTools)
		if err != nil {
			return "", fmt.Errorf("API call after tool execution failed: %w", err)
		}
	}

	// Get final response
	response := resp.GetContent()

	// Store assistant response
	if err := a.memory.Add("assistant", response); err != nil {
		return "", fmt.Errorf("failed to store assistant message: %w", err)
	}

	return response, nil
}

// ChatSimple processes a user message without tool support
func (a *Agent) ChatSimple(userMessage string) (string, error) {
	// Store user message
	if err := a.memory.Add("user", userMessage); err != nil {
		return "", fmt.Errorf("failed to store user message: %w", err)
	}

	// Get conversation history
	history, err := a.memory.ToProviderFormat(a.maxHistory)
	if err != nil {
		return "", fmt.Errorf("failed to get history: %w", err)
	}

	// Convert to provider message format
	messages := make([]zhipu.Message, len(history))
	for i, msg := range history {
		messages[i] = zhipu.Message{
			Role:    msg["role"],
			Content: msg["content"],
		}
	}

	// Make API call
	resp, err := a.client.ChatSimple(messages)
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	// Get response
	response := resp.GetContent()

	// Store assistant response
	if err := a.memory.Add("assistant", response); err != nil {
		return "", fmt.Errorf("failed to store assistant message: %w", err)
	}

	return response, nil
}

// SetMaxHistory sets the maximum history length
func (a *Agent) SetMaxHistory(max int) {
	a.maxHistory = max
}

// GetMemory returns the memory store
func (a *Agent) GetMemory() *memory.Store {
	return a.memory
}

// ClearHistory clears the conversation history
func (a *Agent) ClearHistory() error {
	return a.memory.Clear()
}

// SaveConversation saves the current conversation to a markdown file
func (a *Agent) SaveConversation(title string) (string, error) {
	// Create conversations directory if not exists
	conversationsDir := filepath.Join(a.cfg.Memory.Workspace, "memory", "conversations")
	if err := os.MkdirAll(conversationsDir, 0755); err != nil {
		return "", fmt.Errorf("创建 conversations 目录失败: %w", err)
	}

	// Generate filename from title or timestamp
	var filename string
	if title != "" {
		// Sanitize title: remove special characters, replace spaces with dashes
		safeTitle := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, title)
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("%s-%s.md", timestamp, safeTitle)
	} else {
		filename = time.Now().Format("20060102-150405.md")
	}

	// Get conversation history
	messages, err := a.memory.GetHistory(1000) // Get all messages
	if err != nil {
		return "", fmt.Errorf("获取对话历史失败: %w", err)
	}

	// Build markdown content
	var content strings.Builder

	// Header
	content.WriteString(fmt.Sprintf("# %s\n\n", titleOrDefault(title, "对话记录")))
	content.WriteString(fmt.Sprintf("**时间**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("**消息数**: %d\n\n", len(messages)))
	content.WriteString("---\n\n")

	// Messages
	for _, msg := range messages {
		role := msg.Role
		if role == "user" {
			content.WriteString(fmt.Sprintf("### 用户 [%s]\n\n", msg.Timestamp.Format("15:04")))
		} else {
			content.WriteString(fmt.Sprintf("### GoClaw [%s]\n\n", msg.Timestamp.Format("15:04")))
		}
		content.WriteString(msg.Content)
		content.WriteString("\n\n")
	}

	// Write file
	filePath := filepath.Join(conversationsDir, filename)
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return filePath, nil
}

func titleOrDefault(title, defaultVal string) string {
	if title == "" {
		return defaultVal
	}
	return title
}
