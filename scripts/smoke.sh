#!/bin/bash
# 25types — Journey smoke test
# Tests full-stack user journeys through Caddy + Go + SQLite.
# Endpoint-level behavior (status codes, error mapping, envelope) is covered
# by Go httptest (internal/http/*_test.go). This script ONLY tests what
# httptest cannot: Caddy routing, static files, cross-endpoint stateful
# journeys, rate limiting.
#
# Usage:
#   ./smoke.sh                          # → http://localhost
#   ./smoke.sh https://25types.com      # → production

set -uo pipefail

BASE="${1:-http://localhost}"
PASS=0; FAIL=0; SKIP=0
T=$(date +%s)
USER_A="a-$T"
USER_B="b-$T"
PASSWD="test-pass-123"
NEWPASSWD="new-test-pass-123"
TOKEN_A=""; TOKEN_B=""; ANON_TOKEN=""
REVIEW_TOKEN=""; REVIEW_ID=""
MATCH_LINK_ID=""; MATCH_LINK_TOKEN=""; B_USER_ID=""; A_USER_ID=""

RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

TMP=$(mktemp -d /tmp/smoke-XXXXXX)
trap 'rm -rf "$TMP"' EXIT

# --- JSON helpers ---

json_val() {
  if command -v jq &>/dev/null; then
    echo "$1" | jq -r "$2" 2>/dev/null || true
  else
    local k="${2##*.}"
    echo "$1" | grep -o "\"$k\":\"[^\"]*\"" | head -1 | sed "s/\"$k\":\"//;s/\"$//" || true
  fi
}

json_code() {
  if command -v jq &>/dev/null; then
    echo "$1" | jq -r ".error.code // empty" 2>/dev/null || true
  else
    echo "$1" | grep -o '"code":"[^"]*"' | head -1 | sed 's/"code":"//;s/"//' || true
  fi
}

# --- assertion helpers ---

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

check_ok()   { check "$1" "ok" "${2:-}" "${3:-}"; }
check_200()  { check "$1 HTTP" "200" "${2:-}" "${3:-}"; }
check_201()  { check "$1 HTTP" "201" "${2:-}" "${3:-}"; }
check_400()  { check "$1 HTTP" "400" "${2:-}" "${3:-}"; }
check_401()  { check "$1 HTTP" "401" "${2:-}" "${3:-}"; }
check_404()  { check "$1 HTTP" "404" "${2:-}" "${3:-}"; }
check_409()  { check "$1 HTTP" "409" "${2:-}" "${3:-}"; }

# --- curl wrapper ---
# api METHOD PATH [BODY] [TOKEN]
# Writes response body to $TMP/body, prints HTTP status code to stdout.

api() {
  local method="$1" path="$2" data="${3:-}" token="${4:-}"
  local f="$TMP/body"
  if [ -n "$data" ]; then
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$BASE$path" \
      -H 'Content-Type: application/json' \
      ${token:+-H "Authorization: Bearer $token"} \
      -d "$data" || { echo "000"; return; }
  else
    curl -s -w '%{http_code}' -o "$f" -X "$method" "$BASE$path" \
      -H 'Content-Type: application/json' \
      ${token:+-H "Authorization: Bearer $token"} || { echo "000"; return; }
  fi
}

body() { cat "$TMP/body" 2>/dev/null || true; }

# --- page helpers ---

page() {
  local path="$1" token="${2:-}"
  local f="$TMP/page_body"
  if [ -n "$token" ]; then
    curl -s -w '%{http_code}' -o "$f" "$BASE$path" \
      -H "Authorization: Bearer $token" || { echo "000"; return; }
  else
    curl -s -w '%{http_code}' -o "$f" "$BASE$path" || { echo "000"; return; }
  fi
}

page_body() { cat "$TMP/page_body" 2>/dev/null || true; }

check_page() {
  local desc="$1" path="$2" needle="${3:-}" token="${4:-}"
  local s; s=$(page "$path" "$token")
  check "  page $path HTTP" "200" "$s" ""
  if [ -n "$needle" ]; then
    local b; b=$(page_body)
    if echo "$b" | grep -q "$needle"; then
      check "  page $path contains \"$needle\"" "ok" "ok" ""
    else
      check "  page $path contains \"$needle\"" "ok" "missing" ""
    fi
  fi
}

