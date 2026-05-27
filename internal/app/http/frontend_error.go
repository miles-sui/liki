package http

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// collectFrontendError accepts JS error reports via sendBeacon and stores them in SQLite.
// Silently drops when daily cap (1000) is reached.
func collectFrontendError(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type fe struct {
			Message  string `json:"message"`
			Filename string `json:"filename"`
			Lineno   int    `json:"lineno"`
			Colno    int    `json:"colno"`
			Stack    string `json:"stack"`
			URL      string `json:"url"`
		}
		var e fe
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			respondJSON(w, http.StatusOK, map[string]string{"status": "dropped"})
			return
		}

		if db == nil {
			respondJSON(w, http.StatusOK, map[string]string{"status": "logged"})
			log.Printf("[frontend-err] %s %s:%d:%d — %s", e.URL, e.Filename, e.Lineno, e.Colno, e.Message)
			return
		}

		var todayCount int
		db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM frontend_errors WHERE created_at >= date('now')`).Scan(&todayCount)
		if todayCount >= 1000 {
			respondJSON(w, http.StatusOK, map[string]string{"status": "dropped"})
			return
		}

		db.ExecContext(r.Context(),
			`INSERT INTO frontend_errors (message, filename, lineno, colno, stack, url)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			strings.TrimSpace(e.Message), strings.TrimSpace(e.Filename),
			e.Lineno, e.Colno, strings.TrimSpace(e.Stack), strings.TrimSpace(e.URL))

		if todayCount%100 == 0 {
			db.ExecContext(r.Context(),
				`DELETE FROM frontend_errors WHERE created_at < date('now', '-30 days')`)
		}

		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
