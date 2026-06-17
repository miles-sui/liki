#!/bin/bash
# LingJi — API smoke test
set -uo pipefail

BASE="${1:-http://localhost:8080}"
# API goes directly to Go (bypasses Caddy so X-Forwarded-For rotation works).
API="${2:-http://localhost:8081}"
PASS=0; FAIL=0
HAS_JQ=false

RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

TMP=$(mktemp -d /tmp/smoke-lingji-XXXXXX)
trap 'rm -rf "$TMP"' EXIT

command -v jq &>/dev/null && HAS_JQ=true

# IP rotation: each fake IP gets its own rate-limit bucket (core: 1 req/s burst 10).
# Caddy overwrites X-Forwarded-For, so we hit Go :8081 directly.
# Use $RANDOM instead of a counter — api() is called in $() subshells,
# so variable mutations don't propagate back to the parent.

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

# caddy() — for static file routes served by Caddy (not API).
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
    echo "$1" | grep -o "\"$k\":\"[^\"]*\"" | head -1 | sed "s/\"$k\":\"//;s/\"$//" || true
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

check_total() {
  local desc="$1" body="$2" expected="$3"
  if $HAS_JQ; then
    local t=$(echo "$body" | jq '.data.total' 2>/dev/null)
    check "  $desc total=$expected" "$expected" "$t"
  fi
}

echo "${BOLD}LingJi — API Smoke Test${NC}"
echo "Target: Caddy=$BASE API=$API"
$HAS_JQ && echo "jq: available" || echo "jq: not available (limited validation)"
echo ""

# --- Shared test data ---
BIRTH_C='{"year":1984,"month":2,"day":4,"hour":18,"minute":30,"longitude":116.4,"timezone":8,"gender":"male"}'
BIRTH_A='{"year":1990,"month":3,"day":20,"hour":10,"minute":30,"longitude":120,"timezone":8,"gender":"male"}'
BIRTH_B='{"year":1992,"month":7,"day":8,"hour":14,"minute":30,"longitude":120,"timezone":8,"gender":"female"}'

# Hardcoded numeric bazi for 1984-02-04 18:30 (甲子 丁丑 戊辰 辛酉)
BAZI_NUM='{"nian":{"gan":1,"zhi":1},"yue":{"gan":4,"zhi":2},"ri":{"gan":5,"zhi":5},"shi":{"gan":8,"zhi":10}}'
DAYUN_GANZHI='{"gan":5,"zhi":3}'
# YongShen for 1984 male (甲子年 戊辰日)
YONGSHEN='{"fuyi":{"qiangruo":"身强","geju":"建禄格","yong":"木","xi":"水","ji":"土"},"tiaohou":{"season":"冬","yong":"火","xi":"木","ji":"木"}}'

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
# Homepage
# ============================================================================
echo ""
echo "${BOLD}── Homepage ──${NC}"
s=$(caddy GET /)
check_200 "GET /" "$s"
b=$(body)
check "  has 灵机" "false" "$(echo "$b" | grep -q '灵机' && echo false || echo true)"

# ============================================================================
# llms.txt
# ============================================================================
echo ""
echo "${BOLD}── llms.txt ──${NC}"
s=$(caddy GET /llms.txt)
check_200 "GET /llms.txt" "$s"
b=$(body)
check "  non-empty" "false" "$([ -z "$b" ] && echo true || echo false)"

# ============================================================================
# Payment checkout / download / webhook
# ============================================================================
echo ""
echo "${BOLD}── Payment ──${NC}"
CHECKOUT='{"order_id":"nonexistent","email":"test@example.com"}'
s=$(api POST /api/payments/checkout "$CHECKOUT")
check_404 "POST /api/payments/checkout (no order)" "$s"

s=$(api GET "/api/orders/nonexistent/download")
check_302 "GET /api/orders/nonexistent/download" "$s"

# Webhook with no signature
s=$(api POST /api/payments/webhook '{"type":"order.paid","data":{"order_id":"test"}}')
check_400 "POST /api/payments/webhook (no sig)" "$s"

# ============================================================================
# Reference data
# ============================================================================
echo ""
echo "${BOLD}── Reference Data ──${NC}"

s=$(api GET /api/reference/stems)
check_200 "GET /api/reference/stems" "$s"
b=$(body)
check_total "stems" "$b" "10"
check_has "stem name" "$b" '.data.items[0].name'
check_has "stem element" "$b" '.data.items[0].element'
check_has "stem yin_yang" "$b" '.data.items[0].yin_yang'

