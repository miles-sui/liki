package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	minglihttp "github.com/25types/25types/internal/mingli/http"
	minglimcp "github.com/25types/25types/internal/mingli/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		*addr = v
	}

	mux := http.NewServeMux()
	minglihttp.RegisterRoutes(mux)

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "mingli",
		Title:   "Mingli — Chinese Metaphysics Server",
		Version: "1.0.0",
	}, nil)
	minglimcp.RegisterTools(mcpServer)
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return mcpServer }, &mcp.StreamableHTTPOptions{
		Stateless: true,
	}))

	srv := &http.Server{
		Addr:         *addr,
		Handler:      minglihttp.WrapWithManifest(minglihttp.CORSMiddleware(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Printf("received %s, shutting down gracefully...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server forced to shutdown: %v", err)
		}
	}()

	log.Printf("mingli-server (mingli API) listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
	log.Println("server stopped")
}
