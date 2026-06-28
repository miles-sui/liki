package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"liki/internal/agent"
)

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event agent.ChatEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

// flushSSE sends an SSE comment to ensure framing is complete before handler returns.
// Prevents data loss when the handler returns immediately after streaming.
func flushSSE(w http.ResponseWriter, flusher http.Flusher) {
	_, err := fmt.Fprint(w, ": ok\n\n")
	if err != nil {
		slog.Warn("naming: flush sse", "err", err)
		return
	}
	flusher.Flush()
}
