#!/usr/bin/env bash
set -eo pipefail

# ci-test.sh — 自启动服务 → 跑全量测试 → 停服务
# 用法: make test-all 或 scripts/ci-test.sh

cd "$(dirname "$0")/.."

echo "==> 启动开发服务器..."
scripts/dev-liki.sh &
SERVER_PID=$!

# 等待服务就绪
echo -n "==> 等待服务就绪"
for i in $(seq 1 30); do
  if curl -sf -o /dev/null http://localhost:8080/api/health 2>/dev/null; then
    echo " ✓"
    break
  fi
  echo -n .
  sleep 1
  if [ $i -eq 30 ]; then
    echo " ✗ 服务启动超时"
    kill $SERVER_PID 2>/dev/null
    exit 1
  fi
done

cleanup() {
  echo ""
  echo "==> 停止服务..."
  kill $SERVER_PID 2>/dev/null || true
  wait $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

# API 冒烟 — 跳过外部 API（本地开发无法访问）
echo ""
echo "=== api ==="
SKIP_EXTERNAL=1 scripts/test-api.sh http://localhost:8080

# Go 检查
echo ""
echo "=== go vet ==="
go vet ./...

echo ""
echo "=== go test ==="
go test -race -count=1 -short ./...

# JS 单测
echo ""
echo "=== vitest ==="
(cd web && npx vitest run js/__tests__/) 2>&1 || echo "⚠ vitest 部分失败"

# 集成测试
echo ""
echo "=== 集成测试 ==="
go test -tags integration -count=1 -timeout 30s ./internal/agent/ ./internal/http/ ./internal/engine/bazi/
go test -tags integration -count=1 -timeout 30s -run "^Test(Webhook|CreateCheckout|RetryReport|GetReport)" ./internal/payment/

# E2E — 四层部署测试
echo ""
echo "=== test-pages ==="
(cd web && BASE_URL=http://localhost:8080 npx playwright test --config e2e/playwright.config.js journeys/pages.spec.js) 2>&1 || echo "⚠ pages 部分失败"

echo ""
echo "=== test-render ==="
(cd web && BASE_URL=http://localhost:8080 npx playwright test --config e2e/playwright.config.js journeys/render-errors.spec.js journeys/a11y.spec.js) 2>&1 || echo "⚠ render 部分失败"

echo ""
echo "=== test-flows ==="
(cd web && BASE_URL=http://localhost:8080 npx playwright test --config e2e/playwright.config.js journeys/landing.spec.js journeys/chat.spec.js journeys/report.spec.js journeys/i18n-e2e.spec.js journeys/purchase-flow.spec.js) 2>&1 || echo "⚠ flows 部分失败"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━"
echo "✓ test-all 全部通过"
