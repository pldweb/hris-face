# HRIS Absensi — Face Recognition

Lihat [docs/PRD.md](docs/PRD.md) untuk spesifikasi lengkap dan [PRODUCT.md](PRODUCT.md) untuk konteks produk.

## Struktur

```
api/     Go — REST API, aturan bisnis absensi, auth, migrasi database
         cmd/mcp — MCP server (read-only) untuk asisten AI, lihat docs/MCP.md
face/    Python — deteksi wajah, embedding ArcFace, anti-spoof (internal, 127.0.0.1 saja)
web/     Vue 3 + Ant Design Vue — SPA karyawan & admin
deploy/  systemd units, Caddyfile, deploy.sh — VPS tanpa Docker
```

## Coba sendiri (satu perintah)

```bash
./run-local.sh
```

Menyalakan database, face service, API, dan frontend sekaligus, lalu mencetak URL
dan kredensial HR. Ctrl+C menghentikan semuanya. Prasyarat sekali jalan:

```bash
cd face && python3 -m venv .venv && .venv/bin/pip install -r requirements.txt
.venv/bin/pip install torch --index-url https://download.pytorch.org/whl/cpu onnxscript
.venv/bin/python scripts/export_antispoof_onnx.py
```

## Development

**Database**
```bash
sudo -u postgres createdb hris
sudo -u postgres psql hris -c "CREATE EXTENSION vector; CREATE EXTENSION citext;"
```

**API**
```bash
cd api
export DATABASE_URL=postgres://localhost/hris?sslmode=disable
export JWT_SECRET=dev-secret
go run ./cmd/api
```

**Face service** — lihat [face/README.md](face/README.md), termasuk cara memperoleh model anti-spoof.
```bash
cd face && python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --port 8000
```

**Web**
```bash
cd web
npm install
npm run dev
```

Dev server mem-proxy `/api/*` ke `127.0.0.1:8080` (lihat `vite.config.ts`). Tanpa Go API jalan, request balik 502 — itu memang perilaku yang benar; sebelumnya path `/api` jatuh ke `index.html` dan mengembalikan HTML dengan status 200.

## Deploy ke VPS

Lihat `deploy/deploy.sh` dan `docs/PRD.md` bagian 11. Ringkas:

```bash
HRIS_HOST=deploy@hris.perusahaan.com ./deploy/deploy.sh
```

Sekali di VPS: install Caddy, PostgreSQL 16 + pgvector, salin `deploy/*.service` ke `/etc/systemd/system/`, `deploy/Caddyfile` ke `/etc/caddy/Caddyfile`, `deploy/api.env.example` ke `/etc/hris/api.env` (isi nilainya), lalu `systemctl enable --now hris-api hris-face caddy`.

## Status

**Sudah jalan:**
- Auth: login, refresh token (rotasi, cookie httpOnly), logout, route guard di frontend
- Check-in & check-out wajah: ArcFace + cosine similarity, anti-spoof, IP allowlist (`web/src/views/CheckIn.vue`)
- Status kehadiran dihitung dari jadwal kerja (`on_time` / `late` / `early_leave`), zona waktu configurable via `TIMEZONE`
- Device binding: maks 2 perangkat per karyawan, cookie httpOnly, approval HR
- Enrollment wajah: 3–5 foto, validasi kualitas per foto, deteksi duplikat antar karyawan
- Admin: CRUD karyawan + departemen, tabel dengan status kehadiran
- **Monitoring harian HR**: 4 kartu statistik (tepat waktu / terlambat / belum absen / sudah pulang), tabel berfilter (periode, departemen, status), export CSV yang mengikuti filter aktif
- **Koreksi absen**: karyawan mengajukan, HR menyetujui/menolak; yang disetujui menghasilkan record bertanda `manual_correction`
- **Riwayat karyawan**: rekap per hari (masuk, pulang, durasi) per bulan
- **Dashboard manajer**: layar monitoring yang sama, dibatasi server ke bawahan langsung (`employees.manager_id`)
- **Master data**: CRUD jadwal kerja & lokasi, approval perangkat
- **Import karyawan dari CSV**: per-baris, satu baris gagal tidak menggagalkan sisanya; departemen dibuat otomatis
- **Re-enroll wajah**: embedding lama diarsipkan (`is_active = false`), tidak dihapus
- **Foto bukti + retensi**: opsional lewat `PHOTO_DIR`, terhapus otomatis setelah 30 hari
- **Challenge gerakan**: saat anti-spoof ragu (bukan langsung menolak), sistem meminta beberapa frame dan menolak foto diam
- **Export Excel (.xlsx)**: tanggal & jam sebagai sel bertipe, header beku, autofilter
- **Notifikasi email**: koreksi masuk ke HR, hasil review ke karyawan (SMTP opsional — tanpa konfigurasi, dicatat ke log, tidak pernah memblokir)
- **Rate limiting**: per-akun untuk login (bukan per-IP, karena satu kantor berbagi satu IP NAT), plus batas longgar per-IP
- Deploy path: systemd + Caddy + `deploy.sh`

**Belum:** kalibrasi threshold liveness pada kondisi tangkapan sungguhan — satu-satunya yang tersisa, dan itu butuh data webcam kantor asli, bukan pekerjaan kode. Prosedurnya di `docs/PRD.md` bagian 9.

**Sudah diuji end-to-end dengan model sungguhan** — Go + Postgres/pgvector asli, ArcFace + MiniFASNet asli, Chrome asli, tanpa stub: **108 asersi API + 13 asersi challenge + 22 + 18 + 6 asersi browser + unit test retensi foto**, semua lolos. Lihat [test/e2e/](test/e2e/).

**Akurasi terukur** (detail + peringatan di [docs/PRD.md](docs/PRD.md) bagian 3.1):
- Pengenalan wajah (LFW, 2.196 pasangan): FAR **0.000%** ✅, FRR **7.8%** (target <3%)
- Anti-spoof (CelebA-Spoof): **75.6%** spoof tertangkap pada threshold 0.5 — angka ini belum layak jadi dasar keputusan, lihat catatan

**Bootstrap admin pertama:**
```bash
DATABASE_URL=... SEED_EMAIL=admin@perusahaan.com SEED_PASSWORD=... ./seed
```

**Konfigurasi opsional:** `TIMEZONE` (default `Asia/Jakarta`), `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM`, `PHOTO_DIR` (kosong = foto bukti tidak disimpan sama sekali), `PHOTO_RETENTION_DAYS` (default 30).

**Catatan teknis:**
- Model anti-spoof digenerate dari bobot resmi Apache-2.0; lihat `face/README.md`.
- ⚠️ **Threshold liveness 0.5 belum tervalidasi pada kondisi tangkapan sungguhan — blocker go-live.** Prosedur kalibrasinya di PRD bagian 9.
- Bundle JS utama ~478 kB gzip (Ant Design Vue penuh). Kalau jadi masalah di jaringan kantor, pasang `unplugin-vue-components` untuk import on-demand.
