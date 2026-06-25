.PHONY: build vet clean check dev test-api test-pages test-render test-flows test-deploy deploy us cn test-setup test-js test-frontend-watch test-integration test-all
# Legacy aliases
smoke: test-api
test-smoke: test-pages
test-e2e: test-flows
check-deploy: test-deploy

build:
	BUILD_TIME=$$(date -u '+%Y-%m-%dT%H:%M:%SZ'); \
	node web/scripts/compile-vue-template.cjs; \
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.BuildTime=$$BUILD_TIME" -o bin/lingji ./cmd/lingji/

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

lint:
	golangci-lint run --timeout=3m

check: lint vet test

hooks:
	git config core.hooksPath .githooks

clean:
	rm -rf bin/

# ---- LingJi targets ----

dev:
	scripts/dev-lingji.sh

# ---- 部署后测试：API → 页面 → 流程 ----

test-api:
	scripts/test-api.sh $(URL)

test-pages:
	cd web && BASE_URL=$(URL) npx playwright test --config e2e/playwright.config.js journeys/pages.spec.js

test-render:
	cd web && BASE_URL=$(URL) npx playwright test --config e2e/playwright.config.js journeys/render-errors.spec.js

FLOW_SPECS = journeys/landing.spec.js journeys/chart.spec.js journeys/naming.spec.js \
             journeys/chat.spec.js journeys/report.spec.js journeys/i18n-e2e.spec.js \
             journeys/purchase-flow.spec.js
test-flows:
	cd web && BASE_URL=$(URL) npx playwright test --config e2e/playwright.config.js $(FLOW_SPECS)

# 四层按序全跑，任一步失败即停
test-deploy:
	@echo "=== 1/4 API 层 ==="
	@$(MAKE) test-api URL=$(URL) || { echo "❌ API 层失败"; exit 1; }
	@echo ""
	@echo "=== 2/4 页面层 ==="
	@$(MAKE) test-pages URL=$(URL) || { echo "❌ 页面层失败"; exit 1; }
	@echo ""
	@echo "=== 3/4 渲染层 ==="
	@$(MAKE) test-render URL=$(URL) || { echo "❌ 渲染层失败"; exit 1; }
	@echo ""
	@echo "=== 4/4 流程层 ==="
	@$(MAKE) test-flows URL=$(URL) || { echo "❌ 流程层失败"; exit 1; }
	@echo ""
	@echo "✅ 部署验证全部通过。"

deploy:
	scripts/deploy-lingji.sh $${TARGET:-all}

# Usage: TARGET=us make deploy

# ---- Frontend test targets ----

test-setup:
	cd web && npm install

test-js:
	cd web && npx vitest run js/__tests__/

test-frontend-watch:
	cd web && npx vitest js/__tests__/

test-integration:
	go test -tags integration -count=1 -timeout 30s ./internal/agent/ ./internal/http/ ./internal/engine/bazi/
	go test -tags integration -count=1 -timeout 30s -run "^Test(Webhook|CreateCheckout|RetryReport|GetReport)" ./internal/payment/

test-all: build
	scripts/ci-test.sh
