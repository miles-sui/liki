#!/bin/bash
# Liki — API smoke test
set -uo pipefail

BASE="${1:-http://localhost:8080}"
API="${2:-http://localhost:8081}"
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
  if $HAS_JQ; then
    echo "$1" | jq -r "$2" 2>/dev/null || true
  else
    local k="${2##*.}"
    echo "$1" | /bin/grep -o "\"$k\":\"[^\"]*\"" | head -1 | sed "s/\"$k\":\"//;s/\"$//" || true
  fi
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

check_error_shape() {
  local desc="$1" body="$2"
  if $HAS_JQ; then
    local code=$(echo "$body" | jq -r '.error.code' 2>/dev/null)
    local msg=$(echo "$body" | jq -r '.error.message' 2>/dev/null)
    if [ -n "$code" ] && [ "$code" != "null" ] && [ -n "$msg" ] && [ "$msg" != "null" ]; then
      check "  error shape: $desc" "false" "false"
    else
      check "  error shape: $desc" "false" "true"
    fi
  fi
}

check_has() {
  local desc="$1" body="$2" jqpath="$3"
  if $HAS_JQ; then
    local v=$(echo "$body" | jq -r "$jqpath" 2>/dev/null)
    if [ -n "$v" ] && [ "$v" != "null" ]; then
      check "  has $desc" "false" "false"
    else
      check "  has $desc" "false" "true"
    fi
  fi
}

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
check_has "status" "$b" '.data.status'
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

	  s=$(caddy GET /llms.txt)
	  check_200 "GET /llms.txt" "$s"
	  b=$(body)
	  check "  non-empty" "false" "$([ -z "$b" ] && echo true || echo false)"
	  check "  has skill" "false" "$(echo "$b" | /bin/grep -q '/skills' && echo false || echo true)"

	  # Skill files
	  s=$(caddy GET /skills/liki.md)
	  check_200 "GET /skills/liki.md" "$s"
	  b=$(body)
	  check "  has version" "false" "$(echo "$b" | /bin/grep -q 'version:' && echo false || echo true)"
	  check "  has brand def" "false" "$(echo "$b" | /bin/grep -q '灵机' && echo false || echo true)"
	  check "  has tool calls" "false" "$(echo "$b" | /bin/grep -q 'POST /api/' && echo false || echo true)"

	  s=$(caddy GET /skills/report-chart.md)
	  check_200 "GET /skills/report-chart.md" "$s"
	  b=$(body)
	  check "  has report structure" "false" "$(echo "$b" | /bin/grep -q '格局总论' && echo false || echo true)"

	  s=$(caddy GET /skills/report-bond.md)
	  check_200 "GET /skills/report-bond.md" "$s"
	  b=$(body)
	  check "  has bond structure" "false" "$(echo "$b" | /bin/grep -q '合盘报告模板' && echo false || echo true)"

	  s=$(caddy GET /skills/report-naming.md)
	  check_200 "GET /skills/report-naming.md" "$s"
	  b=$(body)
	  check "  has naming structure" "false" "$(echo "$b" | /bin/grep -q '候选名字速览' && echo false || echo true)"
	  else
	  echo "  (Caddy not running — skipping static checks)"
	  fi

# ============================================================================
# Payment
# ============================================================================
echo ""
echo "${BOLD}── Payment ──${NC}"
s=$(api POST /api/payments/checkout '{"order_id":"nonexistent","email":"test@example.com"}')
check_404 "POST /api/payments/checkout (no order)" "$s"
b=$(body)
check_error_shape "checkout 404" "$b"

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
b=$(body)
check_has "build_time" "$b" '.data.build_time'

s=$(api GET /api/location)
check_200 "GET /api/location" "$s"
b=$(body)
check_has "country" "$b" '.data.country'

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
b=$(body)
check_has "greeting" "$b" '.data.greeting'

s=$(api GET /api/agent/session)
check_400 "GET /api/agent/session (missing session_id)" "$s"

s=$(api GET '/api/agent/session?session_id=nonexistent')
check_404 "GET /api/agent/session (not found)" "$s"

# Chat SSE — verify headers
chat_code=$(curl -s -w '%{http_code}' -o "$TMP/chat_body" "$API/api/agent/chat" \
  -H 'Content-Type: application/json' -d '{"message":"hello"}' \
  -D "$TMP/chat_headers")
