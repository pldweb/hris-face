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
rsync -az --delete face/app face/scripts face/requirements.txt "$HRIS_HOST:/opt/hris/face/"

echo "==> restart"
ssh "$HRIS_HOST" bash -s <<'REMOTE'
set -euo pipefail
mv /opt/hris/api/api.new /opt/hris/api/api
chmod +x /opt/hris/api/api
source /opt/hris/face/.venv/bin/activate && pip install -q -r /opt/hris/face/requirements.txt

# The anti-spoof models are deliberately not vendored (face/README.md: generated
# from upstream weights so provenance is verifiable, not pulled as an opaque
# binary) and this script never generates them. Without this check a missing
# model crash-loops hris-face silently after every deploy until someone notices
# absensi is down.
if ! ls /opt/hris/face/models/*.onnx >/dev/null 2>&1; then
  echo "FATAL: /opt/hris/face/models/*.onnx missing -- hris-face will refuse to start." >&2
  echo "One-time setup on this host (see face/README.md 'Anti-spoof models'):" >&2
  echo "  cd /opt/hris/face && .venv/bin/pip install torch --index-url https://download.pytorch.org/whl/cpu onnxscript" >&2
  echo "  cd /opt/hris/face && .venv/bin/python scripts/export_antispoof_onnx.py" >&2
  exit 1
fi

sudo systemctl restart hris-api hris-face
sudo systemctl reload caddy
REMOTE

echo "==> done"
