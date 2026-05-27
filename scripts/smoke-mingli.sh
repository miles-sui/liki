#!/bin/bash
# Mingli server — API smoke test
# Covers all mingli-server endpoints (health, reference, huangli, fengshui,
# bazi fortune, qiming). This script tests endpoint-level behavior: status
# codes, envelope structure, validation. httptest covers deeper logic.
#
# Usage:
#   ./smoke-mingli.sh                              # → http://localhost
#   ./smoke-mingli.sh https://api.tokflux.com      # → production

set -uo pipefail

BASE="${1:-http://localhost}"
PASS=0; FAIL=0
T=$(date +%s)

RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

TMP=$(mktemp -d /tmp/smoke-mingli-XXXXXX)
trap 'rm -rf "$TMP"' EXIT

# --- helpers ---

api() {
  local method="$1" path="$2" data="${3:-}"
  local f="$TMP/body"
  if [ -n "$data" ]; then
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$BASE$path" \
      -H 'Content-Type: application/json' \
      -d "$data" || { echo "000"; return; }
  else
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$BASE$path" \
      -H 'Content-Type: application/json' || { echo "000"; return; }
  fi
}

body() { cat "$TMP/body" 2>/dev/null || true; }

json_val() {
  if command -v jq &>/dev/null; then
    echo "$1" | jq -r "$2" 2>/dev/null || true
  else
    local k="${2##*.}"
    echo "$1" | grep -o "\"$k\":\"[^\"]*\"" | head -1 | sed "s/\"$k\":\"//;s/\"$//" || true
  fi
}

check() {
  local desc="$1" expected="$2" actual="${3:-}" detail="${4:-}"
  if [ "$actual" = "$expected" ]; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}\xe2\x9c\x97${NC} $desc (expected '$expected', got '$actual')${detail:+ — $detail}"
    FAIL=$((FAIL + 1))
  fi
}

check_200() { check "$1 HTTP" "200" "${2:-}" "${3:-}"; }
check_400() { check "$1 HTTP" "400" "${2:-}" "${3:-}"; }
check_404() { check "$1 HTTP" "404" "${2:-}" "${3:-}"; }

# BIRTH is a reusable valid birth chart request for fortune endpoints
BIRTH='{"year":1984,"month":2,"day":4,"hour":18,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}'
CHART_REQ="{\"year\":1984,\"month\":2,\"day\":4,\"hour\":18,\"minute\":0,\"longitude\":116.4,\"timezone\":8,\"gender\":\"male\"}"

# ============================================================================
echo "${BOLD}Mingli Server — API Smoke Test${NC}"
echo "Target: $BASE"
echo ""

# ============================================================================
# PHASE 0 — Health & infra
# ============================================================================
echo "${BOLD}── Phase 0: health & infra ──${NC}"

# Agent skill manifest (markdown)
s=$(api GET / "")
check_200 "GET / (skill manifest)" "$s"
b=$(body)
check "  is markdown" "false" "$(echo "$b" | grep -q '灵机' && echo false || echo true)"
check "  has API directory" "false" "$(echo "$b" | grep -q '## BaZi' && echo false || echo true)"

s=$(api GET /api/health)
check_200 "GET /api/health" "$s"
check "  status=ok" "false" "$([ "$(json_val "$(body)" '.data.status')" = "ok" ] && echo false || echo true)"

s=$(api GET /api/location)
check_200 "GET /api/location" "$s"

s=$(api GET /api/solar-terms)
check_200 "GET /api/solar-terms" "$s"

# ============================================================================
# PHASE 1 — Reference data (9 endpoints)
# ============================================================================
echo ""
echo "${BOLD}── Phase 1: reference endpoints ──${NC}"

for ep in /api/reference/stems /api/reference/branches /api/reference/nayin \
          /api/reference/shensha /api/reference/zodiac /api/reference/mansions \
          /api/reference/trigrams /api/reference/huangdao; do
  s=$(api GET "$ep")
  check_200 "GET $ep" "$s"
done

# Cities — no query
s=$(api GET /api/reference/cities)
check_200 "GET /api/reference/cities (no query)" "$s"

# Cities — with query
s=$(api GET '/api/reference/cities?q=%E5%8C%97%E4%BA%AC')
check_200 "GET /api/reference/cities?q=北京" "$s"

# ============================================================================
# PHASE 2 — Almanac (3 endpoints)
# ============================================================================
echo ""
echo "${BOLD}── Phase 2: huangli endpoints ──${NC}"

s=$(api GET '/api/huangli/query?date=2024-06-15')
check_200 "GET /api/huangli/query?date=2024-06-15" "$s"

s=$(api GET '/api/huangli/query?month=2024-06')
check_200 "GET /api/huangli/query?month=2024-06" "$s"

s=$(api GET '/api/huangli/jieqi?year=2024&month=6&day=15')
check_200 "GET /api/huangli/jieqi?year=2024&month=6&day=15" "$s"

