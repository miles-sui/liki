package handler

import "net/http"

var prodOrigins = map[string]bool{
	"https://liki.hk": true,
}

var devOrigins = map[string]bool{
	"http://localhost:8080": true,
	"http://localhost:8081": true,
}

var devMode bool

// SetDevMode enables localhost CORS origins for development.
func SetDevMode(v bool) { devMode = v }

// CORSMiddleware adds permissive CORS headers for allowed origins and handles OPTIONS preflight.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := prodOrigins[origin]
		if !allowed && devMode {
			allowed = devOrigins[origin]
		}
		if origin == "" || !allowed {
			origin = "https://liki.hk"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds X-Content-Type-Options: nosniff and X-Frame-Options: DENY.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CSP is managed by Caddy for static assets; API responses are JSON only.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// BodyLimit limits request body size to 1 MB.
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
		next.ServeHTTP(w, r)
	})
}
