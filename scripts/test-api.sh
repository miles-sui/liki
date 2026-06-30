#!/bin/bash
# Liki — JSON-RPC smoke test
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

RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

TMP=$(mktemp -d /tmp/test-api-XXXXXX)
trap 'rm -rf "$TMP"' EXIT

# ── helpers ──────────────────────────────────────────────────────

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

json_val() { echo "$1" | jq -r "$2" 2>/dev/null || true; }

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
check_403()  { check "$1 HTTP" "403" "${2:-}"; }
check_404()  { check "$1 HTTP" "404" "${2:-}"; }
check_422()  { check "$1 HTTP" "422" "${2:-}"; }

# ── RPC helpers ───────────────────────────────────────────────────

RPC_ID=0
RPC_CODE=""
RPC_BODY=""
RPC_DATA=""

# rpc <method> <params_json>
# Sets RPC_CODE (HTTP status) and RPC_BODY (full response).
rpc() {
  local method="$1"
  local params='{}'
  [ $# -ge 2 ] && params="$2"
  RPC_ID=$((RPC_ID + 1))
  local payload
  payload=$(jq -nc --arg m "$method" --argjson p "$params" --argjson id "$RPC_ID" \
    '{jsonrpc:"2.0", id:$id, method:$m, params:$p}')
  RPC_CODE=$(api POST /jsonrpc "$payload")
  RPC_BODY=$(body)
}

# check_rpc <desc> <jq_filter> <expected>
check_rpc() {
  local desc="$1" filter="$2" expected="$3"
  local actual
  actual=$(json_val "$RPC_BODY" "$filter")
  check "$desc" "$expected" "$actual"
}

# check_rpc_ok <desc>
check_rpc_ok() {
  local has_err
  has_err=$(json_val "$RPC_BODY" '.error != null')
  if [ "$has_err" = "false" ]; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} $1"
    PASS=$((PASS + 1))
  else
    local emsg
    emsg=$(json_val "$RPC_BODY" '.error.message')
    echo -e "  ${RED}\xe2\x9c\x97${NC} $1 (RPC error: $emsg)"
    FAIL=$((FAIL + 1))
  fi
}

# check_rpc_err <desc> <expected_code>
check_rpc_err() {
  local desc="$1" expected_code="$2"
  local has_err actual_code
  has_err=$(json_val "$RPC_BODY" '.error != null')
  actual_code=$(json_val "$RPC_BODY" '.error.code')
  if [ "$has_err" = "true" ] && [ "$actual_code" = "$expected_code" ]; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}\xe2\x9c\x97${NC} $desc (expected error $expected_code, has_err=$has_err code=$actual_code)"
    FAIL=$((FAIL + 1))
  fi
}

# ── test data ─────────────────────────────────────────────────────

BT='{"time":"1984-02-04T18:30:00+08:00","longitude":116.4}'
BT_A='{"time":"1990-03-20T10:30:00+08:00","longitude":120}'
BT_B='{"time":"1992-07-08T14:30:00+08:00","longitude":120}'
BR="{\"birth\":$BT,\"gender\":\"male\"}"
BR_A="{\"birth\":$BT_A,\"gender\":\"male\"}"
BR_B="{\"birth\":$BT_B,\"gender\":\"female\"}"

# ============================================================================
echo "${BOLD}Liki — JSON-RPC Smoke Test${NC}"
echo "Target: Caddy=$BASE API=$API"
command -v jq &>/dev/null || { echo "jq is required for RPC smoke tests"; exit 1; }
echo ""

# ============================================================================
# Health
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
if curl -s --connect-timeout 2 -o /dev/null -w '%{http_code}' "$BASE/api/health" 2>/dev/null | /bin/grep -q .; then
  s=$(caddy GET /)
  check_200 "GET /" "$s"
  b=$(body)
  check "  has 靈機" "false" "$(echo "$b" | /bin/grep -q '靈機' && echo false || echo true)"
else
  echo "  (Caddy not running — skipping static checks)"
fi

# ============================================================================
# Wiki
# ============================================================================
echo ""
echo "${BOLD}── Wiki ──${NC}"
if [ "$BASE" = "http://localhost:8080" ] || [ "$BASE" = "http://127.0.0.1:8080" ]; then
  echo "  (local Caddy — skipping wiki checks)"
