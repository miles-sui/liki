#!/bin/bash
# Liki — API smoke test
set -uo pipefail

BASE="${1:-http://localhost:8080}"
if [ $# -ge 2 ]; then
  API="$2"
elif echo "$BASE" | /bin/grep -q '^http://localhost'; then
  API="http://localhost:8081"
else
  API="$BASE"
fi
PASS=0; FAIL=0
HAS_JQ=false

RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

TMP=$(mktemp -d /tmp/smoke-lingji-XXXXXX)
trap 'rm -rf "$TMP"' EXIT

command -v jq &>/dev/null && HAS_JQ=true

api() {
  local method="$1" path="$2" data="${3:-}"
  local f="$TMP/body"
  local fake_ip="10.$((RANDOM % 256)).$((RANDOM % 256)).$((RANDOM % 254 + 1))"
  if [ -n "$data" ]; then
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$API$path" \
      -H 'Content-Type: application/json' \
      -H "X-Forwarded-For: $fake_ip" \
      -d "$data" || echo "000"
  else
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$API$path" \
      -H 'Content-Type: application/json' \
      -H "X-Forwarded-For: $fake_ip" || echo "000"
  fi
  sleep 0.1
}

caddy() {
  local method="$1" path="$2" data="${3:-}"
  local f="$TMP/body"
  if [ -n "$data" ]; then
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$BASE$path" \
      -H 'Content-Type: application/json' \
      -d "$data" || echo "000"
  else
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$BASE$path" \
      -H 'Content-Type: application/json' || echo "000"
  fi
}

body() { cat "$TMP/body" 2>/dev/null || true; }

json_val() {
  echo "$1" | jq -r "$2" 2>/dev/null || true
}

check() {
  local desc="$1" expected="$2" actual="${3:-}"
  if [ "$actual" = "$expected" ]; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}\xe2\x9c\x97${NC} $desc (expected '$expected', got '$actual')"
    FAIL=$((FAIL + 1))
  fi
}

check_200()  { check "$1 HTTP" "200" "${2:-}"; }
check_204()  { check "$1 HTTP" "204" "${2:-}"; }
check_302()  { check "$1 HTTP" "302" "${2:-}"; }
check_400()  { check "$1 HTTP" "400" "${2:-}"; }
check_404()  { check "$1 HTTP" "404" "${2:-}"; }
check_422()  { check "$1 HTTP" "422" "${2:-}"; }

check_403()  { check "$1 HTTP" "403" "${2:-}"; }

echo "${BOLD}Liki — API Smoke Test${NC}"
echo "Target: Caddy=$BASE API=$API"
$HAS_JQ && echo "jq: available" || echo "jq: not available (limited validation)"
echo ""

# --- Shared test data ---
BT='{"time":"1984-02-04T18:30:00+08:00","longitude":116.4}'
BT_A='{"time":"1990-03-20T10:30:00+08:00","longitude":120}'
BT_B='{"time":"1992-07-08T14:30:00+08:00","longitude":120}'
BR="{\"birth\":$BT,\"gender\":\"male\"}"
BR_A="{\"birth\":$BT_A,\"gender\":\"male\"}"
BR_B="{\"birth\":$BT_B,\"gender\":\"female\"}"

# ============================================================================
# Health check
# ============================================================================
echo "${BOLD}── Health ──${NC}"
s=$(api GET /api/health)
check_200 "GET /api/health" "$s"
b=$(body)
check "  status ok" "ok" "$(json_val "$b" '.data.status')"

# ============================================================================
# Static (Caddy)
# ============================================================================
echo ""
echo "${BOLD}── Static ──${NC}"
# Only test if Caddy is running on $BASE
if curl -s --connect-timeout 2 -o /dev/null -w '%{http_code}' "$BASE/api/health" 2>/dev/null | /bin/grep -q .; then
  s=$(caddy GET /)
  check_200 "GET /" "$s"
  b=$(body)
  check "  has 灵机" "false" "$(echo "$b" | /bin/grep -q '灵机' && echo false || echo true)"
else
  echo "  (Caddy not running — skipping static checks)"
fi

