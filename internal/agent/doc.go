// Package agent provides the LLM chat agent with tool-calling support.
//
// Interfaces (contracts for dependency injection)
//
//	LLMClient        — streaming chat client, implemented by *llm.Client
//	ToolRegistry     — named tool execution + LLM schema, implemented by *ChatToolRegistry
//	OrderCreator     — payment store operations, implemented by *payment.Store
//
// Chat
//
//	ChatAgent        — holds LLM client, tools, prompt, report prompts, amounts, greeting
//	NewChatAgent     — constructor
//	Chat             — streaming chat: collection → compute → teaser → Q&A → purchase
//	GenerateFromData — non-streaming report generation from pre-computed data
//	ChatResult       — outcome of Chat: updated messages + optional PurchaseInfo
//	PurchaseInfo     — order details triggered by purchase call
//
// Products
//
//	Product                    — chart / bond / naming
//	ProductChart, ProductBond, ProductNaming
//
// Events (SSE streaming)
//
//	ChatEventType              — classification for client-side routing
//	EventTextDelta, EventPhase, EventDone, EventError, EventThinking, EventThinkingDelta
//	ChatEvent                  — structured event emitted during streaming
//
// Test helpers
//
//	MockLLM           — stub LLM client
//	MockToolRegistry  — stub tool registry
//	ToolCall, ChatRes — fixture builders
package agent
