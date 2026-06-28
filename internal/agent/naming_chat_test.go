package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"liki/internal/llm"
)

func TestNamingChat_TextOnlyResponse(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(nil, "你好，请告诉我你的出生信息"),
		},
	}
	mockTools := &MockToolRegistry{}

	a := NewChatAgent(mockLLM, mockTools, "test prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "你好"},
	}

	result, err := a.NamingChat(context.Background(), "zh-Hans", messages, nil)
	if err != nil {
		t.Fatalf("NamingChat: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("message count = %d, want 3 (system + user + assistant)", len(result))
	}
	if result[0].Role != llm.RoleSystem {
		t.Errorf("result[0].Role = %q, want system", result[0].Role)
	}
	if !strings.Contains(result[0].Content, "zh-Hans") {
		t.Error("system prompt should have locale replaced")
	}
	if result[2].Role != llm.RoleAssistant {
		t.Errorf("result[2].Role = %q, want assistant", result[2].Role)
	}
}

func TestNamingChat_WithToolExecution(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes([]llm.ToolCall{ToolCall("compute_chart", `{"birth":{"time":"2026-06-25T14:00:00+08:00","longitude":116.4},"gender":"male"}`)}, ""),
			ChatRes(nil, "你的八字排盘已完成"),
		},
	}
	mockTools := &MockToolRegistry{
		Results: map[string]json.RawMessage{
			"compute_chart": json.RawMessage(`{"_product":"chart","data":{"nian":{"Gan":"丙","Zhi":"午"}}}`),
		},
	}

	a := NewChatAgent(mockLLM, mockTools, "prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "帮我排八字"},
	}

	result, err := a.NamingChat(context.Background(), "zh-Hans", messages, nil)
	if err != nil {
		t.Fatalf("NamingChat: %v", err)
	}

	if len(result) != 5 {
		t.Fatalf("message count = %d, want 5 (system + user + assistant+toolcall + tool + assistant)", len(result))
	}

	foundTool := false
	for _, m := range result {
		if m.Role == llm.RoleTool && strings.Contains(m.Content, "丙") {
			foundTool = true
		}
	}
	if !foundTool {
		t.Error("tool result not found in messages")
	}
}

func TestNamingChat_MaxRounds(t *testing.T) {
	resps := make([]*llm.ChatResult, maxChatRounds)
	for i := range resps {
		resps[i] = ChatRes([]llm.ToolCall{ToolCall("compute_chart", `{}`)}, "")
	}
	mockLLM := &MockLLM{ToolResps: resps}
	mockTools := &MockToolRegistry{}

	a := NewChatAgent(mockLLM, mockTools, "prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "排盘"},
	}

	result, err := a.NamingChat(context.Background(), "zh-Hans", messages, nil)
	if err != nil {
		t.Fatalf("NamingChat: %v", err)
	}

	// system + user + 20*(assistant+tool) = 42
	want := 2 + maxChatRounds*2
	if len(result) != want {
		t.Errorf("message count = %d, want %d", len(result), want)
	}
}

func TestNamingChat_LLMError(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{nil},
		ToolErrs:  []error{fmt.Errorf("api unavailable")},
	}
	mockTools := &MockToolRegistry{}

	a := NewChatAgent(mockLLM, mockTools, "prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "你好"},
	}

	_, err := a.NamingChat(context.Background(), "zh-Hans", messages, nil)
	if err == nil {
		t.Error("expected error from LLM")
	}
}

func TestNamingChat_EventEmission(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(nil, "你好，请告诉我你的出生日期"),
		},
	}
	mockTools := &MockToolRegistry{}

	a := NewChatAgent(mockLLM, mockTools, "prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "你好"},
	}

	var events []ChatEvent
	_, err := a.NamingChat(context.Background(), "zh-Hans", messages, func(ev ChatEvent) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatalf("NamingChat: %v", err)
	}

	hasThinking := false
	hasTextDelta := false
	for _, ev := range events {
		switch ev.Type {
		case EventThinking:
			hasThinking = true
		case EventTextDelta:
			hasTextDelta = true
		}
	}
	if !hasThinking {
		t.Error("missing thinking event")
	}
	if !hasTextDelta {
		t.Error("missing text-delta event")
	}
}

