package agent

import (
	"context"
	"encoding/json"

	"liki/internal/llm"
)

// LLMClient is the subset of *llm.Client used by the agent. Mock in tests.
type LLMClient interface {
	ChatStreamWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDef) (<-chan llm.StreamEvent, error)
	ChatStream(ctx context.Context, systemPrompt, userMessage string) (<-chan string, error)
}

// ToolRegistry executes named tools and provides their LLM schemas.
type ToolRegistry interface {
	Execute(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error)
	Schemas() []llm.ToolDef
}

// ChatAgent handles LLM conversation via tool-calling. Engine computation is
// delegated to *engine.Service injected through the tool registry.
type ChatAgent struct {
	llm    LLMClient
	tools  ToolRegistry
	prompt string
}

// NewChatAgent creates a new ChatAgent with the given prompt and tools.
func NewChatAgent(llmClient LLMClient, tools ToolRegistry, prompt string) *ChatAgent {
	return &ChatAgent{llm: llmClient, tools: tools, prompt: prompt}
}
