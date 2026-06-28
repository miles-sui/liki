package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"liki/internal/agent"
)

// API contract: all endpoints return envelope {"data":{...}} or {"error":{...}}

func TestEnvelope_JSONRPC_Discover(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"jsonrpc":"2.0","method":"rpc.discover","id":1}`))
	w := httptest.NewRecorder()
	handleRPC(agent.NewRPCRegistry())(w, r)

	var env struct {
		JSONRPC string          `json:"jsonrpc"`
		Result  json.RawMessage `json:"result"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
		return
	}
	if w.Code == http.StatusOK {
		if env.Result == nil {
			t.Error("rpc.discover response should have result")
		}
	}
}

func TestEnvelope_JSONRPC_Error(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"jsonrpc":"2.0","method":"nonexistent","id":1}`))
	w := httptest.NewRecorder()
	handleRPC(agent.NewRPCRegistry())(w, r)

	var env struct {
		JSONRPC string          `json:"jsonrpc"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
		return
	}
	if env.Error == nil {
		t.Error("error response should have error field")
	}
}

func TestEnvelope_ErrorFormat_HasCodeAndMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"jsonrpc invalid version", `{"jsonrpc":"1.0","method":"x","id":1}`},
		{"jsonrpc missing method", `{"jsonrpc":"2.0","id":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			handleRPC(agent.NewRPCRegistry())(w, r)

			var env struct {
				Error struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Errorf("decode error response: %v", err)
				return
			}
			if env.Error.Code == 0 {
				t.Error("error response missing code")
			}
			if env.Error.Message == "" {
				t.Error("error response missing message")
			}
		})
	}
}

func TestEdge_BodyNotJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("not json at all"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleRPC(agent.NewRPCRegistry())(w, r)
	if w.Code >= 500 {
		t.Errorf("non-JSON body caused 5xx: %d", w.Code)
	}
}

func TestEdge_BodyWithBOM(t *testing.T) {
	body := "\xef\xbb\xbf" + `{"jsonrpc":"2.0","method":"rpc.discover","id":1}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleRPC(agent.NewRPCRegistry())(w, r)
	if w.Code >= 500 {
		t.Errorf("UTF-8 BOM caused 5xx: %d", w.Code)
	}
}
