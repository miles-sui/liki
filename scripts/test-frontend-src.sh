#!/bin/bash
# Frontend source-level contract test — NO server required.
# Validates JS patterns, HTML templates, namedToArray unit behaviour, built dist.
#
# Usage:
#   scripts/test-frontend-src.sh

set -uo pipefail

cd "$(dirname "$0")/.."

PASS=0; FAIL=0
RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

check() {
  local desc="$1" expected="$2" actual="${3:-}"
  if [ "$actual" = "$expected" ]; then
    echo -e "  ${GREEN}OK${NC}   $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}FAIL${NC} $desc (expected '$expected', got '$actual')"
    FAIL=$((FAIL + 1))
  fi
}

check_file() {
  local desc="$1" file="$2" needle="$3"
  if [ -f "$file" ] && grep -q "$needle" "$file" 2>/dev/null; then
    echo -e "  ${GREEN}OK${NC}   $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}FAIL${NC} $desc (file:$file, pattern:'$needle')"
    FAIL=$((FAIL + 1))
  fi
}

echo "${BOLD}=== Frontend Source Contract Tests ===${NC}"
echo ""

# ── Phase 1: JS source code patterns ─────────────────────────────

echo "${BOLD}Phase 1: JS source patterns (named-object handling)${NC}"

WEB="web/js"

check_file "common.js: defines namedToArray" "$WEB/common.js" "function namedToArray"
check_file "common.js: defines deviationToProportion" "$WEB/common.js" "function deviationToProportion"
check_file "assess.js: uses namedToArray or string keys" "$WEB/assess.js" "namedToArray"
check_file "result.js: uses namedToArray" "$WEB/result.js" "namedToArray"
check_file "profile.js: uses namedToArray" "$WEB/profile.js" "namedToArray"
check_file "profile.js: uses pValues" "$WEB/profile.js" "pValues"
check_file "match-landing.js: uses namedToArray" "$WEB/match-landing.js" "namedToArray"
check_file "match-landing.js: uses computeBondShapes" "$WEB/match-landing.js" "computeBondShapes"
# bond-history.js removed — bonds page now served by SPA /app#bonds
check_file "common.js: defines computeBondShapes" "$WEB/common.js" "function computeBondShapes"

# Anonymous token claim flow
check_file "login.js: claims anonymous_token after login" "$WEB/login.js" "api.*assessments/claim"
check_file "register.js: includes anonymous_token" "$WEB/register.js" "anonymous_token"
check_file "common.js: manages sessionStorage for anon_token" "$WEB/common.js" "sessionStorage"