# ============================================================================
echo "${BOLD}25types — Journey Smoke Test${NC}"
echo "Target: $BASE"
echo ""

# ============================================================================
# PHASE 0 — 静态页面 (Caddy file_server, 无替代覆盖)
# ============================================================================
echo "${BOLD}── Phase 0: static pages ──${NC}"

# 0a: 纯静态页面 (9 个, 无需 auth)
for p in /about /faq /privacy /terms /refund /cookies /error_404 /error_500; do
  check_page "$p" "$p"
done
s=$(page "/")
check "  page / HTTP" "302" "$s"

# 0b: 公开功能页面 (8 个)
for p in /assess /types /result /login /register /forgot-password /reset-password /verify-email; do
  check_page "$p" "$p"
done

# 0b2: Mingli 页面
for p in /mingli/bazi /hehun; do
  check_page "$p" "$p"
done

# 0c: 需 auth 页面 (/bonds, /settings 已迁移到 SPA /app#bonds, /app#settings)
for p in /donate; do
  check_page "$p" "$p"
done

# 0c2: SPA entry point (serves /app#bonds, /app#settings, /app#overview, etc.)
check_page "/app" "/app"

# 0d: 动态路径
check_page "/types/WF" "/types/WF"
check_page "/r/test-token" "/r/test-token"

# 0e: 内容校验 — 确保返回正确页面而非被重写到 profile
check_page "/donate" "/donate" "<title>Donate"
check_page "/login" "/login" "<title>Login"
check_page "/about" "/about" "<title>About"
check_page "/app" "/app" "25types"
check_page "/result" "/result" "Share Card"
check_page "/profile" "/profile" "<title>Profile"

echo ""
echo "  (no-locale prefix paths — must serve correct page via try_files fallback)"
check_page "/reset-password" "/reset-password" "<title>Reset Password"
check_page "/verify-email" "/verify-email" "<title>Verify Email"
check_page "/forgot-password" "/forgot-password" "<title>Forgot"

# ============================================================================
# PHASE 1 — 匿名评估旅程
# ============================================================================
echo ""
echo "${BOLD}── Phase 1: anonymous assessment journey ──${NC}"

# Health canary — simplest endpoint, verifies server is alive
s=$(api GET /api/health)
check_200 "GET /api/health" "$s"

# Questions endpoint — used by assess page
s=$(api GET '/api/assessments/questions?locale=en')
check_200 "GET /api/assessments/questions?locale=en" "$s"

# Anonymous assessment — round 1
ANON_TOKEN=$(uuidgen 2>/dev/null || echo "anon-$T-$RANDOM")
ANSWERS='{"answers":[
  {"qid":"Q01","selections":["W","F"]},
  {"qid":"Q02","selections":["W","M"]},
  {"qid":"Q03","selections":["F","E"]},
  {"qid":"Q04","selections":["W","F"]},
  {"qid":"Q05","selections":["E","M"]}
],"anonymous_token":"'"$ANON_TOKEN"'"}'
s=$(api POST /api/assessments "$ANSWERS")
check_201 "POST /api/assessments (anon round 1)" "$s"
b=$(body)
check "  anonymous_token echoed" "$ANON_TOKEN" "$(json_val "$b" '.data.anonymous_token')" "$b"

# Anonymous round 2 — completes assessment
ANSWERS2='{"answers":[
  {"qid":"Q06","selections":["W","F"]},
  {"qid":"Q07","selections":["W","E"]},
  {"qid":"Q08","selections":["F","E"]},
  {"qid":"Q09","selections":["W","E"]},
  {"qid":"Q10","selections":["F","M"]}
],"anonymous_token":"'"$ANON_TOKEN"'"}'
s=$(api POST /api/assessments "$ANSWERS2")
check_201 "POST /api/assessments (anon round 2)" "$s"

# ============================================================================
# PHASE 1b — BaZi (八字) 排盘 & 合婚
# ============================================================================
echo ""
echo "${BOLD}── Phase 1b: Mingli chart & match ──${NC}"