s=$(api GET /api/reference/branches)
check_200 "GET /api/reference/branches" "$s"
b=$(body)
check_total "branches" "$b" "12"
check_has "branch name" "$b" '.data.items[0].name'
check_has "hidden_stems" "$b" '.data.items[0].hidden_stems'

s=$(api GET /api/reference/nayin)
check_200 "GET /api/reference/nayin" "$s"
b=$(body)
check_total "nayin" "$b" "60"
check_has "nayin name" "$b" '.data.items[0].nayin'

s=$(api GET /api/reference/shensha)
check_200 "GET /api/reference/shensha" "$s"
b=$(body)
check_has "shensha envelope" "$b" '.data'

s=$(api GET /api/reference/zodiac)
check_200 "GET /api/reference/zodiac" "$s"
b=$(body)
check_has "six_he" "$b" '.data.six_he'
check_has "triple_he" "$b" '.data.triple_he'
check_has "six_chong" "$b" '.data.six_chong'
check_has "six_hai" "$b" '.data.six_hai'

s=$(api GET /api/reference/mansions)
check_200 "GET /api/reference/mansions" "$s"
b=$(body)
check_total "mansions" "$b" "28"

s=$(api GET /api/reference/trigrams)
check_200 "GET /api/reference/trigrams" "$s"
b=$(body)
check_total "trigrams" "$b" "8"

s=$(api GET /api/reference/huangdao)
check_200 "GET /api/reference/huangdao" "$s"
b=$(body)
check_total "huangdao" "$b" "12"

s=$(api GET /api/reference/24-shan)
check_200 "GET /api/reference/24-shan" "$s"
b=$(body)
check_has "mountains" "$b" '.data.mountains'

# ============================================================================
# Solar terms
# ============================================================================
echo ""
echo "${BOLD}── Solar Terms ──${NC}"
s=$(api GET /api/solar-terms)
check_200 "GET /api/solar-terms" "$s"
b=$(body)
check_has "year" "$b" '.data.year'
check_has "current" "$b" '.data.current'
check_has "months" "$b" '.data.months'
if $HAS_JQ; then
  mc=$(echo "$b" | jq '.data.months | length' 2>/dev/null)
  check "  12 solar months" "12" "$mc"
fi

# ============================================================================
# Huangli
# ============================================================================
echo ""
echo "${BOLD}── Huangli ──${NC}"

# Single date query
s=$(api GET '/api/huangli/query?date=2026-06-15&event=wedding')
check_200 "GET /api/huangli/query?date=&event=wedding" "$s"
b=$(body)
check_has "day pillar gan" "$b" '.data.day_pillar.gan'
check_has "jian_chu" "$b" '.data.jian_chu'

# Month query
s=$(api GET '/api/huangli/query?month=2026-06&event=travel')
check_200 "GET /api/huangli/query?month=&event=travel" "$s"
b=$(body)
check_has "year_month" "$b" '.data.year_month'
check_has "days" "$b" '.data.days'

# Jieqi
s=$(api GET '/api/huangli/jieqi?year=2026&month=6&day=15')
check_200 "GET /api/huangli/jieqi" "$s"
b=$(body)
check_has "jieqi_depth" "$b" '.data.jieqi_depth'
check_has "ren_yuan" "$b" '.data.ren_yuan'

# Huangli bond — single date
HL_BOND="{\"birth_info\":$BIRTH_C,\"date\":\"2026-06-15\",\"event_type\":\"wedding\"}"
s=$(api POST /api/huangli/bond "$HL_BOND")
check_200 "POST /api/huangli/bond (date)" "$s"
b=$(body)
check_has "bond day pillar" "$b" '.data.day_pillar.gan'
check_has "bond gan_relation" "$b" '.data.gan_relation'

# Huangli bond — month mode
HL_BOND_MONTH="{\"birth_info\":$BIRTH_C,\"month\":\"2026-06\",\"event_type\":\"wedding\"}"
s=$(api POST /api/huangli/bond "$HL_BOND_MONTH")
check_200 "POST /api/huangli/bond (month)" "$s"
b=$(body)
check_has "bond month year_month" "$b" '.data.year_month'
check_has "bond month days" "$b" '.data.days'

# Huangli negative: missing event
s=$(api GET /api/huangli/query)
check_422 "GET /api/huangli/query (missing event)" "$s"
b=$(body)
check_error_shape "huangli missing event" "$b"