# ============================================================================
# Wiki
# ============================================================================
echo ""
echo "${BOLD}── Wiki ──${NC}"
s=$(caddy GET /wiki/)
check_200 "GET /wiki/" "$s"
b=$(body)
check "  redirects to zh-hant" "false" "$(echo "$b" | /bin/grep -q 'zh-hant/index.html' && echo false || echo true)"
s=$(caddy GET /wiki/zh-hant/index.html)
check_200 "GET /wiki/zh-hant/index.html" "$s"
s=$(caddy GET /wiki/entity/)
check_200 "GET /wiki/entity/" "$s"
b=$(body)
check "  entity redirects to ../zh-hant" "false" "$(echo "$b" | /bin/grep -q '../zh-hant/entity/index.html' && echo false || echo true)"
s=$(caddy GET /wiki/zh-hant/entity/index.html)
check_200 "GET /wiki/zh-hant/entity/index.html" "$s"

# ============================================================================
# Payment
# ============================================================================
echo ""
echo "${BOLD}── Payment ──${NC}"
s=$(api POST /api/payments/checkout '{"order_id":"nonexistent","email":"test@example.com"}')
check_404 "POST /api/payments/checkout (no order)" "$s"

s=$(api GET /api/orders/nonexistent/report)
check_302 "GET /api/orders/{id}/report" "$s"

s=$(api POST /api/payments/webhook '{"type":"order.paid","data":{"order_id":"test"}}')
check_400 "POST /api/payments/webhook (no sig)" "$s"

s=$(api GET /api/orders/nonexistent/status)
check_404 "GET /api/orders/{id}/status (not found)" "$s"

# ============================================================================
# Misc
# ============================================================================
echo ""
echo "${BOLD}── Misc ──${NC}"
s=$(api GET /api/version)
check_200 "GET /api/version" "$s"

s=$(api GET /api/location)
check_200 "GET /api/location" "$s"

s=$(api GET /api/stats)
check_200 "GET /api/stats" "$s"

s=$(api POST /api/analytics/pageview '{"path":"/zh/"}')
check_204 "POST /api/analytics/pageview" "$s"

# ============================================================================
# Agent
# ============================================================================
echo ""
echo "${BOLD}── Agent ──${NC}"
s=$(api GET /api/agent/greeting)
check_200 "GET /api/agent/greeting" "$s"

# Chat SSE — verify headers + no error in stream
chat_code=$(curl -s -w '%{http_code}' -o "$TMP/chat_body" "$API/api/agent/chat" \
  -H 'Content-Type: application/json' -d '{"message":"hello"}' \
  -D "$TMP/chat_headers")
check_200 "POST /api/agent/chat" "$chat_code"
chat_ct=$(/bin/grep -i 'content-type:' "$TMP/chat_headers" 2>/dev/null | tr -d '\r' || true)
check "  SSE content-type" "false" "$(echo "$chat_ct" | /bin/grep -q 'text/event-stream' && echo false || echo true)"
chat_sid=$(/bin/grep -i 'x-session-id:' "$TMP/chat_headers" 2>/dev/null | tr -d '\r' || true)
check "  has X-Session-ID" "false" "$([ -n "$chat_sid" ] && echo false || echo true)"
# Verify no error event in SSE stream
if echo "$chat_code" | /bin/grep -q '200' && /bin/grep -q '"type":"error"' "$TMP/chat_body" 2>/dev/null; then
  chat_err=$(/bin/grep -o '"content":"[^"]*"' "$TMP/chat_body" 2>/dev/null | head -1 || true)
  check "  no SSE error" "true" "false — SSE stream contains error: $chat_err"
else
  check "  no SSE error" "false" "false"
fi

s=$(api POST /api/agent/chat '{"message":""}')
check_400 "POST /api/agent/chat (empty message)" "$s"

# ============================================================================
# Huangli
# ============================================================================
echo ""
echo "${BOLD}── Huangli ──${NC}"

s=$(api GET '/api/huangli/date?date=2026-06-19&event=%E5%AB%81%E5%A8%B6')
check_200 "GET /api/huangli/date" "$s"