check_200 "POST /api/agent/chat" "$chat_code"
chat_ct=$(/bin/grep -i 'content-type:' "$TMP/chat_headers" 2>/dev/null | tr -d '\r' || true)
check "  SSE content-type" "false" "$(echo "$chat_ct" | /bin/grep -q 'text/event-stream' && echo false || echo true)"
chat_sid=$(/bin/grep -i 'x-session-id:' "$TMP/chat_headers" 2>/dev/null | tr -d '\r' || true)
check "  has X-Session-ID" "false" "$([ -n "$chat_sid" ] && echo false || echo true)"

s=$(api POST /api/agent/chat '{"message":""}')
check_400 "POST /api/agent/chat (empty message)" "$s"

# ============================================================================
# Huangli
# ============================================================================
echo ""
echo "${BOLD}── Huangli ──${NC}"

s=$(api GET '/api/huangli/date?date=2026-06-19&event=嫁娶')
check_200 "GET /api/huangli/date" "$s"
b=$(body)
check_has "date entry" "$b" '.data.entry'

s=$(api GET '/api/huangli/month?month=2026-06&event=嫁娶')
check_200 "GET /api/huangli/month" "$s"
b=$(body)
check_has "month entries" "$b" '.data.entries'

HL_BD="{\"birth\":$BT,\"event_type\":\"嫁娶\",\"date\":\"2026-06-19\"}"
s=$(api POST /api/huangli/bond/date "$HL_BD")
check_200 "POST /api/huangli/bond/date" "$s"
b=$(body)
check_has "bond date" "$b" '.data.entry'

HL_BM="{\"birth\":$BT,\"event_type\":\"嫁娶\",\"month\":\"2026-06\"}"
s=$(api POST /api/huangli/bond/month "$HL_BM")
check_200 "POST /api/huangli/bond/month" "$s"
b=$(body)
check_has "bond month" "$b" '.data.entries'

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
b=$(body)
check_has "minggua gua" "$b" '.data.gua'
check_has "minggua group" "$b" '.data.group'

s=$(api POST /api/bazhai/chart "$BR")
check_200 "POST /api/bazhai/chart" "$s"
b=$(body)
check_has "bazhai ming_gua" "$b" '.data.ming_gua'
check_has "bazhai ba_zhai_dirs" "$b" '.data.ba_zhai_dirs'
check_has "bazhai year_stars" "$b" '.data.year_stars'

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
b=$(body)
check_has "sanyuan current" "$b" '.data.current'

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
b=$(body)
	check_has "nian.gan" "$b" '.data.nian.gan'
	check_has "yue.gan" "$b" '.data.yue.gan'
	check_has "ri.gan" "$b" '.data.ri.gan'
	check_has "shi.gan" "$b" '.data.shi.gan'
	check_has "fu_yi" "$b" '.data.fu_yi.yong'
	check_has "da_yun" "$b" '.data.da_yun'
	check_has "wuxing_count" "$b" '.data.wuxing_count'

# BaZi bond — returns Bond struct directly
BOND="{\"a\":$BR_A,\"b\":$BR_B}"
s=$(api POST /api/bazi/bond "$BOND")
check_200 "POST /api/bazi/bond" "$s"
b=$(body)
check_has "bond zhu_cross" "$b" '.data.zhu_cross'
check_has "bond shi_shen_cross" "$b" '.data.shi_shen_cross'
check_has "bond nayin_cross" "$b" '.data.nayin_cross'
check_has "bond structure" "$b" '.data.structure'

	# BaZi luck cycles (gender required for 顺排/逆排)
	s=$(api POST /api/bazi/liunian "{\"year\":2026,\"gender\":\"male\",\"birth\":$BT}")
	check_200 "POST /api/bazi/liunian" "$s"
	check_has "liunian" "$b" '.data'

	s=$(api POST /api/bazi/liuyue "{\"year\":2026,\"month\":6,\"gender\":\"male\",\"birth\":$BT}")
	check_200 "POST /api/bazi/liuyue" "$s"

	s=$(api POST /api/bazi/liuri "{\"year\":2026,\"month\":6,\"day\":19,\"gender\":\"male\",\"birth\":$BT}")
	check_200 "POST /api/bazi/liuri" "$s"

	s=$(api POST /api/bazi/liushi "{\"year\":2026,\"month\":6,\"day\":19,\"hour\":14,\"gender\":\"male\",\"birth\":$BT}")
	check_200 "POST /api/bazi/liushi" "$s"