else
  s=$(caddy GET /wiki/)
  check_200 "GET /wiki/" "$s"
  b=$(body)
  check "  redirects to zh-hant" "false" "$(echo "$b" | /bin/grep -q 'zh-hant/index.html' && echo false || echo true)"
  s=$(caddy GET /wiki/zh-hant/index.html)
  check_200 "GET /wiki/zh-hant/index.html" "$s"
  s=$(caddy GET /wiki/zh-hant/how-to-choose-a-good-name.html)
  check_200 "GET /wiki/zh-hant/how-to-choose-a-good-name.html" "$s"
fi

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

s=$(api POST /api/analytics/pageview '{"path":"/zh-Hans/"}')
check_204 "POST /api/analytics/pageview" "$s"

# ============================================================================
# Agent
# ============================================================================
echo ""
echo "${BOLD}── Agent ──${NC}"

s=$(api POST /api/agent/naming '{"message":"hello","lang":"zh-Hans"}')
check "POST /api/agent/naming (no JWT)" "401" "$s"

s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$API/api/agent/naming" \
  -H 'Content-Type: application/json' \
  -H "Cookie: liki_token=not-a-valid-jwt" \
  -d '{"message":"hello","lang":"zh-Hans"}')
check "POST /api/agent/naming (invalid JWT)" "401" "$s"

# ============================================================================
# JSON-RPC Engine
# ============================================================================
echo ""
echo "${BOLD}── RPC Protocol ──${NC}"

# Parse error
rpc '""' '{}'
check_rpc_err "rpc: invalid method" "-32601"

# Missing jsonrpc
RPC_ID=$((RPC_ID + 1))
s=$(curl -s -w '%{http_code}' -o "$TMP/body" -X POST "$API/jsonrpc" -H 'Content-Type: application/json' -d '{"id":1,"method":"bazi.chart","params":{}}')
check "rpc: missing jsonrpc 200" "200" "$s"
RPC_BODY=$(body)
check_rpc_err "rpc: missing jsonrpc error" "-32600"

# rpc.discover
rpc rpc.discover '{}'
check_rpc_ok "rpc.discover"
check_rpc "  openrpc version" '.result.openrpc' '1.4.1'

# ============================================================================
# BaZi (8 methods)
# ============================================================================
echo ""
echo "${BOLD}── BaZi ──${NC}"

rpc bazi.chart "$BR"
check_rpc_ok "bazi.chart"
check_rpc "  has nian" '.result.data.nian.gan != null' 'true'
check_rpc "  has da_yun" '.result.data.da_yun != null' 'true'
check_rpc "  has san_yuan" '.result.data.san_yuan != null' 'true'

rpc bazi.chart '{"birth":{"time":"1984-02-04T18:30:00+08:00"}}'
check_rpc_err "bazi.chart (missing gender)" "-32000"

rpc bazi.bond "{\"a\":$BR_A,\"b\":$BR_B}"
check_rpc_ok "bazi.bond"
check_rpc "  has zhu_cross" '.result.data.zhu_cross != null' 'true'

rpc bazi.liunian "{\"year\":2026,\"birth\":$BT,\"gender\":\"male\"}"
check_rpc_ok "bazi.liunian"

rpc bazi.liuyue "{\"year\":2026,\"month\":6,\"birth\":$BT,\"gender\":\"male\"}"
check_rpc_ok "bazi.liuyue"

rpc bazi.liuri "{\"year\":2026,\"month\":6,\"day\":15,\"birth\":$BT,\"gender\":\"male\"}"
check_rpc_ok "bazi.liuri"

rpc bazi.liushi "{\"year\":2026,\"month\":6,\"day\":15,\"hour\":12,\"birth\":$BT,\"gender\":\"male\"}"
check_rpc_ok "bazi.liushi"

rpc bazi.xiaoyun "{\"birth\":$BT,\"gender\":\"male\"}"
check_rpc_ok "bazi.xiaoyun"

rpc bazi.xiaoxian '{"gender":"male"}'
check_rpc_ok "bazi.xiaoxian"

# ============================================================================
# ZiWei (6 methods)
# ============================================================================
echo ""
echo "${BOLD}── ZiWei ──${NC}"

rpc ziwei.chart "$BR"
check_rpc_ok "ziwei.chart"
check_rpc "  has palaces" '.result.data.palaces != null' 'true'

