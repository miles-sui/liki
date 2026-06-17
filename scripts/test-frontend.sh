#!/usr/bin/env bash
set -eo pipefail

# test-frontend.sh — 前端三层测试: smoke → render-errors → e2e
# 用法: scripts/test-frontend.sh [URL]
# 默认: http://localhost:8080

URL="${1:-http://localhost:8080}"
CFG="e2e/playwright.config.js"
DIR="e2e/journeys"
PASS=0
FAIL=0

red()  { echo -e "\033[31m$1\033[0m"; }
green(){ echo -e "\033[32m$1\033[0m"; }

run_layer() {
  local layer="$1"; shift
  local specs="$*"
  echo ""
  echo "=== $layer ==="
  if BASE_URL="$URL" npx playwright test --config="$CFG" $specs 2>&1; then
    green "✓ $layer passed"
    PASS=$((PASS + 1))
  else
    red "✗ $layer failed"
    FAIL=$((FAIL + 1))
    # 冒烟失败则终止
    if [ "$layer" = "smoke" ]; then
      red "smoke 失败，终止后续测试"
      exit 1
    fi
  fi
}

cd "$(dirname "$0")/../web"

# Layer 1: Smoke — 快速冒烟，失败则终止
run_layer "smoke" "$DIR/smoke.spec.js"

# Layer 2: Render errors — 框架渲染问题
run_layer "render-errors" "$DIR/render-errors.spec.js"

# Layer 3: Full E2E — 页面交互 + 业务流程
run_layer "e2e" \
  "$DIR/landing.spec.js" \
  "$DIR/chart.spec.js" \
  "$DIR/naming.spec.js" \
  "$DIR/chat.spec.js" \
  "$DIR/report.spec.js" \
  "$DIR/i18n-e2e.spec.js" \
  "$DIR/purchase-flow.spec.js"

echo ""
echo "---"
echo "结果: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] && green "全部通过" || red "部分失败"
exit "$FAIL"