# Huangli negative: missing date/month with event
s=$(api GET '/api/huangli/query?event=wedding')
check_422 "GET /api/huangli/query (missing date/month)" "$s"

# Huangli bond negative: missing birth_info
s=$(api POST /api/huangli/bond '{"date":"2026-06-15","event_type":"wedding"}')
check_422 "POST /api/huangli/bond (missing birth)" "$s"

# ============================================================================
# Fengshui
# ============================================================================
echo ""
echo "${BOLD}── Fengshui ──${NC}"

s=$(api GET '/api/fengshui/san-yuan?year=2026')
check_200 "GET /api/fengshui/san-yuan?year=2026" "$s"
b=$(body)
check_has "current" "$b" '.data.current'
check_has "all_periods" "$b" '.data.all_periods'
if $HAS_JQ; then
  pc=$(echo "$b" | jq '.data.all_periods | length' 2>/dev/null)
  check "  9 periods" "9" "$pc"
fi

# San-yuan negative: missing year
s=$(api GET /api/fengshui/san-yuan)
check_422 "GET /api/fengshui/san-yuan (missing year)" "$s"

# MingGua
MG='{"year":1984,"gender":"male"}'
s=$(api POST /api/fengshui/minggua "$MG")
check_200 "POST /api/fengshui/minggua" "$s"
b=$(body)
check_has "ming_gua" "$b" '.data.ming_gua.gua_number'
check_has "all_trigrams" "$b" '.data.all_trigrams'

# MingGua negative: invalid gender
s=$(api POST /api/fengshui/minggua '{"year":1984,"gender":"other"}')
check_422 "POST /api/fengshui/minggua (bad gender)" "$s"
b=$(body)
check_error_shape "minggua bad gender" "$b"

# MingGua negative: missing year
s=$(api POST /api/fengshui/minggua '{"gender":"male"}')
check_422 "POST /api/fengshui/minggua (missing year)" "$s"

# ============================================================================
# BaZi free chart
# ============================================================================
echo ""
echo "${BOLD}── BaZi Free Chart ──${NC}"
s=$(api POST /api/bazi/chart "$BIRTH_C")
check_200 "POST /api/bazi/chart" "$s"
b=$(body)
check_has "day_master" "$b" '.data.riyuan'
check_has "yong_shen" "$b" '.data.yong_shen.fuyi.yong'
check_has "dayun" "$b" '.data.dayun'
check_has "year_pillar" "$b" '.data.nianzhu.gan'
check_has "element_count" "$b" '.data.wuxing'

# BaZi chart negative: bad JSON
s=$(api POST /api/bazi/chart 'not-json')
check_400 "POST /api/bazi/chart (bad json)" "$s"
b=$(body)
check_error_shape "bazi chart bad json" "$b"

# BaZi chart negative: missing gender
s=$(api POST /api/bazi/chart '{"year":1984,"month":2,"day":4,"hour":18,"minute":30,"longitude":116.4,"timezone":8}')
check_422 "POST /api/bazi/chart (missing gender)" "$s"

# ============================================================================
# BaZi bond (free)
# ============================================================================
echo ""
echo "${BOLD}── BaZi Free Bond ──${NC}"
BAZI_BOND='{"a":{"year":1990,"month":3,"day":20,"hour":10,"minute":30,"gender":"male"},"b":{"year":1992,"month":7,"day":8,"hour":14,"minute":30,"gender":"female"}}'
s=$(api POST /api/bazi/bond "$BAZI_BOND")
check_200 "POST /api/bazi/bond" "$s"
b=$(body)
check_has "chart_a" "$b" '.data.chart_a.riyuan'
check_has "chart_b" "$b" '.data.chart_b.riyuan'
check_has "bond pillar_cross" "$b" '.data.bond.zhuzhu_rel'

# BaZi bond negative: missing a
s=$(api POST /api/bazi/bond '{"b":{"year":1992,"month":7,"day":8,"hour":14,"minute":30,"gender":"female"}}')
check_400 "POST /api/bazi/bond (missing a)" "$s"

