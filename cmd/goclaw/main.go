package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/goclaw2/internal/agent"
	"github.com/user/goclaw2/internal/config"
	"github.com/user/goclaw2/internal/memory"
	"github.com/user/goclaw2/internal/tools"
)

var (
	cfgFile string
	cfg     *config.Config
	agt     *agent.Agent
	mem     *memory.Store
	toolReg *tools.Registry
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "goclaw",
		Short: "GoClaw2 - AI Assistant with Tool Support",
		Long: `GoClaw2 is a Go-based AI assistant that supports conversations,
tool execution (file operations, commands), and memory management.`,
		PersistentPreRunE: prerun,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	var chatCmd = &cobra.Command{
		Use:   "chat",
		Short: "Start interactive chat",
		RunE:  runChat,
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		RunE:  runConfig,
	}

	var memoryCmd = &cobra.Command{
		Use:   "memory",
		Short: "Memory management commands",
	}

	var memoryClearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear conversation history",
		RunE:  runMemoryClear,
	}

	var memoryShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show conversation history",
		RunE:  runMemoryShow,
	}

	memoryCmd.AddCommand(memoryClearCmd, memoryShowCmd)
	rootCmd.AddCommand(chatCmd, configCmd, memoryCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func prerun(cmd *cobra.Command, args []string) error {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize memory store
	sessionID := "default"
	mem, err = memory.New(cfg.Memory.FilePath, sessionID)
	if err != nil {
		return fmt.Errorf("failed to initialize memory: %w", err)
	}

	// Initialize tool registry
	toolReg = tools.New()
	toolReg.Register(&tools.ReadFile{})
	toolReg.Register(&tools.WriteFile{})
	toolReg.Register(&tools.ListDir{})
	toolReg.Register(&tools.ExecCommand{})

	// Register memory tools with workspace path
	workspaceDir := cfg.Memory.Workspace
	toolReg.Register(&tools.MemorySearch{WorkspaceDir: workspaceDir})
	toolReg.Register(&tools.MemoryGet{WorkspaceDir: workspaceDir})
	toolReg.Register(&tools.UpdateMemory{WorkspaceDir: workspaceDir})

	// Initialize agent
	agt = agent.New(cfg, mem, toolReg)

	return nil
}

func runChat(cmd *cobra.Command, args []string) error {
	color.Cyan("╔════════════════════════════════════════╗")
	color.Cyan("║        GoClaw2 - AI Assistant         ║")
	color.Cyan("║   Powered by Zhipu GLM-4              ║")
	color.Cyan("╚════════════════════════════════════════╝")
	color.White("\nCommands:")
	color.White("  /clear  - Clear conversation history")
	color.White("  /quit   - Exit")
	color.White("  /help   - Show available tools")
	color.White("\nType your message and press Enter.\n")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		color.Yellow("\n\nShutting down gracefully...")
		if mem != nil {
			mem.Close()
		}
		os.Exit(0)
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		color.Green("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			if err := handleCommand(input); err != nil {
				return err
			}
			continue
		}

		// Process with agent
		color.Yellow("Thinking...")
		response, err := agt.Chat(input)
		if err != nil {
			color.Red("\nError: %v\n", err)
			continue
		}

		fmt.Print("\r") // Clear "Thinking..."
		color.Cyan("AI: %s\n\n", response)
	}
}

func handleCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "/quit", "/exit":
		color.Yellow("Goodbye!")
		if mem != nil {
			mem.Close()
		}
		os.Exit(0)

	case "/clear":
		if err := agt.ClearHistory(); err != nil {
			return fmt.Errorf("failed to clear history: %w", err)
		}
		color.Yellow("Conversation history cleared.")

	case "/help":
		color.Yellow("\nAvailable Tools:")
		for _, tool := range toolReg.List() {
			color.White("  • %s - %s", tool.Name(), tool.Description())
		}
		color.White("\n")

	default:
		color.Yellow("Unknown command: %s", parts[0])
		color.Yellow("Available: /quit, /clear, /help")
	}

	return nil
}

func runConfig(cmd *cobra.Command, args []string) error {
	color.Yellow("Current Configuration:")
	color.White("  Zhipu API Key: %s", maskAPIKey(cfg.Zhipu.APIKey))
	color.White("  Zhipu Model: %s", cfg.Zhipu.Model)
	color.White("  Temperature: %.2f", cfg.Zhipu.Temperature)
	color.White("  Max Tokens: %d", cfg.Zhipu.MaxTokens)
	color.White("  Memory Path: %s", cfg.Memory.FilePath)
	color.White("  Max History: %d", cfg.Agent.MaxHistory)

	count, err := mem.Count()
	if err == nil {
		color.White("  Message Count: %d", count)
	}

	return nil
}

func runMemoryClear(cmd *cobra.Command, args []string) error {
	if err := agt.ClearHistory(); err != nil {
		return fmt.Errorf("failed to clear memory: %w", err)
	}
	color.Yellow("✓ Conversation history cleared")
	return nil
}

func runMemoryShow(cmd *cobra.Command, args []string) error {
	messages, err := mem.GetHistory(100)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if len(messages) == 0 {
		color.Yellow("No messages in history")
		return nil
	}

	color.Yellow("\nConversation History (%d messages):", len(messages))
	color.White("─────────────────────────────────────")

	for _, msg := range messages {
		if msg.Role == "user" {
			color.Green("\n[User] %s", msg.Content)
		} else if msg.Role == "assistant" {
			color.Cyan("\n[AI] %s", msg.Content)
		}
	}

	color.White("\n─────────────────────────────────────\n")
	return nil
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