HL_BD="{\"birth\":$BT,\"event_type\":\"嫁娶\",\"date\":\"2026-06-19\"}"
s=$(api POST /api/huangli/bond/date "$HL_BD")
check_200 "POST /api/huangli/bond/date" "$s"

# Negative: missing params
s=$(api GET /api/huangli/date)
check_400 "GET /api/huangli/date (missing params)" "$s"

# Empty birth → validation_error (time is required)
s=$(api POST /api/huangli/bond/date '{"birth":{},"event_type":"嫁娶","date":"2026-06-19"}')
check_422 "POST /api/huangli/bond/date (empty birth)" "$s"

# ============================================================================
# Bazhai
# ============================================================================
echo ""
echo "${BOLD}── Bazhai ──${NC}"

s=$(api POST /api/bazhai/minggua '{"gender":"male","birth_year":1984}')
check_200 "POST /api/bazhai/minggua" "$s"

s=$(api POST /api/bazhai/chart "$BR")
check_200 "POST /api/bazhai/chart" "$s"

# Negative
s=$(api POST /api/bazhai/minggua '{"gender":"other","birth_year":1984}')
check_422 "POST /api/bazhai/minggua (bad gender)" "$s"

s=$(api POST /api/bazhai/minggua '{"gender":"male"}')
check_422 "POST /api/bazhai/minggua (missing year)" "$s"

# ============================================================================
# Xuankong
# ============================================================================
echo ""
echo "${BOLD}── Xuankong ──${NC}"

s=$(api GET '/api/xuankong/sanyuan?year=2026')
check_200 "GET /api/xuankong/sanyuan" "$s"

XK="{\"birth\":$BT,\"sit_mountain\":0,\"face_mountain\":11}"
s=$(api POST /api/xuankong/chart "$XK")
check_200 "POST /api/xuankong/chart" "$s"

s=$(api POST /api/xuankong/chart "{\"birth\":$BT}")
check_422 "POST /api/xuankong/chart (missing mountains)" "$s"

# ============================================================================
# BaZi
# ============================================================================
echo ""
echo "${BOLD}── BaZi ──${NC}"

s=$(api POST /api/bazi/chart "$BR")
check_200 "POST /api/bazi/chart" "$s"

# BaZi bond — returns Bond struct directly
BOND="{\"a\":$BR_A,\"b\":$BR_B}"
s=$(api POST /api/bazi/bond "$BOND")
check_200 "POST /api/bazi/bond" "$s"

# BaZi luck cycles (gender required for 顺排/逆排)
s=$(api POST /api/bazi/liunian "{\"year\":2026,\"gender\":\"male\",\"birth\":$BT}")
check_200 "POST /api/bazi/liunian" "$s"

# Negative
s=$(api POST /api/bazi/chart 'not-json')
check_400 "POST /api/bazi/chart (bad json)" "$s"

s=$(api POST /api/bazi/chart "{\"birth\":$BT}")
check_422 "POST /api/bazi/chart (missing gender)" "$s"

# ============================================================================
# ZiWei
# ============================================================================
echo ""
echo "${BOLD}── ZiWei ──${NC}"

s=$(api POST /api/ziwei/chart "$BR")
check_200 "POST /api/ziwei/chart" "$s"
b=$(body)

# Dependent endpoints need the full chart
if $HAS_JQ; then
  ZW_CHART=$(echo "$b" | jq -c '.data' 2>/dev/null)

  s=$(api POST /api/ziwei/daxian "{\"chart\":$ZW_CHART,\"gender\":\"male\"}")
  check_200 "POST /api/ziwei/daxian" "$s"
fi

s=$(api POST /api/ziwei/chart "{\"birth\":$BT}")
check_422 "POST /api/ziwei/chart (missing gender)" "$s"

# ============================================================================
# QiMen
# ============================================================================
echo ""
echo "${BOLD}── QiMen ──${NC}"

s=$(api POST /api/qimen/pan "{\"birth\":$BT,\"kind\":\"shi\"}")
check_200 "POST /api/qimen/pan (shi)" "$s"

