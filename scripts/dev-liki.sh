#!/usr/bin/env bash
# Start Liki development environment WITHOUT Docker.
# Go API on :8081, Caddy static files + reverse proxy on :8080.
set -euo pipefail
cd "$(dirname "$0")/.."

# Load .env if present
if [ -f .env ]; then
  set -a; source .env; set +a
fi

export DEEPSEEK_API_KEY="${DEEPSEEK_API_KEY:-}"
export DODO_API_KEY="${DODO_API_KEY:-}"
export DODO_WEBHOOK_KEY="${DODO_WEBHOOK_KEY:-}"
export DODO_TEST_MODE="${DODO_TEST_MODE:-true}"
export XUNHU_APPID="${XUNHU_APPID:-}"
export XUNHU_APPSECRET="${XUNHU_APPSECRET:-}"
export RESEND_API_KEY="${RESEND_API_KEY:-}"
export RESEND_FROM="${RESEND_FROM:-noreply@localhost}"
export RETURN_URL="${RETURN_URL:-http://localhost:8080}"
export DB_PATH="${DB_PATH:-$(pwd)/data/liki.db}"
export LISTEN_ADDR="${LISTEN_ADDR:-:8081}"

# Caddy binary
if [ -z "${CADDY_BIN:-}" ]; then
  if command -v caddy &>/dev/null; then
    CADDY_BIN="caddy"
  elif [ -f /tmp/caddy ]; then
    chmod +x /tmp/caddy
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
    curl -sL "https://caddyserver.com/api/download?os=linux&arch=${caddy_arch}&format=checksum" -o /tmp/caddy.sha256
    expected=$(awk '{print $1}' /tmp/caddy.sha256)
    actual=$(sha256sum /tmp/caddy | awk '{print $1}')
    if [ "$expected" != "$actual" ]; then
      echo "ERROR: Caddy checksum verification failed"
      exit 1
    fi
    rm -f /tmp/caddy.sha256
    chmod +x /tmp/caddy
    CADDY_BIN="/tmp/caddy"
  fi
fi

echo "==> Building Liki server..."
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
node web/scripts/compile-vue-template.cjs
BIN="/tmp/liki"
go build -ldflags="-s -w -X 'main.BuildTime=$BUILD_TIME'" -o "$BIN" ./cmd/liki/

# Replace BUILD_TIME_PLACEHOLDER in web files for dev display
for f in web/*.html; do
  sed -i "s/BUILD_TIME_PLACEHOLDER/$BUILD_TIME/g" "$f"
done

# Free ports (warn if something else is using them)
for port in 8081 8080; do
  if fuser "$port/tcp" 2>/dev/null; then
    echo "==> Warning: killing existing process on port $port"
    fuser -k "$port/tcp" 2>/dev/null || true
  fi
done
sleep 1


echo "==> Starting API server on $LISTEN_ADDR"
"$BIN" &
API_PID=$!
sleep 1

echo "==> Starting Caddy on :8080"
"$CADDY_BIN" run --config deploy/liki/Caddyfile.local --adapter caddyfile &
CADDY_PID=$!

trap "for f in web/*.html; do sed -i \"s|build: $BUILD_TIME|build: BUILD_TIME_PLACEHOLDER|g\" \"\$f\" || true; done; kill \$API_PID \$CADDY_PID 2>/dev/null; exit" INT TERM EXIT

echo "==> Ready: http://localhost:8080"
echo "    Press Ctrl+C to stop"
wait