s=$(api POST /api/bazi/xiaoyun "{\"birth\":$BT,\"gender\":\"male\",\"count\":10}")
check_200 "POST /api/bazi/xiaoyun" "$s"

s=$(api POST /api/bazi/xiaoxian '{"gender":"male","count":10}')
check_200 "POST /api/bazi/xiaoxian" "$s"

# Negative
s=$(api POST /api/bazi/chart 'not-json')
check_400 "POST /api/bazi/chart (bad json)" "$s"
check_error_shape "bazi bad json" "$(body)"

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
check_has "ziwei palaces" "$b" '.data.palaces'
check_has "ziwei si_hua" "$b" '.data.si_hua'
check_has "ziwei ju_shu" "$b" '.data.ju_shu'

# Dependent endpoints need the full chart
if $HAS_JQ; then
  ZW_CHART=$(echo "$b" | jq -c '.data' 2>/dev/null)

  s=$(api POST /api/ziwei/daxian "{\"chart\":$ZW_CHART,\"gender\":\"male\"}")
  check_200 "POST /api/ziwei/daxian" "$s"

  s=$(api POST /api/ziwei/liunian "{\"liu_year\":2026,\"chart\":$ZW_CHART}")
  check_200 "POST /api/ziwei/liunian" "$s"

  s=$(api POST /api/ziwei/liuyue "{\"liu_year\":2026,\"lunar_month\":5,\"chart\":$ZW_CHART}")
  check_200 "POST /api/ziwei/liuyue" "$s"

  s=$(api POST /api/ziwei/liuri "{\"liu_year\":2026,\"lunar_month\":5,\"lunar_day\":15,\"chart\":$ZW_CHART}")
  check_200 "POST /api/ziwei/liuri" "$s"

  s=$(api POST /api/ziwei/bond "{\"a\":$ZW_CHART,\"b\":$ZW_CHART}")
  check_200 "POST /api/ziwei/bond" "$s"
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
check_has "qimen pan" "$b" '.data'

s=$(api POST /api/qimen/pan "{\"birth\":$BT,\"kind\":\"ri\"}")
check_200 "POST /api/qimen/pan (ri)" "$s"

s=$(api POST /api/qimen/pan "{\"birth\":$BT}")
check_200 "POST /api/qimen/pan (default)" "$s"

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
b=$(body)
check_has "liuyao ben_gua" "$b" '.data.ben_gua'
check_has "liuyao lines" "$b" '.data.lines'

LYF="{\"birth\":$BT,\"yong_shen\":\"父母\",\"fixed\":[6,7,8,9,6,7]}"
s=$(api POST /api/liuyao/chart "$LYF")
check_200 "POST /api/liuyao/chart (fixed)" "$s"

# Negative: bad fixed values (1-5 are invalid)
s=$(api POST /api/liuyao/chart "{\"birth\":$BT,\"fixed\":[1,2,3,4,5,6]}")
check_422 "POST /api/liuyao/chart (bad fixed)" "$s"
check_error_shape "liuyao bad fixed" "$(body)"

# ============================================================================
# Qiming
# ============================================================================
echo ""
echo "${BOLD}── Qiming ──${NC}"

s=$(api POST /api/qiming/wuge '{"surname":"李","yong_shen":"水","xi_shen":["金"]}')
check_200 "POST /api/qiming/wuge" "$s"
check_has "wuge" "$b" '.data'

s=$(api POST /api/qiming/evaluate '{"surname":"李","given_name":"沐泽","yong_shen":"水"}')
check_200 "POST /api/qiming/evaluate" "$s"
b=$(body)
check_has "evaluate wu_ge" "$b" '.data.wu_ge'
check_has "evaluate san_cai" "$b" '.data.san_cai'

s=$(api POST /api/qiming/detail '{"surname":"李","names":["沐洪"]}')
check_200 "POST /api/qiming/detail" "$s"
check_has "detail" "$b" '.data'

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
