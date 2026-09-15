#!/usr/bin/env bash
# docs/PRD.md section 11.5 -- daily pg_dump, 14-day local retention, synced
# off-site. Run by hris-backup.timer as the hris user. A backup that only
# lives on the same VPS as the database isn't a backup (PRD 11.5, 11.6).
set -euo pipefail

ENV_FILE="/etc/hris/api.env"
BACKUP_DIR="/var/backups/hris"
RETENTION_DAYS=14

# shellcheck source=/dev/null
source "$ENV_FILE"
: "${DATABASE_URL:?DATABASE_URL missing from $ENV_FILE}"

mkdir -p "$BACKUP_DIR"
STAMP="$(date +%F_%H%M%S)"
DUMP_PATH="$BACKUP_DIR/hris-$STAMP.sql.gz"

pg_dump "$DATABASE_URL" | gzip > "$DUMP_PATH"
echo "==> wrote $DUMP_PATH ($(du -h "$DUMP_PATH" | cut -f1))"

find "$BACKUP_DIR" -name 'hris-*.sql.gz' -mtime "+${RETENTION_DAYS}" -delete

# Off-site sync: set BACKUP_REMOTE in api.env to an rsync destination, e.g.
#   BACKUP_REMOTE=backup@offsite.example.com:/backups/hris/
# Left unset, this step is skipped -- but per PRD 11.5/11.6 that means one
# VPS failure can take the only copy of attendance data with it.
if [ -n "${BACKUP_REMOTE:-}" ]; then
  rsync -az "$BACKUP_DIR"/ "$BACKUP_REMOTE"
  echo "==> synced to $BACKUP_REMOTE"
else
  echo "WARNING: BACKUP_REMOTE not set in $ENV_FILE -- backups are local-only." >&2
fi
