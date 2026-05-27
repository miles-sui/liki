#!/usr/bin/env bash
# Start development environment using Docker Compose.
# Same Caddyfile + same Docker images as production.
set -e
cd "$(dirname "$0")/.."

# Build frontend if needed
if [ ! -f web/dist/manifest.json ]; then
  echo "==> Building frontend..."
  (cd web && npm run build)
fi

# Clean SQLite WAL/SHM in case they were created by a different user
rm -f data/25types.db-wal data/25types.db-shm

echo "==> Starting development environment (Caddy :8080 + API :8081)"
docker compose -f deploy/app/docker-compose.yml up --build
