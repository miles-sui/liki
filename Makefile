.PHONY: build vet clean check dev smoke deploy us cn test-setup test-js test-frontend-watch test-e2e test-smoke test-render test-frontend test-integration test-all

build:
	BUILD_TIME=$$(date -u '+%Y-%m-%dT%H:%M:%SZ'); \
	node web/scripts/compile-vue-template.cjs; \
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.BuildTime=$$BUILD_TIME" -o bin/lingji ./cmd/lingji/

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

check: vet test

hooks:
	git config core.hooksPath .githooks

clean:
	rm -rf bin/

# ---- LingJi targets ----

dev:
	scripts/dev-lingji.sh

smoke:
	scripts/smoke-lingji.sh $(URL)

deploy:
	scripts/deploy-lingji.sh $(filter-out deploy,$(MAKECMDGOALS))

us cn:
	@:

# ---- Frontend test targets ----

test-setup:
	cd web && npm install

test-js:
	cd web && npx vitest run js/__tests__/

test-frontend-watch:
	cd web && npx vitest js/__tests__/

test-e2e:
	cd web && BASE_URL=$(URL) npx playwright test --config e2e/playwright.config.js

test-smoke:
	cd web && BASE_URL=$(URL) npx playwright test --config e2e/playwright.config.js journeys/smoke.spec.js

test-render:
	cd web && BASE_URL=$(URL) npx playwright test --config e2e/playwright.config.js journeys/render-errors.spec.js

test-frontend:
	scripts/test-frontend.sh $(URL)

test-integration:
	go test -tags integration -count=1 -timeout 30s ./internal/agent/ ./internal/http/ ./internal/engine/bazi/
	go test -tags integration -count=1 -timeout 30s -run "^Test(Webhook|CreateCheckout|RetryReport|GetReport)" ./internal/payment/

test-all: build
	scripts/ci-test.sh
