package handler

import (
	"log"
	"net/http"

	doc "liki"
)

func handleOpenAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if _, err := w.Write(doc.OpenAPIJSON); err != nil {
			log.Println("handleOpenAPI: write error:", err)
		}
	}
}