# --- Cities lookup ---
s=$(api GET /api/reference/cities)
check_200 "GET /api/reference/cities (no query)" "$s"
b=$(body)
check "  returns items" "false" "$([ "$(json_val "$b" '.data.total')" -gt 0 ] && echo false || echo true)"

s=$(api GET '/api/reference/cities?q=bei')
check_200 "GET /api/reference/cities?q=bei" "$s"
b=$(body)
check "  bei matches Beijing" "false" "$(echo "$b" | grep -q '"Beijing"' && echo false || echo true)"

s=$(api GET '/api/reference/cities?q=上海')
check_200 "GET /api/reference/cities?q=上海" "$s"

# --- Chart: valid request (Beijing 2024-02-04 12:00, known case) ---
CHART='{"year":2024,"month":2,"day":4,"hour":12,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}'
s=$(api POST /api/bazi/chart "$CHART")
check_200 "POST /api/bazi/chart (Beijing 2024-02-04)" "$s"
b=$(body)
check "  day_master present" "false" "$([ -z "$(json_val "$b" '.data.day_master')" ] && echo true || echo false)"
check "  has year_pillar" "false" "$([ -z "$(json_val "$b" '.data.year_pillar.stem')" ] && echo true || echo false)"
check "  has na_yin" "false" "$([ -n "$(echo "$b" | grep -o 'na_yin')" ] && echo false || echo true)"
check "  has element_count" "false" "$([ -z "$(json_val "$b" '.data.element_count."1"')" ] && echo true || echo false)"
check "  has life_stages" "false" "$([ -n "$(echo "$b" | grep -o 'life_stages')" ] && echo false || echo true)"
check "  has big_fortune" "false" "$([ -z "$(json_val "$b" '.data.big_fortune.start_age')" ] && echo true || echo false)"

# --- Chart: defaults (longitude/timezone = 0 → auto 120) ---
CHART_MIN='{"year":2000,"month":6,"day":15,"hour":8,"minute":30}'
s=$(api POST /api/bazi/chart "$CHART_MIN")
check_200 "POST /api/bazi/chart (minimal fields)" "$s"

# --- Chart: with DST (1986-07-15 Beijing, DST active in China 1986-1991) ---
CHART_DST='{"year":1986,"month":7,"day":15,"hour":12,"minute":0,"longitude":116.4,"timezone":8,"is_dst":true}'
s=$(api POST /api/bazi/chart "$CHART_DST")
check_200 "POST /api/bazi/chart (DST enabled)" "$s"

# --- Chart: validation errors ---
s=$(api POST /api/bazi/chart '{"year":1800,"month":6,"day":15,"hour":12,"minute":0}')
check_400 "POST /api/bazi/chart (year<1900)" "$s"

s=$(api POST /api/bazi/chart '{"year":2020,"month":13,"day":1,"hour":12,"minute":0}')
check_400 "POST /api/bazi/chart (month>12)" "$s"

s=$(api POST /api/bazi/chart '{"year":2020,"month":6,"day":15,"hour":25,"minute":0}')
check_400 "POST /api/bazi/chart (hour>23)" "$s"

s=$(api POST /api/bazi/chart '{"year":2020,"month":6,"day":15,"hour":12,"minute":0,"gender":"other"}')
check_400 "POST /api/bazi/chart (invalid gender)" "$s"

s=$(api POST /api/bazi/chart 'not-json')
check_400 "POST /api/bazi/chart (malformed JSON)" "$s"

s=$(api POST /api/bazi/chart '{"year":2020,"month":6,"day":15,"hour":12,"minute":0,"longitude":200}')
check_400 "POST /api/bazi/chart (longitude>180)" "$s"

# --- Match: valid request ---
MATCH='{"a":{"year":1984,"month":2,"day":4,"hour":18,"minute":0,"gender":"male"},"b":{"year":1990,"month":5,"day":20,"hour":8,"minute":0,"gender":"female"}}'
s=$(api POST /api/bazi/match "$MATCH")
check_200 "POST /api/bazi/match" "$s"
b=$(body)
check "  has total score" "false" "$([ -z "$(json_val "$b" '.data.total')" ] && echo true || echo false)"
check "  has level" "false" "$([ -z "$(json_val "$b" '.data.level')" ] && echo true || echo false)"
check "  has details" "false" "$([ -n "$(echo "$b" | grep -o 'details')" ] && echo false || echo true)"
check "  has chart_a" "false" "$([ -z "$(json_val "$b" '.data.chart_a.day_master')" ] && echo true || echo false)"
check "  has chart_b" "false" "$([ -z "$(json_val "$b" '.data.chart_b.day_master')" ] && echo true || echo false)"

