package minglihttp

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed docs/*
var docsRoot embed.FS

// docsFS is the docs/ subdirectory as a filesystem root.
var docsFS fs.FS

func init() {
	var err error
	docsFS, err = fs.Sub(docsRoot, "docs")
	if err != nil {
		panic(err)
	}
}

// RegisterDocs registers AI-native documentation routes on the given ServeMux.
func RegisterDocs(mux *http.ServeMux) {
	docsHandler := http.StripPrefix("/docs", http.FileServer(http.FS(docsFS)))

	mux.HandleFunc("GET /docs/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		docsHandler.ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /llms.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFileFS(w, r, docsFS, "llms.txt")
	})
}
