// Package agent provides the LLM chat agent with tool-calling support.
//
// Interfaces (contracts for dependency injection)
//
//	LLMClient        — streaming chat client, implemented by *llm.Client
//	ToolRegistry     — named tool execution + LLM schema, implemented by *ChatToolRegistry
//
// Chat
//
//	ChatAgent        — holds LLM client, tools, prompt
//	NewChatAgent     — constructor
//	NamingChat       — streaming naming chat: collection → computation → report
//
// Products
//
//	ProductNaming    — the only active product
//
// Events (SSE streaming)
//
//	ChatEventType              — classification for client-side routing
//	EventTextDelta, EventPhase, EventError, EventThinking, EventThinkingDelta, EventReportReady
//	ChatEvent                  — structured event emitted during streaming
//
// Test helpers
//
//	MockLLM           — stub LLM client
//	MockToolRegistry  — stub tool registry
//	ToolCall, ChatRes — fixture builders
package agent