# --- Match: validation ---
s=$(api POST /api/bazi/match '{"a":{"year":2000,"month":6,"day":15,"hour":12,"minute":0}}')
check_400 "POST /api/bazi/match (missing b)" "$s"

s=$(api POST /api/bazi/match '{}')
check_400 "POST /api/bazi/match (empty body)" "$s"

# --- Mingli match links ---
s=$(api POST /api/mingli-match-links '' "$TOKEN_A")
check_201 "POST /api/mingli-match-links" "$s"
MINGLI_LINK_TOKEN=$(json_val "$(body)" '.data.token')
check "  has token" "false" "$([ -z "$MINGLI_LINK_TOKEN" ] && echo true || echo false)"

MINGLI_LINK_URL=$(json_val "$(body)" '.data.url')
check "  url starts with /ml/" "false" "$(echo "$MINGLI_LINK_URL" | grep -q '^/ml/' && echo false || echo true)"

s=$(api GET /api/mingli-match-links "" "$TOKEN_A")
check_200 "GET /api/mingli-match-links" "$s"

# Public landing: get link info
s=$(api GET "/api/ml/$MINGLI_LINK_TOKEN")
check_200 "GET /api/ml/{token}" "$s"
check "  has chart_a" "false" "$([ -z "$(echo "$(body)" | grep -o 'chart_a')" ] && echo true || echo false)"

# Submit with birth_info (anonymous)
s=$(api POST "/api/ml/$MINGLI_LINK_TOKEN" '{"birth_info":{"year":1990,"month":6,"day":15,"hour":12,"minute":0,"longitude":120.0,"timezone":8.0,"gender":"female"}}')
check_201 "POST /api/ml/{token} (birth_info)" "$s"
check "  has total score" "false" "$([ -z "$(echo "$(body)" | grep -o '"total"')" ] && echo true || echo false)"

# Submit with use_existing (authenticated user with birth_info)
s=$(api POST "/api/ml/$MINGLI_LINK_TOKEN" '{"use_existing":true}' "$TOKEN_A")
check_201 "POST /api/ml/{token} (use_existing)" "$s"

# List links again to verify match_count
s=$(api GET /api/mingli-match-links "" "$TOKEN_A")
check "  match_count > 0" "false" "$(echo "$(body)" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['data']['items'][0]['match_count'] > 0" 2>/dev/null && echo false || echo true)"

# Delete link
MINGLI_LINK_ID=$(echo "$(body)" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['items'][0]['id'])" 2>/dev/null)
s=$(api DELETE "/api/mingli-match-links/$MINGLI_LINK_ID" "" "$TOKEN_A")
check_200 "DELETE /api/mingli-match-links/{id}" "$s"

# Verify link is gone
s=$(api GET "/api/ml/$MINGLI_LINK_TOKEN")
check_404 "GET /api/ml/{token} (deleted)" "$s"

# --- Profile with Mingli: save birth_info via PATCH ---
s=$(api PATCH /api/users/me '{"birth_info":{"year":1995,"month":8,"day":21,"hour":14,"minute":30,"longitude":116.4,"timezone":8,"gender":"male"}}' "$TOKEN_A")
check_200 "PATCH /api/users/me (with birth_info)" "$s"
b=$(body)
check "  birth_info saved" "false" "$([ -z "$(json_val "$b" '.data.birth_info.year')" ] && echo true || echo false)"

# Verify BaZi chart appears in profile after saving birth_info
GET_PROFILE=$(api GET "/api/profiles/$USER_A" "" "$TOKEN_A")
check_200 "GET /api/profiles (after birth_info save)" "$GET_PROFILE"
check "  profile has bazi_chart" "false" "$([ -n "$(echo "$GET_PROFILE" | grep -o 'bazi_chart')" ] && echo false || echo true)"

# ============================================================================
# PHASE 2 — 认证旅程
# ============================================================================
echo ""
echo "${BOLD}── Phase 2: auth journey ──${NC}"