func TestEnsureNamingPrompt_PrependsSystem(t *testing.T) {
	a := NewChatAgent(&MockLLM{}, &MockToolRegistry{}, "test {locale}")

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "你好"},
	}
	result := a.ensureNamingPrompt("zh-Hans", messages)

	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].Role != llm.RoleSystem {
		t.Errorf("result[0].Role = %q, want system", result[0].Role)
	}
	if !strings.Contains(result[0].Content, "zh-Hans") {
		t.Error("system prompt should have locale replaced")
	}
}

func TestIsNamingReport(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"standard_zh_Hans", "# 起名报告\n\n## 命理基础与用神", true},
		{"standard_zh_Hant", "# 起名報吿\n\n## 候選名字分析", true},
		{"preceded_by_text", "好的，以下是你的报告：\n\n# 起名报告\n\n## 命理基础", true},
		{"whitespace_before_heading", "  \n  # 起名报告\n\n内容", true},
		{"h2_not_h1", "## 起名报告", false},
		{"h3_not_h1", "### 起名报告\n\n内容", false},
		{"h1_later_in_content", "前面内容\n# 起名报告\n后面内容", true},
		{"no_report_keyword", "# 八字分析报告", false},
		{"empty", "", false},
		{"plain_text", "好的，已经为你生成了报告，请查看", false},
		{"only_hash", "#", false},
		{"hash_with_space_no_keyword", "# 其他内容", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNamingReport(tt.content)
			if got != tt.want {
				t.Errorf("IsNamingReport(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

func TestEnsureNamingPrompt_SkipsIfSystemPresent(t *testing.T) {
	a := NewChatAgent(&MockLLM{}, &MockToolRegistry{}, "ignored")
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "existing system prompt"},
		{Role: llm.RoleUser, Content: "你好"},
	}
	result := a.ensureNamingPrompt("zh-Hans", messages)

	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].Content != "existing system prompt" {
		t.Error("should not replace existing system prompt")
	}
}

func TestEnsureNamingPrompt_CachesByLocale(t *testing.T) {
	a := NewChatAgent(&MockLLM{}, &MockToolRegistry{}, "prompt-{locale}")

	r1 := a.ensureNamingPrompt("zh-Hans", []llm.Message{{Role: llm.RoleUser, Content: "hi"}})
	r2 := a.ensureNamingPrompt("zh-Hans", []llm.Message{{Role: llm.RoleUser, Content: "hi"}})

	if r1[0].Content != r2[0].Content {
		t.Error("cached prompts should be identical for same locale")
	}
	if r1[0].Content != "prompt-zh-Hans" {
		t.Errorf("prompt = %q, want prompt-zh-Hans", r1[0].Content)
	}
}

func TestEnsureNamingPrompt_DifferentLocales(t *testing.T) {
	a := NewChatAgent(&MockLLM{}, &MockToolRegistry{}, "prompt-{locale}")

	r1 := a.ensureNamingPrompt("zh-Hans", []llm.Message{{Role: llm.RoleUser, Content: "hi"}})
	r2 := a.ensureNamingPrompt("zh-Hant", []llm.Message{{Role: llm.RoleUser, Content: "hi"}})
	r3 := a.ensureNamingPrompt("en", []llm.Message{{Role: llm.RoleUser, Content: "hi"}})

	if r1[0].Content != "prompt-zh-Hans" {
		t.Errorf("zh-Hans = %q", r1[0].Content)
	}
	if r2[0].Content != "prompt-zh-Hant" {
		t.Errorf("zh-Hant = %q", r2[0].Content)
	}
	if r3[0].Content != "prompt-en" {
		t.Errorf("en = %q", r3[0].Content)
	}
}

func TestNamingChat_ToolErrorRecovery(t *testing.T) {
	// Tool execution error → error JSON fed to LLM as tool result → LLM can respond.
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes([]llm.ToolCall{ToolCall("compute_chart", `{"birth":{"time":"2026-06-25T14:00:00+08:00","longitude":116.4},"gender":"male"}`)}, ""),
			ChatRes(nil, "排盘暂时失败，请稍后重试"),
		},
	}
	mockTools := &MockToolRegistry{
		Errors: map[string]error{"compute_chart": fmt.Errorf("chart computation timeout")},
	}

	a := NewChatAgent(mockLLM, mockTools, "prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "帮我排八字"},
	}

	result, err := a.NamingChat(context.Background(), "zh-Hans", messages, nil)
	if err != nil {
		t.Fatalf("NamingChat: %v", err)
	}

	// system + user + assistant(toolcall) + tool(error) + assistant(response)
	if len(result) != 5 {
		t.Fatalf("message count = %d, want 5", len(result))
	}
	toolMsg := result[3]
	if toolMsg.Role != llm.RoleTool {
		t.Errorf("tool message role = %q, want tool", toolMsg.Role)
	}
	if !strings.Contains(toolMsg.Content, "error") {
		t.Errorf("tool error result should contain 'error': %s", toolMsg.Content)
	}
	// LLM received the error and responded with a message.
	lastMsg := result[4]
	if lastMsg.Role != llm.RoleAssistant {
		t.Errorf("last message role = %q, want assistant", lastMsg.Role)
	}
}