s=$(api POST /api/qimen/pan "{\"birth\":$BT,\"kind\":\"invalid\"}")
check_422 "POST /api/qimen/pan (bad kind)" "$s"

# ============================================================================
# LiuYao
# ============================================================================
echo ""
echo "${BOLD}── LiuYao ──${NC}"

LY="{\"birth\":$BT}"
s=$(api POST /api/liuyao/chart "$LY")
check_200 "POST /api/liuyao/chart" "$s"

LYF="{\"birth\":$BT,\"yong_shen\":\"父母\",\"fixed\":[6,7,8,9,6,7]}"
s=$(api POST /api/liuyao/chart "$LYF")
check_200 "POST /api/liuyao/chart (fixed)" "$s"

# Negative: bad fixed values (1-5 are invalid)
s=$(api POST /api/liuyao/chart "{\"birth\":$BT,\"fixed\":[1,2,3,4,5,6]}")
check_422 "POST /api/liuyao/chart (bad fixed)" "$s"

# ============================================================================
# Qiming
# ============================================================================
echo ""
echo "${BOLD}── Qiming ──${NC}"

s=$(api POST /api/qiming/wuge '{"surname":"李","yong_shen":"水","xi_shen":["金"]}')
check_200 "POST /api/qiming/wuge" "$s"

s=$(api POST /api/qiming/evaluate '{"surname":"李","given_name":"沐泽","yong_shen":"水"}')
check_200 "POST /api/qiming/evaluate" "$s"

# Negative
s=$(api POST /api/qiming/wuge '{"surname":"李"}')
check_422 "POST /api/qiming/wuge (missing yong_shen)" "$s"

s=$(api POST /api/qiming/evaluate '{"surname":"李"}')
check_422 "POST /api/qiming/evaluate (missing params)" "$s"

# ============================================================================
# Access Policy
# ============================================================================
if curl -s --connect-timeout 2 -o /dev/null -w '%{http_code}' "$BASE/api/health" 2>/dev/null | /bin/grep -q .; then
  echo ""
  echo "${BOLD}── Access Policy ──${NC}"

  # Origin blocking is a Caddy-level feature (production Caddyfile only).
  # Skip when running against local Caddy (no @external rules in Caddyfile.local).
  if [ "$BASE" = "http://localhost:8080" ] || [ "$BASE" = "http://127.0.0.1:8080" ]; then
    echo "  (local Caddy — skipping access policy checks)"
  else
    s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$BASE/api/agent/chat" \
      -H 'Content-Type: application/json' \
      -H 'Origin: https://evil.com' \
      -d '{"message":"hello"}')
    check_403 "POST /api/agent/chat (external Origin)" "$s"

    s=$(curl -s -w '%{http_code}' -o /dev/null "$BASE/api/agent/greeting" \
      -H 'Origin: https://evil.com')
    check_403 "GET /api/agent/greeting (external Origin)" "$s"

    s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$BASE/api/orders/nonexistent/retry" \
      -H 'Origin: https://evil.com')
    check_403 "POST /api/orders/{id}/retry (external Origin)" "$s"

    # Public APIs with external Origin → normal response
    s=$(curl -s -w '%{http_code}' -o /dev/null "$BASE/api/health" \
      -H 'Origin: https://evil.com')
    check_200 "GET /api/health (external Origin ignored)" "$s"

    s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$BASE/api/bazi/chart" \
      -H 'Content-Type: application/json' \
      -H 'Origin: https://evil.com' \
      -d "{\"birth\":$BT,\"gender\":\"male\"}")
    check_200 "POST /api/bazi/chart (external Origin ignored)" "$s"
  fi
else
  echo "  (Caddy not running — skipping access policy checks)"
fi

# ============================================================================
echo ""
echo "${BOLD}────────────────────────────────────────${NC}"
TOTAL=$((PASS + FAIL))
echo "Total: $TOTAL  ${GREEN}PASS: $PASS${NC}  ${RED}FAIL: $FAIL${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo "${RED}Some tests failed.${NC}"
  exit 1
else
  echo "${GREEN}All smoke tests passed.${NC}"
  exit 0
fi