# Register A (with anonymous token to auto-claim)
REG_A='{"name":"'"$USER_A"'","email":"suiqiang+smoke-'"$USER_A"'@foxmail.com","password":"'"$PASSWD"'","anonymous_token":"'"$ANON_TOKEN"'"}'
s=$(api POST /api/auth/register "$REG_A")
check_201 "POST /api/auth/register (A + claim)" "$s"
TOKEN_A=$(json_val "$(body)" '.data.token')
check "  token present" "false" "$([ -z "$TOKEN_A" ] && echo true || echo false)" ""

# Login A
LOGIN='{"name":"'"$USER_A"'","password":"'"$PASSWD"'"}'
s=$(api POST /api/auth/login "$LOGIN")
check_200 "POST /api/auth/login (A)" "$s"
TOKEN_A=$(json_val "$(body)" '.data.token')

# GetMe — verify identity
s=$(api GET /api/users/me "" "$TOKEN_A")
check_200 "GET /api/users/me" "$s"
check "  name matches" "$USER_A" "$(json_val "$(body)" '.data.name')"

# UpdateMe — rename + public
s=$(api PATCH /api/users/me '{"name":"a-renamed-'"$T"'","is_public":true}' "$TOKEN_A")
check_200 "PATCH /api/users/me (rename+public)" "$s"
USER_A="a-renamed-$T"

# Change password
CHPWD='{"current_password":"'"$PASSWD"'","new_password":"new-'"$PASSWD"'"}'
s=$(api PUT /api/auth/password "$CHPWD" "$TOKEN_A")
check_200 "PUT /api/auth/password" "$s"
TOKEN_A=$(json_val "$(body)" '.data.token')
LOGIN='{"name":"'"$USER_A"'","password":"new-'"$PASSWD"'"}'

# Authenticated assessment — round 3 (completes 30 questions)
ANSWERS3='{"answers":[
  {"qid":"Q11","selections":["W","E"]},
  {"qid":"Q12","selections":["F","E"]},
  {"qid":"Q13","selections":["W","F"]},
  {"qid":"Q14","selections":["W","F"]},
  {"qid":"Q15","selections":["E","M"]}
]}'
s=$(api POST /api/assessments "$ANSWERS3" "$TOKEN_A")
check_201 "POST /api/assessments (auth)" "$s"

# Logout
s=$(api POST /api/auth/logout "" "$TOKEN_A")
check_200 "POST /api/auth/logout" "$s"

# Verify post-logout — old token rejected
s=$(api GET /api/users/me "" "$TOKEN_A")
check_401 "GET /api/users/me (after logout)" "$s"

# Re-login to continue
s=$(api POST /api/auth/login "$LOGIN")
check_200 "POST /api/auth/login (re-login)" "$s"
TOKEN_A=$(json_val "$(body)" '.data.token')

# Get A's user ID
s=$(api GET /api/users/me "" "$TOKEN_A")
A_USER_ID=$(json_val "$(body)" '.data.id')

# ============================================================================
# PHASE 3 — 他评旅程 (peer review)
# ============================================================================
echo ""
echo "${BOLD}── Phase 3: peer review journey ──${NC}"

# Register B
REG_B='{"name":"'"$USER_B"'","email":"suiqiang+smoke-'"$USER_B"'@foxmail.com","password":"'"$PASSWD"'"}'
s=$(api POST /api/auth/register "$REG_B")
check_201 "POST /api/auth/register (B)" "$s"
TOKEN_B=$(json_val "$(body)" '.data.token')

# B does assessment
ANSWERS_B='{"answers":[
  {"qid":"Q01","selections":["F","W"]},
  {"qid":"Q02","selections":["E","F"]},
  {"qid":"Q03","selections":["M","E"]},
  {"qid":"Q04","selections":["R","M"]},
  {"qid":"Q05","selections":["W","R"]}
]}'
s=$(api POST /api/assessments "$ANSWERS_B" "$TOKEN_B")
check_201 "POST /api/assessments (B)" "$s"

# Get B's user ID
s=$(api GET /api/users/me "" "$TOKEN_B")
B_USER_ID=$(json_val "$(body)" '.data.id')

