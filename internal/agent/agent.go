package agent

import (
	"context"
	"encoding/json"
	"strings"

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

// OrderCreator creates an order in the payment store.
type OrderCreator interface {
	CreateOrder(ctx context.Context, orderID string, product Product, amount int, currency, chartJSON, llmJSON, locale string) error
	UpdateEmail(ctx context.Context, orderID, email string) error
}

// Product identifies a paid report product.
type Product string

const (
	ProductChart  Product = "chart"
	ProductBond   Product = "bond"
	ProductNaming Product = "naming"
)

func (p Product) EmailSubject() string {
	switch p {
	case ProductChart:
		return "您的八字报告"
	case ProductBond:
		return "您的合盘报告"
	case ProductNaming:
		return "您的起名报告"
	default:
		return "您的命理报告"
	}
}

// ChatAgent handles LLM conversation via tool-calling. Engine computation is
// delegated to *engine.Service injected through the tool registry.
type ChatAgent struct {
	llm           LLMClient
	tools         ToolRegistry
	prompt        string
	ReportPrompts map[Product]string
	Amounts       map[Product]int
	Greeting      string
}

// NewChatAgent creates a new ChatAgent with the given prompt and tools.
func NewChatAgent(llmClient LLMClient, tools ToolRegistry, prompt string) *ChatAgent {
	return &ChatAgent{
		llm:    llmClient,
		tools:  tools,
		prompt: prompt,
	}
}

// systemPrompt returns the chat system prompt with {locale} replaced.
func (a *ChatAgent) systemPrompt(locale string) string {
	return strings.ReplaceAll(a.prompt, "{locale}", locale)
}

// ChatResult holds the outcome of a Chat call.
type ChatResult struct {
	Messages []llm.Message // updated conversation history
	Purchase *PurchaseInfo    // nil if purchase not triggered
}

// PurchaseInfo holds order details when purchase is triggered.
type PurchaseInfo struct {
	OrderID string
	Amount  int
	Product Product
}