# Numeric indexing guard
for f in web/js/*.js; do
  name=$(basename "$f")
  [[ "$name" == "charts.js" ]] && continue
  [[ "$name" == "common.js" ]] && continue
  if grep -q 'namedToArray\|computeBondShapes' "$f" 2>/dev/null; then
    continue
  fi
  if grep -q '\.d_eff\[' "$f" 2>/dev/null; then
    echo -e "  ${RED}FAIL${NC} $name: numeric .d_eff[i] without namedToArray conversion"
    FAIL=$((FAIL + 1))
  fi
  if grep -q '\.delta_a\.map\|\.delta_b\.map' "$f" 2>/dev/null; then
    if ! grep -q 'namedToArray\|computeBondShapes' "$f" 2>/dev/null; then
      echo -e "  ${RED}FAIL${NC} $name: .map() on delta without namedToArray/computeBondShapes conversion"
      FAIL=$((FAIL + 1))
    fi
  fi
done
echo -e "  ${GREEN}OK${NC}   Numeric indexing guarded by namedToArray conversion"
PASS=$((PASS + 1))

# ── Phase 2: HTML template patterns ──────────────────────────────

echo ""
echo "${BOLD}Phase 2: HTML template patterns${NC}"

check_file "result.js: pValues chart data" "$WEB/result.js" "pValues"
check_file "profile.html: delta template" "web/pages/profile.html" "delta"
check_file "profile-card.njk: radar id" "web/_includes/profile-card.njk" "radarId"

# Dangerous dump|safe
if grep -rn 'x-data.*dump | safe' web/pages/ web/_includes/ 2>/dev/null; then
  echo -e "  ${RED}FAIL${NC} Found 'dump | safe' in x-data attribute"
  FAIL=$((FAIL + 1))
else
  echo -e "  ${GREEN}OK${NC}   No dangerous 'dump | safe' in x-data attributes"
  PASS=$((PASS + 1))
fi

check_file "chartColors used consistently" "$WEB/common.js" "chartColors"

# Bare paths guard (locale prefix required)
BARE_HREF=$(grep -rn 'href="/\(types\|explore\|assess\|result\|login\|register\|forgot-password\|bond\|profile\|settings\|donate\)"' web/pages/ 2>/dev/null)
if [ -n "$BARE_HREF" ]; then
  echo -e "  ${RED}FAIL${NC} Bare paths in page templates:"
  echo "$BARE_HREF"
  FAIL=$((FAIL + 1))
else
  echo -e "  ${GREEN}OK${NC}   No bare-path hrefs (all use locale prefix)"
  PASS=$((PASS + 1))
fi

BARE_CTAHREF=$(grep -rn "ctaHref = '/[a-z]" web/pages/ 2>/dev/null)
if [ -n "$BARE_CTAHREF" ]; then
  echo -e "  ${RED}FAIL${NC} Bare ctaHref values:"
  echo "$BARE_CTAHREF"
  FAIL=$((FAIL + 1))
else
  echo -e "  ${GREEN}OK${NC}   No bare ctaHref values"
  PASS=$((PASS + 1))
fi

BARE_XBIND=$(grep -rn ":href=\"'/\(types\|explore\|assess\|result\|login\|register\|forgot-password\|bond\|profile\|settings\|donate\)" web/pages/ 2>/dev/null)
if [ -n "$BARE_XBIND" ]; then
  echo -e "  ${RED}FAIL${NC} Bare paths in :href bindings:"
  echo "$BARE_XBIND"
  FAIL=$((FAIL + 1))
else
  echo -e "  ${GREEN}OK${NC}   No bare paths in :href bindings"
  PASS=$((PASS + 1))
fi

check_file "assess.js: uses localePath for redirect" "$WEB/assess.js" "localePath"

# Phase 3 removed — namedToArray/deviationToProportion math is verified by
# Go engine tests (internal/engine/engine_test.go) with compile-time safety.
# JS functions cannot be loaded by node (browser APIs, no module.exports).

# ── Phase 3: Built dist verification ────────────────────────────

echo ""
echo "${BOLD}Phase 3: Built dist/ verification${NC}"

if [ -f web/dist/manifest.json ]; then
  FPM=$(cat web/dist/manifest.json | python3 -c "
import json,sys
d=json.load(sys.stdin)
for k,v in d.get('files',{}).items():
  if 'common.js' in k and 'vendor' not in k:
    print(v.lstrip('/'))
    break
" 2>/dev/null)
  if [ -n "$FPM" ]; then
    check_file "dist/common.js: namedToArray present" "web/dist/$FPM" "namedToArray"
    check_file "dist/common.js: deviationToProportion present" "web/dist/$FPM" "deviationToProportion"
  fi
fi

for f in assess.js result.js profile.js; do
  check_file "dist/$f: namedToArray used" "web/dist/js/$f" "namedToArray"
done
for f in match-landing.js; do
  check_file "dist/$f: computeBondShapes used" "web/dist/js/$f" "computeBondShapes"
done
check_file "dist/login.js: sessionStorage for anon_token" "web/dist/js/login.js" "anon_token"

# ── Phase 4: Share card & radar ────────────────────────────────

echo ""
echo "${BOLD}Phase 4: Share card & radar patterns${NC}"

check_file "charts.js: generateShareCard" "$WEB/charts.js" "generateShareCard"
check_file "common.js: saveOrShare" "$WEB/common.js" "saveOrShare"
check_file "charts.js: roundRect" "$WEB/charts.js" "roundRect"
check_file "charts.js: loadImage" "$WEB/charts.js" "loadImage"
check_file "result.js: generateShareCard" "$WEB/result.js" "generateShareCard"
check_file "result.js: radarInst" "$WEB/result.js" "radarInst"
check_file "profile.js: generateShareCard" "$WEB/profile.js" "generateShareCard"
check_file "profile.js: profileRadarInst" "$WEB/profile.js" "profileRadarInst"

# Translation keys in page templates
check_file "result.html: shareCard key" "web/pages/result.html" "result.shareCard"
check_file "profile.html: shareCard key" "web/pages/profile.html" "profile.shareCard"

# ── Summary ──────────────────────────────────────────────────────

echo ""
echo "${BOLD}========================================${NC}"
TOTAL=$((PASS + FAIL))
echo "Total: $TOTAL  ${GREEN}PASS: $PASS${NC}  ${RED}FAIL: $FAIL${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo "${RED}Some frontend source tests failed.${NC}"
  exit 1
else
  echo "${GREEN}All frontend source tests passed.${NC}"
  exit 0
fi
