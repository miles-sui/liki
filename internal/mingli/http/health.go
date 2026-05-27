package minglihttp

import (
	"net/http"

	"github.com/25types/25types/internal/httputil"
)

// Health returns a simple stateless health check.
func Health(w http.ResponseWriter, r *http.Request) {
	httputil.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
