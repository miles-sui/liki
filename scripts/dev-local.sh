#!/usr/bin/env bash
# Start development environment WITHOUT Docker.
# Same Caddyfile + same routing as production.
# Go API on :8081, Caddy static files + reverse proxy on :8080.
# Requires: go, node
set -euo pipefail
cd "$(dirname "$0")/.."

export JWT_SECRET="${JWT_SECRET:-dev-secret-change-me}"
export LISTEN_ADDR=":8081"

# ---- Ensure caddy ----
if [ -z "${CADDY_BIN:-}" ]; then
  if command -v caddy &>/dev/null; then
    CADDY_BIN="caddy"
  elif [ -f /tmp/caddy ]; then
    CADDY_BIN="/tmp/caddy"
  else
    echo "==> Downloading Caddy..."
    arch=$(uname -m)
    case "$arch" in
      x86_64)  caddy_arch="amd64" ;;
      aarch64) caddy_arch="arm64" ;;
      *)       echo "ERROR: unsupported arch $arch"; exit 1 ;;
    esac
    curl -sL "https://caddyserver.com/api/download?os=linux&arch=${caddy_arch}" -o /tmp/caddy
    chmod +x /tmp/caddy
    CADDY_BIN="/tmp/caddy"
  fi
fi

# ---- Build frontend + API server in parallel ----
echo "==> Building frontend..."
(cd web && npm run build) &
FE_PID=$!

echo "==> Building API server..."
BIN="/tmp/25types-server"
go build -ldflags="-s -w" -o "$BIN" ./cmd/app-server/

wait $FE_PID

# Clean stale WAL
rm -f data/25types.db-wal data/25types.db-shm

# Free ports
fuser -k 8081/tcp 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
sleep 1

echo "==> Starting API server on $LISTEN_ADDR"
"$BIN" &
API_PID=$!
sleep 1

echo "==> Starting Caddy on :8080"
"$CADDY_BIN" run --config deploy/caddy/Caddyfile.local &
CADDY_PID=$!

trap "kill $API_PID $CADDY_PID 2>/dev/null; exit" INT TERM EXIT

echo "==> Ready: http://localhost:8080"

# Seed test user with composite-type (WF) assessment.
ANON="seed-$(date +%s%N)"
ANSWERS='{"answers":[
  {"qid":"Q01","selections":["W","F"]},
  {"qid":"Q02","selections":["W","F"]},
  {"qid":"Q03","selections":["W","F"]},
  {"qid":"Q04","selections":["W","F"]},
  {"qid":"Q05","selections":["W","F"]},
  {"qid":"Q06","selections":["W","F"]},
  {"qid":"Q07","selections":["W","E"]},
  {"qid":"Q08","selections":["W","E"]},
  {"qid":"Q09","selections":["W","M"]},
  {"qid":"Q10","selections":["F","E"]}
],"anonymous_token":"'"$ANON"'"}'
SEED_REG='{"name":"miles","email":"suiqiang@foxmail.com","password":"test1234","anonymous_token":"'"$ANON"'"}'

echo "==> Seeding test user (WF composite)..."
# Seed failures are non-fatal; don't let set -e kill the dev server.
set +e
RESP=$(curl -s -X POST http://localhost:8081/api/assessments -H 'Content-Type: application/json' -d "$ANSWERS")
IDENTITY=$(echo "$RESP" | grep -o '"identity":{"label":"[^"]*","id":"[^"]*"' | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
curl -sf -X POST http://localhost:8081/api/auth/register -H 'Content-Type: application/json' -d "$SEED_REG" > /dev/null 2>&1
SEED_OK=$?
set -e
echo "    test user: miles / test1234 / suiqiang@foxmail.com"
echo "    identity: ${IDENTITY:-unknown}"
echo "    Press Ctrl+C to stop"
wait
