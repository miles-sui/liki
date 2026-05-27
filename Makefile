.PHONY: build test vet dev dev-local clean test-unit test-integration test-frontend-src test-e2e check test-all test-all-e2e

# Build API server only
build:
	go build -ldflags="-s -w" -o bin/app-server ./cmd/app-server/

# Run all Go tests (fast, < 5s)
test: test-unit test-integration

# Unit tests — pure logic + questionnaire loader + DB migrations
test-unit:
	go test ./internal/ganzhi/ ./internal/tianwen/ ./internal/25types/ ./internal/httputil/ ./internal/mingli/bazi/ ./internal/mingli/http/

# Integration tests — httptest + real SQLite + real routes
test-integration:
	go test ./internal/app/sqlite/ ./internal/app/http/ ./internal/app/application/...

# E2E tests — Playwright in real browser (requires server running)
test-e2e:
	cd web && npm run test:e2e

# Frontend source contract — static analysis, no server needed
test-frontend-src:
	scripts/test-frontend-src.sh

# Go vet
vet:
	go vet ./...

# CI target — no server needed, fast feedback
check: vet test test-frontend-src

# Full test suite — requires server running (test-all.sh does NOT start it)
test-all:
	scripts/test-all.sh

# Full test suite + Playwright E2E
test-all-e2e:
	scripts/test-all.sh --e2e

# Start development environment (Caddy + API via Docker Compose)
dev:
	scripts/dev.sh

# Start dev environment — bare Go + Caddy (no Docker)
dev-local:
	scripts/dev-local.sh

# Clean build artifacts
clean:
	rm -rf bin/