# ============================================================================
# Fengshui HeCan
# ============================================================================
echo ""
echo "${BOLD}── Fengshui HeCan ──${NC}"
if $HAS_JQ; then
  HECAN="{\"birth_year\":1984,\"gender\":\"male\",\"bazi\":$BAZI_NUM,\"yong_shen\":$YONGSHEN,\"year\":2026}"
  s=$(api POST /api/fengshui/hecan "$HECAN")
  check_200 "POST /api/fengshui/hecan" "$s"
  b=$(body)
  check_has "ming_gua" "$b" '.data.ming_gua.gua_number'
  check_has "ba_zhai_dirs" "$b" '.data.ba_zhai_dirs'
  check_has "year_stars" "$b" '.data.year_stars'

  # HeCan negative: missing bazi
  s=$(api POST /api/fengshui/hecan '{"birth_year":1984,"gender":"male","year":2026}')
  check_422 "POST /api/fengshui/hecan (missing bazi)" "$s"
fi

# ============================================================================
# BaZi luck cycle endpoints
# ============================================================================
echo ""
echo "${BOLD}── BaZi Luck Cycles ──${NC}"

if $HAS_JQ; then
  # LiuNian
  LN_REQ="{\"bazi\":$BAZI_NUM,\"year\":2026,\"current_dayun\":$DAYUN_GANZHI}"
  s=$(api POST /api/bazi/liunian "$LN_REQ")
  check_200 "POST /api/bazi/liunian" "$s"
  b=$(body)
  check_has "liunian year_stem" "$b" '.data.year_stem'
  check_has "liunian shishen" "$b" '.data.shishen'

  # LiuYue
  LY_REQ="{\"bazi\":$BAZI_NUM,\"year\":2026,\"month\":6}"
  s=$(api POST /api/bazi/liuyue "$LY_REQ")
  check_200 "POST /api/bazi/liuyue" "$s"
  b=$(body)
  check_has "liuyue month_stem" "$b" '.data.month_stem'

  # LiuRi
  LR_REQ="{\"bazi\":$BAZI_NUM,\"date\":\"2026-06-15\",\"dayun_pillar\":$DAYUN_GANZHI,\"liunian_pillar\":$DAYUN_GANZHI}"
  s=$(api POST /api/bazi/liuri "$LR_REQ")
  check_200 "POST /api/bazi/liuri" "$s"
  b=$(body)
  check_has "liuri day_stem" "$b" '.data.day_stem'

  # LiuShi
  LS_REQ="{\"bazi\":$BAZI_NUM,\"date\":\"2026-06-15\",\"hour\":12}"
  s=$(api POST /api/bazi/liushi "$LS_REQ")
  check_200 "POST /api/bazi/liushi" "$s"
  b=$(body)
  check_has "liushi hour_stem" "$b" '.data.hour_stem'

  # Negative: liunian with bad bazi
  s=$(api POST /api/bazi/liunian '{"bazi":{"nian":{"gan":0,"zhi":0}},"year":2026}')
  check_422 "POST /api/bazi/liunian (bad bazi)" "$s"
fi

# XiaoYun — uses birth params directly
XY_REQ='{"birth":{"year":1984,"month":2,"day":4,"hour":18,"minute":30,"longitude":116.4,"timezone":8,"gender":"male"},"count":5}'
s=$(api POST /api/bazi/xiao-yun "$XY_REQ")
check_200 "POST /api/bazi/xiao-yun" "$s"
b=$(body)
if $HAS_JQ; then
  check_has "xiaoyun items" "$b" '.data[0].zhi'
fi

# XiaoYun negative: missing birth
s=$(api POST /api/bazi/xiao-yun '{"count":5}')
check_422 "POST /api/bazi/xiao-yun (missing birth)" "$s"

# XiaoXian
XX_REQ='{"gender":"male","count":5}'
s=$(api POST /api/bazi/xiao-xian "$XX_REQ")
check_200 "POST /api/bazi/xiao-xian" "$s"
b=$(body)
if $HAS_JQ; then
  check_has "xiaoxian items" "$b" '.data[0].branch'
fi

# XiaoXian negative: missing gender
s=$(api POST /api/bazi/xiao-xian '{"count":5}')
check_422 "POST /api/bazi/xiao-xian (missing gender)" "$s"

# ============================================================================
# Qiming
# ============================================================================
echo ""
echo "${BOLD}── Qiming ──${NC}"

# Characters — requires element, stroke_min
s=$(api GET '/api/qiming/characters?element=wood&stroke_min=0&stroke_max=30&limit=10')
check_200 "GET /api/qiming/characters" "$s"
b=$(body)
check_has "characters" "$b" '.data.items[0].char'
check_has "character wuxing" "$b" '.data.items[0].wuxing'

# Characters negative: missing element
s=$(api GET '/api/qiming/characters?stroke_min=0&stroke_max=30')
check_422 "GET /api/qiming/characters (missing element)" "$s"
b=$(body)
check_error_shape "qiming missing element" "$b"

