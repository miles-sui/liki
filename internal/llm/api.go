// Package llm provides a DeepSeek LLM client with streaming and tool calling.
//
// Client
//
//	Client                   — DeepSeek API client
//	New                      — constructor with default model (deepseek-v4-pro)
//	ChatStream               — streaming chat, text-only
//	ChatStreamWithTools      — streaming chat with tool calling support
//
// Messages
//
//	Role                     — system / user / assistant / tool
//	RoleSystem, RoleUser, RoleAssistant, RoleTool
//	Message                  — chat message with optional tool calls
//	ToolCall                 — function call requested by LLM
//	FunctionCall             — function name + JSON arguments
//
// Responses
//
//	ChatResult               — LLM response: content + tool calls
//	StreamEvent              — streaming chunk: text delta + reasoning + finish reason
//	ToolDef                  — tool definition passed to LLM API
//
package llm