# Dependent: ziwei.daxian needs chart from ziwei.chart
ZW_CHART=$(json_val "$RPC_BODY" '.result.data')

rpc ziwei.daxian "{\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.daxian"

rpc ziwei.liunian "{\"liu_year\":2026,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liunian"

rpc ziwei.liuyue "{\"liu_year\":2026,\"lunar_month\":5,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liuyue"

rpc ziwei.liuri "{\"liu_year\":2026,\"lunar_month\":5,\"lunar_day\":10,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liuri"

# ziwei.bond needs two charts
rpc ziwei.chart "$BR_A"
CHART_A=$(json_val "$RPC_BODY" '.result.data')
rpc ziwei.chart "$BR_B"
CHART_B=$(json_val "$RPC_BODY" '.result.data')
rpc ziwei.bond "{\"a\":$CHART_A,\"b\":$CHART_B}"
check_rpc_ok "ziwei.bond"

rpc ziwei.chart '{"birth":{"time":"1984-02-04T18:30:00+08:00"}}'
check_rpc_err "ziwei.chart (missing gender)" "-32000"

# ============================================================================
# QiMen (1 method)
# ============================================================================
echo ""
echo "${BOLD}── QiMen ──${NC}"

rpc qimen.pan "{\"birth\":$BT,\"kind\":\"shi\"}"
check_rpc_ok "qimen.pan (shi)"

rpc qimen.pan "{\"birth\":$BT,\"kind\":\"invalid\"}"
check_rpc_err "qimen.pan (bad kind)" "-32000"

# ============================================================================
# QiMing (4 methods)
# ============================================================================
echo ""
echo "${BOLD}── QiMing ──${NC}"

rpc qiming.wuge '{"surname":"李","yong_shen":"水","xi_shen":["金"]}'
check_rpc_ok "qiming.wuge"

rpc qiming.compose '{"surname":"李","combos":[],"yong_chars":{}}'
check_rpc_ok "qiming.compose"

rpc qiming.detail '{"surname":"李","names":["沐泽","沐恩"]}'
check_rpc_ok "qiming.detail"

rpc qiming.evaluate '{"surname":"李","given_name":"沐泽","yong_shen":"水"}'
check_rpc_ok "qiming.evaluate"

rpc qiming.wuge '{"surname":"李"}'
check_rpc_err "qiming.wuge (missing yong_shen)" "-32000"

rpc qiming.evaluate '{"surname":"李"}'
check_rpc_err "qiming.evaluate (missing params)" "-32000"

# ============================================================================
# Bazhai (2 methods)
# ============================================================================
echo ""
echo "${BOLD}── Bazhai ──${NC}"

rpc bazhai.minggua '{"gender":"male","birth_year":1984}'
check_rpc_ok "bazhai.minggua"

rpc bazhai.chart "$BR"
check_rpc_ok "bazhai.chart"

rpc bazhai.minggua '{"gender":"other","birth_year":1984}'
check_rpc_err "bazhai.minggua (bad gender)" "-32000"

rpc bazhai.minggua '{"gender":"male"}'
check_rpc_ok "bazhai.minggua (missing year, defaults to 0)"

# ============================================================================
# XuanKong (2 methods)
# ============================================================================
echo ""
echo "${BOLD}── XuanKong ──${NC}"

rpc xuankong.sanyuan '{"year":2026}'
check_rpc_ok "xuankong.sanyuan"

rpc xuankong.chart "{\"birth\":$BT,\"sit_mountain\":0,\"face_mountain\":11}"
check_rpc_ok "xuankong.chart"

rpc xuankong.chart "{\"birth\":$BT}"
check_rpc_err "xuankong.chart (missing mountains)" "-32000"

# ============================================================================
# LiuYao (1 method)
# ============================================================================
echo ""
echo "${BOLD}── LiuYao ──${NC}"

rpc liuyao.qigua '{}'
check_rpc_ok "liuyao.qigua"

# Extract yaos from qigua result for use in chart
YAOS=$(json_val "$RPC_BODY" '.result.data.yaos')
YAOS_JSON=$(echo "$YAOS" | jq -c '.')

rpc liuyao.chart "{\"birth\":$BT,\"yaos\":$YAOS_JSON}"
check_rpc_ok "liuyao.chart"

rpc liuyao.chart "{\"birth\":$BT,\"yaos\":[6,7,8,9,6,7],\"yong_shen\":\"妻财\"}"
check_rpc_ok "liuyao.chart (with yong_shen)"