# B sets public
s=$(api PATCH /api/users/me '{"is_public":true}' "$TOKEN_B")
check_200 "PATCH /api/users/me (B public)" "$s"

# A views B's public profile
s=$(api GET "/api/profiles/$USER_B" "" "$TOKEN_A")
check_200 "GET /api/profiles/{B_name} (A views B)" "$s"

# A creates review link
s=$(api POST /api/reviews "" "$TOKEN_A")
check_201 "POST /api/reviews" "$s"
REVIEW_ID=$(json_val "$(body)" '.data.id')
REVIEW_TOKEN=$(json_val "$(body)" '.data.token')
check "  has token" "false" "$([ -z "$REVIEW_TOKEN" ] && echo true || echo false)" ""

# Open review link (public)
s=$(api GET "/api/r/$REVIEW_TOKEN")
check_200 "GET /api/r/{token} (valid)" "$s"

# Anonymous peer submit — round 1
PEER1='{"reviewer_name":"peer-bob","anonymous_token":"peer-'"$T"'","answers":[
  {"qid":"Q01","selections":["M","R"]},
  {"qid":"Q02","selections":["R","W"]},
  {"qid":"Q03","selections":["W","F"]},
  {"qid":"Q04","selections":["F","E"]},
  {"qid":"Q05","selections":["E","M"]}
]}'
s=$(api POST "/api/r/$REVIEW_TOKEN" "$PEER1")
check_201 "POST /api/r/{token} (peer round 1)" "$s"

# Anonymous peer — round 2
PEER2='{"reviewer_name":"peer-bob","anonymous_token":"peer-'"$T"'","answers":[
  {"qid":"Q06","selections":["W","F"]},
  {"qid":"Q07","selections":["F","E"]},
  {"qid":"Q08","selections":["E","M"]},
  {"qid":"Q09","selections":["M","R"]},
  {"qid":"Q10","selections":["R","W"]}
]}'
s=$(api POST "/api/r/$REVIEW_TOKEN" "$PEER2")
check_201 "POST /api/r/{token} (peer round 2)" "$s"

# A views peer aggregation
s=$(api GET /api/assessments/peers "" "$TOKEN_A")
check_200 "GET /api/assessments/peers" "$s"

# A deletes review link
s=$(api DELETE "/api/reviews/$REVIEW_ID" "" "$TOKEN_A")
check_200 "DELETE /api/reviews/{id}" "$s"

# Deleted link → 404
s=$(api GET "/api/r/$REVIEW_TOKEN")
check_404 "GET /api/r/{token} (deleted)" "$s"

# ============================================================================
# PHASE 4 — Match & bond 旅程
# ============================================================================
echo ""
echo "${BOLD}── Phase 4: match & bond journey ──${NC}"

# A creates match link
s=$(api POST /api/match-links "" "$TOKEN_A")
check_201 "POST /api/match-links (A create)" "$s"
MATCH_LINK_ID=$(json_val "$(body)" '.data.id')
MATCH_LINK_TOKEN=$(json_val "$(body)" '.data.token')
check "  has id" "false" "$([ -z "$MATCH_LINK_ID" ] && echo true || echo false)"

# Get match link info (public)
s=$(api GET "/api/m/$MATCH_LINK_TOKEN")
check_200 "GET /api/m/{token} (link info)" "$s"

# B (logged in) submits via match link — use_existing
s=$(api POST "/api/m/$MATCH_LINK_TOKEN" '{"use_existing":true}' "$TOKEN_B")
check_201 "POST /api/m/{token} (B use_existing)" "$s"
b=$(body)
check "  has bond" "true" "$([ -n "$(json_val "$b" '.data.bond')" ] && echo true || echo false)" "$b"

# Get bonds for A
s=$(api GET "/api/profiles/$USER_A/bonds" "" "$TOKEN_A")
check_200 "GET /api/profiles/A/bonds" "$s"

# Get bonds for B (perspective swap)
s=$(api GET "/api/profiles/$USER_B/bonds" "" "$TOKEN_B")
check_200 "GET /api/profiles/B/bonds" "$s"

# A × B instant bond
s=$(api POST /api/bond '{"with_user_id":'"$B_USER_ID"'}' "$TOKEN_A")
check_200 "POST /api/bond (A×B instant)" "$s"

