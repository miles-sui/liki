package http

import (
	"net/http"
)

func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// handleVersion returns the build time of the running binary.
func handleVersion(buildTime string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"build_time": buildTime})
	}
}
