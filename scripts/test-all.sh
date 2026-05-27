#!/usr/bin/env bash
# 25types — Full local test suite
# Runs all tests that don't require a browser against a running local server.
#
# Usage:
#   scripts/test-all.sh                     # → localhost:8080
#   scripts/test-all.sh --e2e               # + Playwright E2E
#
# Service lifecycle is YOUR responsibility:
#   scripts/dev-local.sh          # bare Go + Caddy
#   docker compose up -d          # local Docker
set -euo pipefail

cd "$(dirname "$0")/.."

BASE="http://localhost:8080"
E2E=false
for arg in "$@"; do
  case "$arg" in
    --e2e) E2E=true ;;
  esac
done

export JWT_SECRET="${JWT_SECRET:-dev-secret-change-me}"
if [ -f .env ]; then
  set -a; source .env; set +a
fi

PASS=0; FAIL=0
RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

say() { printf "${BOLD}%s${NC}\n" "$*"; }
ok()  { printf "  ${GREEN}PASS${NC} %s\n" "$*"; PASS=$((PASS+1)); }
bad() { printf "  ${RED}FAIL${NC} %s\n" "$*"; FAIL=$((FAIL+1)); }

run_phase() {
  local label="$1"; shift
  say "=== $label ==="
  if "$@"; then ok "$label"; else bad "$label"; fi
}

# ── Phase 1: No server needed ────────────────────────────────────

say "Phase 1: Go + Frontend source"

run_phase "go vet"            go vet ./...
run_phase "go test unit"      go test ./internal/ganzhi/ ./internal/tianwen/ ./internal/25types/ ./internal/httputil/ ./internal/mingli/bazi/ ./internal/mingli/http/
run_phase "go test integ"     go test ./internal/app/sqlite/ ./internal/app/http/ ./internal/app/application/...
run_phase "frontend src"      bash scripts/test-frontend-src.sh

# ── Phase 2: HTTP-level tests (server required) ──────────────────

say "Phase 2: HTTP tests → $BASE"

run_phase "smoke (journey)" bash scripts/smoke.sh "$BASE"

# ── Phase 3: E2E (optional) ──────────────────────────────────────

if "$E2E"; then
  if ! npx playwright --version >/dev/null 2>&1; then
    echo "  SKIP Playwright not installed. Run: cd web && npm ci"
  else
    run_phase "playwright e2e" bash -c "cd web && npm run test:e2e"
  fi
fi

# ── Report ───────────────────────────────────────────────────────

printf "\n${BOLD}── Result ──${NC}\n"
printf "  ${GREEN}%d passed${NC}, ${RED}%d failed${NC}\n" "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
