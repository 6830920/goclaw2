package zhipu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/user/goclaw2/internal/config"
)

// Message represents a chat message
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolID    string     `json:"tool_call_id,omitempty"`
}

// Tool represents a function that can be called
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a call to a tool
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	// Note: tool_calls is inside message, not at choice level
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Client represents the Zhipu AI client
type Client struct {
	cfg       *config.Config
	client    *http.Client
	baseURL   string
	apiKey    string
}

// New creates a new Zhipu AI client
func New(cfg *config.Config) *Client {
	return &Client{
		cfg:     cfg,
		client:  &http.Client{},
		baseURL: cfg.Zhipu.BaseURL,
		apiKey:  cfg.Zhipu.APIKey,
	}
}

// Chat sends a chat completion request
func (c *Client) Chat(req *ChatRequest) (*ChatResponse, error) {
	req.Model = c.cfg.Zhipu.Model
	if req.Temperature == 0 {
		req.Temperature = c.cfg.Zhipu.Temperature
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.cfg.Zhipu.MaxTokens
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// ChatSimple sends a simple text chat request
func (c *Client) ChatSimple(messages []Message) (*ChatResponse, error) {
	return c.Chat(&ChatRequest{
		Messages: messages,
	})
}

// ChatWithTools sends a chat request with available tools
func (c *Client) ChatWithTools(messages []Message, tools []Tool) (*ChatResponse, error) {
	return c.Chat(&ChatRequest{
		Messages: messages,
		Tools:    tools,
	})
}

// HasToolCalls checks if the response contains tool calls
func (r *ChatResponse) HasToolCalls() bool {
	if len(r.Choices) == 0 {
		return false
	}
	return len(r.Choices[0].Message.ToolCalls) > 0
}

// GetContent returns the response content
func (r *ChatResponse) GetContent() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

// GetToolCalls returns tool calls from the response
func (r *ChatResponse) GetToolCalls() []ToolCall {
	if len(r.Choices) == 0 {
		return nil
	}
	return r.Choices[0].Message.ToolCalls
}

// ParseToolCallArgs parses the arguments string from a tool call
func ParseToolCallArgs(args string, target interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(args))
	decoder.UseNumber()
	return decoder.Decode(target)
}