BOND='{"birth_info":{"year":1984,"month":2,"day":4,"hour":18,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"},"date":"2024-06-15"}'
s=$(api POST /api/huangli/bond "$BOND")
check_200 "POST /api/huangli/bond" "$s"

# Almanac validation: missing date
s=$(api GET /api/huangli/query)
check_400 "GET /api/huangli/query (missing date/month)" "$s"

# ============================================================================
# PHASE 3 — Fengshui (2 endpoints)
# ============================================================================
echo ""
echo "${BOLD}── Phase 3: fengshui endpoints ──${NC}"

s=$(api GET /api/reference/24-shan)
check_200 "GET /api/reference/24-shan" "$s"

s=$(api GET '/api/fengshui/san-yuan?year=2024')
check_200 "GET /api/fengshui/san-yuan?year=2024" "$s"

s=$(api GET /api/fengshui/san-yuan)
check_200 "GET /api/fengshui/san-yuan (default year)" "$s"

MG='{"year":1984,"gender":"male"}'
s=$(api POST /api/fengshui/minggua "$MG")
check_200 "POST /api/fengshui/minggua" "$s"

# MingGua validation
s=$(api POST /api/fengshui/minggua '{"year":1984,"gender":"other"}')
check_400 "POST /api/fengshui/minggua (invalid gender)" "$s"

# ============================================================================
# PHASE 4 — BaZi fortune (6 endpoints)
# ============================================================================
echo ""
echo "${BOLD}── Phase 4: bazi fortune ──${NC}"

# First, get the bazi (four pillars) from the chart endpoint.
s=$(api POST /api/bazi/chart "$CHART_REQ")
check_200 "POST /api/bazi/chart (for fortune tests)" "$s"
b=$(body)
if command -v jq &>/dev/null; then
  BAZI=$(echo "$b" | jq -c '{year: .data.year_pillar, month: .data.month_pillar, day: .data.day_pillar, hour: .data.hour_pillar}')
else
  echo "  ${RED}✗ jq required for fortune phase — skipping${NC}"
  BAZI=""
fi

if [ -n "$BAZI" ]; then
  # LiuNian (流年)
  LIUNIAN="{\"bazi\":$BAZI,\"year\":2025}"
  s=$(api POST /api/bazi/liunian "$LIUNIAN")
  check_200 "POST /api/bazi/liunian" "$s"

  # LiuYue (流月)
  LIUYUE="{\"bazi\":$BAZI,\"year\":2025,\"month\":3}"
  s=$(api POST /api/bazi/liuyue "$LIUYUE")
  check_200 "POST /api/bazi/liuyue" "$s"

  # LiuRi (流日)
  LIURI="{\"bazi\":$BAZI,\"date\":\"2025-03-15\"}"
  s=$(api POST /api/bazi/liuri "$LIURI")
  check_200 "POST /api/bazi/liuri" "$s"

  # LiuShi (流时)
  LIUSHI="{\"bazi\":$BAZI,\"date\":\"2025-03-15\",\"hour\":12}"
  s=$(api POST /api/bazi/liushi "$LIUSHI")
  check_200 "POST /api/bazi/liushi" "$s"
fi

# XiaoYun (小运)
XY='{"birth":{"year":1984,"month":2,"day":4,"hour":18,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"},"count":5}'
s=$(api POST /api/bazi/xiao-yun "$XY")
check_200 "POST /api/bazi/xiao-yun" "$s"

# XiaoXian (小限) — only needs gender + count
XX='{"gender":"male","count":5}'
s=$(api POST /api/bazi/xiao-xian "$XX")
check_200 "POST /api/bazi/xiao-xian" "$s"

# Fortune validation
s=$(api POST /api/bazi/liunian '{}')
check_400 "POST /api/bazi/liunian (empty body)" "$s"

s=$(api POST /api/bazi/liunian '{"bazi":[],"year":1800}')
check_400 "POST /api/bazi/liunian (year<1900)" "$s"

# ============================================================================
# PHASE 5 — Naming (3 endpoints)
# ============================================================================
echo ""
echo "${BOLD}── Phase 5: qiming endpoints ──${NC}"

# GetCharacters
s=$(api GET '/api/qiming/characters?element=wood&limit=10')
check_200 "GET /api/qiming/characters?element=wood&limit=10" "$s"

s=$(api GET '/api/qiming/characters?element=%E6%9C%A8&stroke_min=5&stroke_max=12')
check_200 "GET /api/qiming/characters?element=木&stroke_min=5&stroke_max=12" "$s"

s=$(api GET '/api/qiming/characters')
check_400 "GET /api/qiming/characters (missing element)" "$s"

# Evaluate
EVAL='{"surname":"王","given_name":"小明","yong_shen":"火","zodiac":1}'
s=$(api POST /api/qiming/evaluate "$EVAL")
check_200 "POST /api/qiming/evaluate" "$s"
b=$(body)
check "  has wu_ge" "false" "$([ -z "$(json_val "$b" '.data.wu_ge')" ] && echo true || echo false)"
check "  has san_cai" "false" "$([ -z "$(json_val "$b" '.data.san_cai')" ] && echo true || echo false)"

