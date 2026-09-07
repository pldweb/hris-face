# End-to-end tests and accuracy benchmarks

Everything here runs the **real** stack: real Go API, real Postgres/pgvector, real
ArcFace + MiniFASNet, and (for the browser suite) real Chrome. Nothing is stubbed.

| Script | What it proves |
|---|---|
| `e2e_api.py` | 101 assertions over HTTP: auth + refresh rotation, RBAC, employee/department creation, enrollment validation, duplicate-face rejection, pgvector matching, attendance rules |
| `e2e_browser.js` | 22 assertions through Chrome: login → create employee → enrol (3 photos) → check in → check out, plus guard redirects |
| `e2e_admin_browser.js` | 18 assertions: HR monitoring + CSV export, master data CRUD, device approval, and the employee correction → HR approval round trip |
| `test_challenge.py` | 13 assertions on the movement challenge: a still image repeated cannot pass, genuinely different frames can, a stranger's frame mid-sequence fails the whole thing |
| `benchmark_accuracy.py` | FAR/FRR of face matching on LFW verification pairs |
| `benchmark_liveness.py` | Spoof-detection vs live-rejection rate on CelebA-Spoof |

Measured results and their caveats live in `docs/PRD.md` section 3.1.

## Test identities

Real faces, not synthetic drawings: SCRFD rejects drawn faces, and the quality gate
needs a face at least 112px wide. `e2e_api.py` pulls three LFW identities (two photos
each), upscales them 2x to look like a 640x480 webcam capture, and enrols brightness
variants of one photo while checking in with the *other* photo — so matching is tested
across genuinely different images, not against a copy of what was enrolled.

The browser suite feeds Chrome a fake webcam (`--use-file-for-fake-video-capture`)
built from the same faces:

```bash
ffmpeg -y -loop 1 -i /tmp/e2e_faces/A1.jpg -t 6 -r 15 \
  -vf "scale=-1:480,pad=640:480:(ow-iw)/2:0" -pix_fmt yuv420p fakecam/realA1.y4m
```

## Running

```bash
# 1. Test database
docker run -d --name hris-test-db \
  -e POSTGRES_PASSWORD=testpass -e POSTGRES_DB=hris \
  -p 55432:5432 pgvector/pgvector:pg16
# tunggu database benar-benar ada (pg_isready true sebelum POSTGRES_DB dibuat)
until docker exec hris-test-db psql -U postgres -d hris -c 'SELECT 1' >/dev/null 2>&1; do sleep 1; done
docker exec hris-test-db psql -U postgres -d hris \
  -c "CREATE EXTENSION vector; CREATE EXTENSION citext; CREATE EXTENSION pgcrypto;"

# 2. Face service (see face/README.md for the anti-spoof model export)
cd face && .venv/bin/uvicorn app.main:app --host 127.0.0.1 --port 8000 &

# 3. API (migrations run on start) + first admin
cd api && go build -o bin/api ./cmd/api && go build -o bin/seed ./cmd/seed
export DATABASE_URL="postgres://postgres:testpass@127.0.0.1:55432/hris?sslmode=disable"
JWT_SECRET=test FACE_SERVICE_URL=http://127.0.0.1:8000 \
  OFFICE_IP_ALLOWLIST="127.0.0.0/8,::1/128" ./bin/api &
SEED_EMAIL=hr@perusahaan.com SEED_PASSWORD=rahasia-sekali-123 ./bin/seed

# 4. Suites
cd face && .venv/bin/python ../test/e2e/e2e_api.py
cd face && .venv/bin/python ../test/e2e/test_challenge.py
cd web && npm run dev &
node test/e2e/e2e_browser.js
node test/e2e/e2e_admin_browser.js

# 6. Unit tests (no stack needed)
cd api && go test ./...

# 5. Benchmarks (slow: ~30 min for the full LFW sweep)
cd face && .venv/bin/python ../test/e2e/benchmark_accuracy.py [--limit N]
cd face && .venv/bin/python ../test/e2e/benchmark_liveness.py [--limit N]
```

Each suite assumes a freshly truncated database:

```sql
TRUNCATE attendances, face_embeddings, devices, attendance_corrections,
         audit_logs, refresh_tokens, employees, users, departments CASCADE;
```

Running them back to back without a reset fails on leftover records, not on real
defects.

## What these tests still do not cover

- **Liveness under real capture conditions.** CelebA-Spoof's tight crops starve the
  anti-spoof model of the surrounding context it relies on, so its numbers there are
  pessimistic and not a basis for setting the threshold. See `docs/PRD.md` 3.1.
- **Real-world FRR.** LFW is far harder than an office webcam; the production number
  needs measuring on the company's own captures.
