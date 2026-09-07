#!/usr/bin/env bash
# Brings the whole stack up locally for manual testing, then tears it down on Ctrl+C.
# Docker is used ONLY for the throwaway test database -- production deploy stays
# Docker-free (see deploy/).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

DB_CONTAINER=hris-local-db
DB_VOLUME=hris-local-db-data
DB_PORT=55432
export DATABASE_URL="postgres://postgres:testpass@127.0.0.1:${DB_PORT}/hris?sslmode=disable"
export JWT_SECRET="local-dev-secret-change-me"
export FACE_SERVICE_URL="http://127.0.0.1:8000"
export OFFICE_IP_ALLOWLIST="127.0.0.0/8,::1/128"
export LISTEN_ADDR="127.0.0.1:8080"

ADMIN_EMAIL="hr@perusahaan.com"
ADMIN_PASSWORD="rahasia-sekali-123"

GO_BIN="$(command -v go || echo /snap/bin/go)"
PIDS=()

cleanup() {
  echo
  echo "==> menghentikan service..."
  for pid in "${PIDS[@]:-}"; do kill "$pid" 2>/dev/null || true; done
  docker rm -f "$DB_CONTAINER" >/dev/null 2>&1 || true
  echo "==> selesai."
}
trap cleanup EXIT INT TERM

echo "==> 1/6 database (pgvector)"
docker rm -f "$DB_CONTAINER" >/dev/null 2>&1 || true
docker volume create "$DB_VOLUME" >/dev/null
docker run -d --name "$DB_CONTAINER" \
  -e POSTGRES_PASSWORD=testpass -e POSTGRES_DB=hris \
  -v "$DB_VOLUME":/var/lib/postgresql/data \
  -p ${DB_PORT}:5432 pgvector/pgvector:pg16 >/dev/null
# pg_isready goes true during the container's init phase, before POSTGRES_DB
# exists. Wait for the database itself to answer, not just the server.
until docker exec "$DB_CONTAINER" psql -U postgres -d hris -c 'SELECT 1' >/dev/null 2>&1; do sleep 1; done
docker exec "$DB_CONTAINER" psql -U postgres -d hris \
  -c "CREATE EXTENSION IF NOT EXISTS vector; CREATE EXTENSION IF NOT EXISTS citext; CREATE EXTENSION IF NOT EXISTS pgcrypto;" >/dev/null

echo "==> 2/6 face service (ArcFace + anti-spoof)"
if [ ! -d face/.venv ]; then
  echo "    face/.venv belum ada. Jalankan dulu:"
  echo "      cd face && python3 -m venv .venv && .venv/bin/pip install -r requirements.txt"
  exit 1
fi
if ! ls face/models/*.onnx >/dev/null 2>&1; then
  echo "    Model anti-spoof belum ada. Jalankan dulu:"
  echo "      cd face && .venv/bin/pip install torch --index-url https://download.pytorch.org/whl/cpu onnxscript"
  echo "      cd face && .venv/bin/python scripts/export_antispoof_onnx.py"
  exit 1
fi
(cd face && .venv/bin/uvicorn app.main:app --host 127.0.0.1 --port 8000 >/tmp/hris-face.log 2>&1) &
PIDS+=($!)
echo "    memuat model (unduhan pertama ~326 MB, sabar)..."
until curl -s --max-time 2 http://127.0.0.1:8000/healthz >/dev/null 2>&1; do
  sleep 2
  kill -0 "${PIDS[-1]}" 2>/dev/null || { echo "    face service mati, lihat /tmp/hris-face.log"; exit 1; }
done

echo "==> 3/6 build Go"
(cd api && "$GO_BIN" build -o bin/api ./cmd/api && "$GO_BIN" build -o bin/seed ./cmd/seed)

echo "==> 4/6 API (migrasi jalan otomatis)"
(cd api && ./bin/api >/tmp/hris-api.log 2>&1) &
PIDS+=($!)
until curl -s --max-time 2 http://127.0.0.1:8080/healthz >/dev/null 2>&1; do sleep 1; done

echo "==> 5/6 admin pertama"
if ! docker exec "$DB_CONTAINER" psql -U postgres -d hris -tAc \
  "SELECT 1 FROM users WHERE email = '$ADMIN_EMAIL'" | grep -q 1; then
  (cd api && SEED_EMAIL="$ADMIN_EMAIL" SEED_PASSWORD="$ADMIN_PASSWORD" ./bin/seed) 2>&1 | tail -1
fi

echo "==> 6/6 frontend"
[ -d web/node_modules ] || (cd web && npm install)
(cd web && npm run dev >/tmp/hris-web.log 2>&1) &
PIDS+=($!)
until curl -s --max-time 2 http://localhost:5173 >/dev/null 2>&1; do sleep 1; done

cat <<EOF

────────────────────────────────────────────────────────────
  Siap. Buka: http://localhost:5173

  Login HR:  $ADMIN_EMAIL
             $ADMIN_PASSWORD

  Alurnya:
    1. Masuk sebagai HR di atas
    2. /admin/employees -> Tambah Karyawan
       (catat password sementara yang muncul, hanya tampil sekali)
    3. Keluar, masuk sebagai karyawan itu
    4. Buka /enroll -> ambil 3 foto -> Kirim
    5. Otomatis ke /checkin -> absen

  Kamera butuh izin browser. localhost dihitung secure context,
  jadi getUserMedia jalan tanpa HTTPS di sini saja.

  Log: /tmp/hris-{face,api,web}.log
  Ctrl+C untuk menghentikan semuanya.
────────────────────────────────────────────────────────────

EOF

wait