rpc liuyao.chart "{\"birth\":$BT,\"yaos\":[7,8,7,6,9,7]}"
check_rpc_ok "liuyao.chart (mixed yaos)"

# ============================================================================
# Huangli (4 methods)
# ============================================================================
echo ""
echo "${BOLD}── Huangli ──${NC}"

rpc huangli.date '{"date":"2026-06-19","event":"嫁娶"}'
check_rpc_ok "huangli.date"

rpc huangli.month '{"month":"2026-06","event":"嫁娶"}'
check_rpc_ok "huangli.month"

HL_BOND='{"birth":'"$BT"',"event_type":"嫁娶","date":"2026-06-19"}'
rpc huangli.bond.date "$HL_BOND"
check_rpc_ok "huangli.bond.date"

rpc huangli.bond.month '{"birth":'"$BT"',"event_type":"嫁娶","month":"2026-06"}'
check_rpc_ok "huangli.bond.month"

rpc huangli.date '{}'
check_rpc_err "huangli.date (missing params)" "-32000"

HL_BAD_BIRTH='{"birth":{},"event_type":"嫁娶","date":"2026-06-19"}'
rpc huangli.bond.date "$HL_BAD_BIRTH"
check_rpc_err "huangli.bond.date (empty birth)" "-32000"

# ============================================================================
# Infra (2 methods)
# ============================================================================
echo ""
echo "${BOLD}── Infra ──${NC}"

rpc time.now '{}'
check_rpc_ok "time.now"
b=$(body)
check "  has utc" "false" "$(json_val "$b" '.result.data.utc' | grep -q . && echo false || echo true)"
check "  has cst" "false" "$(json_val "$b" '.result.data.cst' | grep -q . && echo false || echo true)"

# city depends on external Nominatim API — skip if SKIP_EXTERNAL is set.
CITY_OK=false
if [ "${SKIP_EXTERNAL:-}" = "1" ]; then
  echo -e "  ${BOLD}↷${NC} city (skipped: SKIP_EXTERNAL=1)"
else
  for _ in 1 2; do
    rpc city '{"city":"Beijing"}'
    if [ "$(json_val "$RPC_BODY" '.error != null')" = "false" ]; then
      CITY_OK=true; break
    fi
    sleep 3
  done
  if $CITY_OK; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} city"
    PASS=$((PASS + 1))
  else
    echo -e "  ${BOLD}↷${NC} city (external API unavailable)"
  fi
fi

# ============================================================================
# Access Policy
# ============================================================================
if curl -s --connect-timeout 2 -o /dev/null -w '%{http_code}' "$BASE/api/health" 2>/dev/null | /bin/grep -q .; then
  echo ""
  echo "${BOLD}── Access Policy ──${NC}"

  if [ "$BASE" = "http://localhost:8080" ] || [ "$BASE" = "http://127.0.0.1:8080" ]; then
    echo "  (local Caddy — skipping access policy checks)"
  else
    s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$BASE/api/agent/naming" \
      -H 'Content-Type: application/json' \
      -H 'Origin: https://evil.com' \
      -d '{"message":"hello"}')
    check_403 "POST /api/agent/naming (external Origin)" "$s"

    s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$BASE/api/orders/nonexistent/retry" \
      -H 'Origin: https://evil.com')
    check_403 "POST /api/orders/{id}/retry (external Origin)" "$s"

    # Free API — accessible by external agents (no Origin header)
    s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$BASE/jsonrpc" \
      -H 'Content-Type: application/json' \
      -d '{"jsonrpc":"2.0","id":1,"method":"bazi.chart","params":{"birth":'"$BT"',"gender":"male"}}')
    check_200 "bazi.chart (no Origin, external agent)" "$s"

    # Free API — also accessible cross-origin
    s=$(curl -s -w '%{http_code}' -o /dev/null -X POST "$BASE/jsonrpc" \
      -H 'Content-Type: application/json' \
      -H 'Origin: https://evil.com' \
      -d '{"jsonrpc":"2.0","id":1,"method":"bazi.chart","params":{"birth":'"$BT"',"gender":"male"}}')
    check_200 "bazi.chart (external Origin ignored)" "$s"
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
  echo "${GREEN}All API tests passed.${NC}"
  exit 0
fi