func TestNamingChat_ToolEmptyResult(t *testing.T) {
	// Tool returns empty JSON → still feeds to LLM as valid tool result.
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes([]llm.ToolCall{ToolCall("compute_chart", `{"birth":{"time":"2026-06-25T14:00:00+08:00","longitude":116.4},"gender":"male"}`)}, ""),
			ChatRes(nil, "收到排盘结果"),
		},
	}
	mockTools := &MockToolRegistry{
		Results: map[string]json.RawMessage{
			"compute_chart": json.RawMessage(`{}`),
		},
	}

	a := NewChatAgent(mockLLM, mockTools, "prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "帮我排八字"},
	}

	result, err := a.NamingChat(context.Background(), "zh-Hans", messages, nil)
	if err != nil {
		t.Fatalf("NamingChat: %v", err)
	}

	// system + user + assistant(toolcall) + tool(empty result) + assistant
	if len(result) != 5 {
		t.Fatalf("message count = %d, want 5", len(result))
	}
	toolMsg := result[3]
	if toolMsg.Role != llm.RoleTool {
		t.Errorf("tool message role = %q, want tool", toolMsg.Role)
	}
	if toolMsg.Content != "{}" {
		t.Errorf("tool result = %q, want {}", toolMsg.Content)
	}
}

func TestNamingChat_ToolCallThenErrorOnRetry(t *testing.T) {
	// First call: tool succeeds. Second call: LLM asks for another tool, which fails.
	// Agent should still feed the error and let LLM continue.
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes([]llm.ToolCall{ToolCall("compute_chart", `{}`)}, "好的，让我排盘"),
			ChatRes([]llm.ToolCall{ToolCall("compute_ziwei", `{}`)}, "再排紫微"),
			ChatRes(nil, "部分数据未能获取，但我会基于已有信息继续"),
		},
	}
	mockTools := &MockToolRegistry{
		Results: map[string]json.RawMessage{
			"compute_chart": json.RawMessage(`{"_product":"chart","data":{"nian":{"Gan":"丙","Zhi":"午"}}}`),
		},
		Errors: map[string]error{
			"compute_ziwei": fmt.Errorf("ziwei service unavailable"),
		},
	}

	a := NewChatAgent(mockLLM, mockTools, "prompt {locale}")
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "帮我排盘"},
	}

	result, err := a.NamingChat(context.Background(), "zh-Hans", messages, nil)
	if err != nil {
		t.Fatalf("NamingChat: %v", err)
	}

	// system + user + assistant(toolcall) + tool(success) + assistant(toolcall) + tool(error) + assistant
	if len(result) != 7 {
		t.Fatalf("message count = %d, want 7, got: %v", len(result), result)
	}
	if result[3].Role != llm.RoleTool {
		t.Errorf("msg[3] role = %q, want tool", result[3].Role)
	}
	if result[5].Role != llm.RoleTool {
		t.Errorf("msg[5] role = %q, want tool", result[5].Role)
	}
	if !strings.Contains(result[5].Content, "error") {
		t.Errorf("second tool (error): %s, want contains 'error'", result[5].Content)
	}
}

func TestEnsureNamingPrompt_ConcurrentCacheSafety(t *testing.T) {
	a := NewChatAgent(&MockLLM{}, &MockToolRegistry{}, "prompt-{locale}")
	locales := []string{"zh-Hans", "zh-Hant", "en", "ja", "ko"}

	done := make(chan struct{})
	for _, loc := range locales {
		go func(locale string) {
			for i := 0; i < 100; i++ {
				r := a.ensureNamingPrompt(locale, []llm.Message{{Role: llm.RoleUser, Content: "hi"}})
				if r[0].Content != "prompt-"+locale {
					panic("bad prompt: " + r[0].Content)
				}
			}
			done <- struct{}{}
		}(loc)
	}
	for range locales {
		<-done
	}
}
