#!/usr/bin/env bash
# docs/PRD.md section 11.4 -- build locally, rsync to the VPS, restart services.
# No Docker, no CI runner: one script, run by hand or from a cron/CI trigger later.
set -euo pipefail

: "${HRIS_HOST:?set HRIS_HOST, e.g. deploy@hris.perusahaan.com}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> build: Go API"
(cd api && GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api)

echo "==> build: web SPA"
(cd web && npm ci && npm run build)

echo "==> upload"
rsync -az --delete api/bin/api "$HRIS_HOST:/opt/hris/api/api.new"
rsync -az --delete web/dist/ "$HRIS_HOST:/var/www/hris/"
rsync -az --delete face/app face/requirements.txt "$HRIS_HOST:/opt/hris/face/"

echo "==> restart"
ssh "$HRIS_HOST" bash -s <<'REMOTE'
set -euo pipefail
mv /opt/hris/api/api.new /opt/hris/api/api
chmod +x /opt/hris/api/api
source /opt/hris/face/.venv/bin/activate && pip install -q -r /opt/hris/face/requirements.txt
sudo systemctl restart hris-api hris-face
sudo systemctl reload caddy
REMOTE

echo "==> done"
