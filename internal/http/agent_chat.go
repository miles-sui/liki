package http

import (
	"context"
	"log/slog"

	"liki/internal/agent"
	"liki/internal/llm"
	"liki/internal/payment"
)

// buildChatMessages loads chat history and builds the message list for the LLM.
func buildChatMessages(ctx context.Context, store *payment.Store, orderID, userMessage string) ([]llm.Message, error) {
	history, err := store.LoadChatHistory(ctx, orderID)
	if err != nil {
		return nil, err
	}
	messages := make([]llm.Message, 0, len(history)+1)
	for _, m := range history {
		messages = append(messages, llm.Message{
			Role:    llm.Role(m.Role),
			Content: m.Content,
		})
	}
	messages = append(messages, llm.Message{Role: llm.RoleUser, Content: userMessage})
	return messages, nil
}

// detectAndSaveReport checks if the last assistant message is a naming report
// and saves it to the store. Calls onReport with the redirect URL on success.
func detectAndSaveReport(ctx context.Context, store *payment.Store, orderID string, result []llm.Message, onReport func(url string)) {
	if len(result) == 0 {
		return
	}
	last := result[len(result)-1]
	if last.Role == llm.RoleAssistant && last.Content != "" && len(last.ToolCalls) == 0 && agent.IsNamingReport(last.Content) {
		if err := store.UpdateLlmJSON(ctx, orderID, last.Content); err != nil {
			slog.Warn("naming: save report", "err", err)
		}
		onReport("/report/" + orderID)
	}
}

// persistChatResult saves new assistant messages from the result to the store.
// skipCount is the number of messages before the current turn (system + history + user).
func persistChatResult(ctx context.Context, store *payment.Store, orderID string, result []llm.Message, skipCount int) {
	// result layout: [system prompt] + [history msgs] + [user msg] + [new msgs...]
	// Skip: 1 system + skipCount history + 1 user = skipCount + 2.
	if len(result) <= skipCount+2 {
		return
	}
	batch := make([]payment.ChatMessage, 0, len(result)-(skipCount+2))
	for _, m := range result[skipCount+2:] {
		if m.Content == "" || m.Role == llm.RoleTool {
			continue
		}
		batch = append(batch, payment.ChatMessage{Role: payment.Role(m.Role), Content: m.Content})
	}
	if len(batch) > 0 {
		if err := store.BatchCreateChatMessages(ctx, orderID, batch); err != nil {
			slog.Warn("naming: persist messages", "err", err)
		}
	}
}