# A deletes match link
s=$(api DELETE "/api/match-links/$MATCH_LINK_ID" "" "$TOKEN_A")
check_200 "DELETE /api/match-links/{id}" "$s"

# Deleted link → 404
s=$(api GET "/api/m/$MATCH_LINK_TOKEN")
check_404 "GET /api/m/{token} (deleted)" "$s"

# ============================================================================
# PHASE 5 — 隐私流程
# ============================================================================
echo ""
echo "${BOLD}── Phase 5: privacy flow ──${NC}"

# B sets private
s=$(api PATCH /api/users/me '{"is_public":false}' "$TOKEN_B")
check_200 "PATCH /api/users/me (B private)" "$s"

# A tries to view B → 404 (private profiles not shown)
s=$(api GET "/api/profiles/$USER_B" "" "$TOKEN_A")
check_404 "GET /api/profiles/{B_name} (private)" "$s"

# B logout
s=$(api POST /api/auth/logout "" "$TOKEN_B")
check_200 "POST /api/auth/logout (B)" "$s"

# ============================================================================
# PHASE 6 — 删号旅程 (放最后，不影响其他)
# ============================================================================
echo ""
echo "${BOLD}── Phase 6: delete account ──${NC}"

s=$(api DELETE /api/users/me)
check_401 "DELETE /api/users/me (no token)" "$s"

s=$(api DELETE /api/users/me "" "$TOKEN_A")
check_200 "DELETE /api/users/me" "$s"
b=$(body)
REACTIVATE=$(json_val "$b" '.data.reactivate_by')
check "  has reactivate_by" "false" "$([ -z "$REACTIVATE" ] && echo true || echo false)"

s=$(api GET /api/users/me "" "$TOKEN_A")
check_401 "GET /api/users/me (after delete)" "$s"

# ============================================================================
# PHASE 7 — i18n + 错误页
# ============================================================================
echo ""
echo "${BOLD}── Phase 7: i18n & error pages ──${NC}"

# API language switching
EN_BODY=$(curl -s "$BASE/api/assessments/questions?locale=en" || true)
ZH_BODY=$(curl -s "$BASE/api/assessments/questions?locale=zh-CN" || true)
if [ "$EN_BODY" != "$ZH_BODY" ]; then
  check "  X-Locale en/zh-CN differ" "ok" "ok"
else
  check "  X-Locale en/zh-CN differ" "ok" "same" "(may be same if both return R01)"
fi

# Error page E2E (Caddy serves error_404.html for unknown paths)
s=$(page "/this-page-does-not-exist-$T")
b=$(page_body)
if echo "$b" | grep -qi "404\|not found\|page not found"; then
  check "  404 page has error content" "ok" "ok"
else
  check "  404 page has error content" "ok" "missing" "(HTTP $s)"
fi

# ============================================================================
# PHASE 8 — 限流 (仅生产环境 — Caddy rate limit)
# ============================================================================
echo ""
echo "${BOLD}── Phase 8: rate limiting ──${NC}"

if echo "$BASE" | grep -q "localhost\|127.0.0.1\|::1"; then
  echo -e "  \xe2\x9c\x93 rate limiting skipped (localhost — Caddy handles rate limits in production)"
  PASS=$((PASS + 1))
else
  BURST=0
  for i in $(seq 1 25); do
    sc=$(api POST /api/auth/login '{"name":"nonexistent-'"$T"'","password":"wrongpass123"}' | tail -1)
    if [ "$sc" = "429" ]; then
      BURST=1
      break
    fi
  done
  check "  rate limit 429 triggered" "1" "$BURST" "after 25 rapid logins"
fi

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "${BOLD}────────────────────────────────────────${NC}"
TOTAL=$((PASS + FAIL + SKIP))
echo "Total: $TOTAL  ${GREEN}PASS: $PASS${NC}  ${RED}FAIL: $FAIL${NC}  SKIP: $SKIP"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo "${RED}Some tests failed.${NC}"
  echo "Re-run a failing step:"
  echo "  curl -v $BASE/api/health"
  exit 1
else
  echo "${GREEN}All journey smoke tests passed.${NC}"
  exit 0
fi
