package minglihttp

import (
	"io/fs"
	"net/http"
)

// WrapWithManifest intercepts GET / to serve the AI agent manifest (llms.txt) as markdown,
// delegating all other requests to next.
func WrapWithManifest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			content, err := fs.ReadFile(docsFS, "llms.txt")
			if err == nil {
				w.Write(content)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