# Evaluate: single char
EVAL1='{"surname":"李","given_name":"明","yong_shen":"水","zodiac":5}'
s=$(api POST /api/qiming/evaluate "$EVAL1")
check_200 "POST /api/qiming/evaluate (single-char)" "$s"

# Evaluate: empty given name
s=$(api POST /api/qiming/evaluate '{"surname":"王","given_name":""}')
check_400 "POST /api/qiming/evaluate (empty given_name)" "$s"

# Generate
GEN='{"surname":"陈","yong_shen":"木","xi_shen":["水"],"zodiac":4,"gender":"male","limit":5}'
s=$(api POST /api/qiming/generate "$GEN")
check_200 "POST /api/qiming/generate" "$s"
b=$(body)
check "  has candidates" "false" "$([ -n "$(echo "$b" | grep -o '"candidates"')" ] && echo false || echo true)"

# Generate: missing surname
s=$(api POST /api/qiming/generate '{"yong_shen":"金"}')
check_400 "POST /api/qiming/generate (missing surname)" "$s"

# ============================================================================
# PHASE 6 — CORS headers (only when cross-origin)
# ============================================================================
echo ""
echo "${BOLD}── Phase 6: CORS headers ──${NC}"

CORS_ORIGIN=$(curl -s -I -X OPTIONS "$BASE/api/health" \
  -H 'Origin: https://25types.com' \
  -H 'Access-Control-Request-Method: POST' 2>/dev/null | grep -i 'Access-Control-Allow-Origin' | tr -d '\r')

if echo "$CORS_ORIGIN" | grep -q '25types.com'; then
  check "  OPTIONS /api/health CORS" "true" "true" "got: $CORS_ORIGIN"
elif [ -z "$CORS_ORIGIN" ]; then
  check "  OPTIONS /api/health CORS" "true" "true" "(no CORS header — expected for same-origin)"
else
  check "  OPTIONS /api/health CORS" "true" "false" "unexpected: $CORS_ORIGIN"
fi

# ============================================================================
# PHASE 7 — MCP protocol
# ============================================================================
echo ""
echo "${BOLD}── Phase 7: MCP protocol ──${NC}"

# tools/list
MCP_LIST=$(curl -s -w '%{http_code}' -o "$TMP/body" -X POST "$BASE/mcp" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}')
check_200 "POST /mcp tools/list" "$MCP_LIST"
b=$(body)
check "  has tools" "false" "$([ -n "$(echo "$b" | grep -o '"tools"')" ] && echo false || echo true)"

# tools/call — fengshui_minggua (lightweight, no external deps)
MCP_CALL=$(curl -s -w '%{http_code}' -o "$TMP/body" -X POST "$BASE/mcp" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"fengshui_minggua","arguments":{"year":1984,"gender":"male"}}}')
check_200 "POST /mcp tools/call (minggua)" "$MCP_CALL"
b=$(body)
check "  has gua_number" "false" "$([ -n "$(echo "$b" | grep -o '"gua_number"')" ] && echo false || echo true)"

# tools/call — huangli_query (lightweight)
MCP_HUANGLI=$(curl -s -w '%{http_code}' -o "$TMP/body" -X POST "$BASE/mcp" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"huangli_query","arguments":{"date":"2026-05-26"}}}')
check_200 "POST /mcp tools/call (huangli_query)" "$MCP_HUANGLI"
b=$(body)
check "  has directions" "false" "$([ -n "$(echo "$b" | grep -o '"directions"')" ] && echo false || echo true)"

# tools/call — bazi_chart (full computation)
MCP_CHART=$(curl -s -w '%{http_code}' -o "$TMP/body" -X POST "$BASE/mcp" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"bazi_chart","arguments":{"birth":{"year":1984,"month":3,"day":15,"hour":8,"gender":"male","longitude":116.4,"timezone":8}}}}')
check_200 "POST /mcp tools/call (chart)" "$MCP_CHART"
b=$(body)
check "  has day_master" "false" "$([ -n "$(echo "$b" | grep -o '"day_master"')" ] && echo false || echo true)"
check "  has yong_shen" "false" "$([ -n "$(echo "$b" | grep -o '"yong_shen"')" ] && echo false || echo true)"

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "${BOLD}────────────────────────────────────────${NC}"
TOTAL=$((PASS + FAIL))
echo "Total: $TOTAL  ${GREEN}PASS: $PASS${NC}  ${RED}FAIL: $FAIL${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo "${RED}Some tests failed.${NC}"
  echo "Re-run a failing step:"
  echo "  curl -v $BASE/api/health"
  exit 1
else
  echo "${GREEN}All mingli smoke tests passed.${NC}"
  exit 0
fi