# Characters negative: invalid element
s=$(api GET '/api/qiming/characters?element=xyz&stroke_min=0&stroke_max=30')
check_422 "GET /api/qiming/characters (bad element)" "$s"

# Characters negative: missing stroke_min
s=$(api GET '/api/qiming/characters?element=wood')
check_422 "GET /api/qiming/characters (missing stroke_min)" "$s"

# Evaluate
EVAL='{"surname":"王","given_name":"小明","yong_shen":"火","zodiac":1}'
s=$(api POST /api/qiming/evaluate "$EVAL")
check_200 "POST /api/qiming/evaluate" "$s"
b=$(body)
check_has "wu_ge" "$b" '.data.wu_ge'
check_has "san_cai" "$b" '.data.san_cai'
check_has "wuxing_match" "$b" '.data.wuxing_match'

# Evaluate negative: empty given_name
s=$(api POST /api/qiming/evaluate '{"surname":"王","given_name":""}')
check_422 "POST /api/qiming/evaluate (empty given_name)" "$s"
b=$(body)
check_error_shape "qiming empty name" "$b"

# Evaluate negative: missing yong_shen
s=$(api POST /api/qiming/evaluate '{"surname":"王","given_name":"小明"}')
check_422 "POST /api/qiming/evaluate (missing yong_shen)" "$s"

# Generate (free)
GEN='{"surname":"陈","yong_shen":"木","xi_shen":["水"],"zodiac":4,"gender":"male","limit":5}'
s=$(api POST /api/qiming/generate "$GEN")
check_200 "POST /api/qiming/generate" "$s"
b=$(body)
check_has "candidates" "$b" '.data.candidates'
check_has "surname_element" "$b" '.data.surname_element'

# Generate negative: missing surname
s=$(api POST /api/qiming/generate '{"yong_shen":"金"}')
check_422 "POST /api/qiming/generate (missing surname)" "$s"

# Generate negative: invalid yong_shen
s=$(api POST /api/qiming/generate '{"surname":"陈","yong_shen":"x"}')
check_422 "POST /api/qiming/generate (bad yong_shen)" "$s"

# ============================================================================
# Agent / Chat
# ============================================================================
echo ""
echo "${BOLD}── Agent / Chat ──${NC}"

# Greeting
s=$(api GET /api/agent/greeting)
check_200 "GET /api/agent/greeting" "$s"
b=$(body)
check_has "greeting" "$b" '.data.greeting'

# Session restore — negative: missing session_id
s=$(api GET /api/agent/session)
check_400 "GET /api/agent/session (missing session_id)" "$s"

# Session restore — negative: nonexistent session
s=$(api GET '/api/agent/session?session_id=nonexistent')
check_404 "GET /api/agent/session (not found)" "$s"

# Chat SSE — basic smoke: check Content-Type and SSE framing
CHAT_MSG='{"message":"hello"}'
chat_code=$(curl -s -w '%{http_code}' -o "$TMP/chat_body" "$API/api/agent/chat" \
  -H 'Content-Type: application/json' -d "$CHAT_MSG" \
  -D "$TMP/chat_headers")
check_200 "POST /api/agent/chat" "$chat_code"
chat_ct=$(grep -i 'content-type:' "$TMP/chat_headers" 2>/dev/null | tr -d '\r' || true)
check "  SSE content-type" "false" "$(echo "$chat_ct" | grep -q 'text/event-stream' && echo false || echo true)"
chat_sid=$(grep -i 'x-session-id:' "$TMP/chat_headers" 2>/dev/null | tr -d '\r' || true)
check "  has X-Session-ID" "false" "$([ -n "$chat_sid" ] && echo false || echo true)"

# Chat negative: empty message
s=$(api POST /api/agent/chat '{"message":""}')
check_400 "POST /api/agent/chat (empty message)" "$s"

# Version
echo ""
echo "${BOLD}── Version ──${NC}"
s=$(api GET /api/version)
check_200 "GET /api/version" "$s"
b=$(body)
check_has "build_time" "$b" '.data.build_time'

# Analytics
echo ""
echo "${BOLD}── Analytics ──${NC}"
s=$(api POST /api/analytics/pageview '{"path":"/zh/","title":"Home"}')
check_204 "POST /api/analytics/pageview" "$s"

s=$(api GET /api/stats)
check_200 "GET /api/stats" "$s"

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
