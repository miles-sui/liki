#!/bin/bash
# backup.sh — SQLite WAL-mode backup using VACUUM INTO.
# Safe to run while the server is live (VACUUM INTO is a read-only operation
# that copies the database into a new file with WAL checkpointed in).
#
# Usage:
#   scripts/backup.sh                    # keep 30 days, output to data/backups/
#   scripts/backup.sh /mnt/backups 7     # custom dir, 7-day retention
#
# Cron (daily at 03:07 UTC):
#   7 3 * * * cd /app && scripts/backup.sh >> /var/log/backup.log 2>&1

set -euo pipefail

BACKUP_DIR="${1:-data/backups}"
RETENTION_DAYS="${2:-30}"
DB_PATH="${DB_PATH:-data/lingji.db}"
DB_NAME="$(basename "$DB_PATH" .db)"

mkdir -p "$BACKUP_DIR"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}-${TIMESTAMP}.db"

# Sanity-check: path must not contain single-quotes (would break VACUUM INTO).
case "$BACKUP_FILE" in *"'"*) echo "[backup] ERROR: path contains single-quote"; exit 1 ;; esac

echo "[backup] $(date -u +%Y-%m-%dT%H:%M:%SZ) starting VACUUM INTO $BACKUP_FILE"

if ! command -v sqlite3 &>/dev/null; then
    echo "[backup] sqlite3 CLI not found — install with: apt install sqlite3"
    exit 1
fi
sqlite3 "$DB_PATH" "VACUUM INTO '$BACKUP_FILE'"
echo "[backup] $(date -u +%Y-%m-%dT%H:%M:%SZ) completed ($(du -h "$BACKUP_FILE" | cut -f1))"

# Prune old backups.
DELETED=$(find "$BACKUP_DIR" -name "${DB_NAME}-*.db" -mtime "+${RETENTION_DAYS}" -delete -print | wc -l)
echo "[backup] $(date -u +%Y-%m-%dT%H:%M:%SZ) pruned $DELETED old backup(s) (retention: ${RETENTION_DAYS}d)"
