package llm

import "testing"

func TestNew(t *testing.T) {
	c := New("sk-test")
	if c == nil {
		t.Fatal("New returned nil")
		return
	}
	if c.apiKey != "sk-test" {
		t.Errorf("apiKey = %q, want sk-test", c.apiKey)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.model != defaultModel {
		t.Errorf("model = %q, want %q", c.model, defaultModel)
	}
}
