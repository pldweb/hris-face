# PRD — HRIS Absensi Berbasis Face Recognition

**Versi:** 0.2
**Tanggal:** 2026-09-07
**Owner:** paldiengineer01@gmail.com
**Status:** Draft untuk review

Perubahan dari 0.1: skala dikunci < 200 karyawan, absen hanya lewat laptop/desktop, Ant Design Vue sebagai library frontend, deploy VPS tanpa Docker, dan penambahan bagian Arah Desain Frontend.

---

## 1. Ringkasan

Web app absensi karyawan yang memverifikasi identitas lewat wajah dari webcam laptop, bukan fingerprint atau PIN. Tujuan utamanya menghilangkan **titip absen** tanpa menambah hardware baru.

Stack: **Go** (API + domain HRIS), **Python** (service face embedding & anti-spoof), **Vue 3 + Ant Design Vue** (web app karyawan + dashboard admin), **PostgreSQL + pgvector**. Deploy ke **satu VPS, tanpa Docker**.

## 2. Masalah

| Masalah | Dampak |
|---|---|
| Titip absen (buddy punching) | Payroll bocor, data kehadiran tidak dipercaya |
| Mesin fingerprint | Biaya per cabang, antre, higienis buruk |
| Rekap manual dari beberapa mesin | HR habis waktu tutup bulan, rawan salah hitung lembur |

## 3. Tujuan & Metrik Sukses

| Tujuan | Metrik | Target rilis 1 | Terukur |
|---|---|---|---|
| Absen cepat | Buka halaman s.d. sukses | < 5 detik (p90) | belum diukur |
| Verifikasi akurat | FAR (orang lain diterima) | < 0.1% | **0.000%** ✅ |
| | FRR (orang benar ditolak) | < 3% | **7.8%** ❌ |
| Anti-spoof | Foto/video di layar HP ditolak | > 95% terdeteksi | **75.6%** ❌ (lihat catatan) |
| Beban HR turun | Waktu rekap bulanan | jam → menit (export 1 klik) | belum ada fitur export |
| Adopsi | Karyawan aktif pakai | > 90% dalam 1 bulan | belum rilis |

**Non-goal rilis 1:** payroll & slip gaji, cuti/approval workflow, mobile web, mobile app native, kios tablet, shift rotasi kompleks, integrasi mesin fingerprint lama.

### 3.1 Hasil pengukuran (2026-09-07)

Diukur dengan kode produksi lewat `test/e2e/benchmark_accuracy.py` dan
`benchmark_liveness.py`.

**Pengenalan wajah — LFW, 1.098 pasangan sama-orang + 1.098 beda-orang:**

| Threshold | FAR | FRR |
|---|---|---|
| 0.20 | 0.000% | 6.47% |
| 0.35 | 0.000% | 6.56% |
| **0.45 (dipakai)** | **0.000%** | **7.83%** |
| 0.50 | 0.000% | 9.84% |

Skor beda-orang tertinggi dari 1.098 pasangan: **0.198** — threshold 0.45 punya
margin 2,3× di atasnya. FAR aman dengan lapang.

FRR meleset dari target dan **tidak bisa diperbaiki dengan menurunkan threshold**:
lantainya 6,5% bahkan di 0.20. Tapi LFW jauh lebih sulit daripada kasus kita —
foto berita selebritas berjarak puluhan tahun, pose dan pencahayaan liar, resolusi
rendah. Kasus kita: karyawan di depan laptopnya sendiri, pencahayaan kantor tetap,
baru enroll. Ditambah sistem menyimpan **3–5 embedding per orang dan lolos bila ada
satu yang cocok**, sementara LFW mengukur 1-lawan-1. FRR produksi akan lebih rendah,
tapi **berapa persisnya belum diketahui** sampai diukur dengan data kantor sendiri.

**Anti-spoof — CelebA-Spoof, 400 live + 398 spoof:**

| Threshold | Spoof tertangkap | Wajah asli ditolak |
|---|---|---|
| 0.05 | 40.2% | 3.5% |
| **0.50 (dipakai)** | **75.6%** | **35.8%** |
| 0.90 | 93.7% | 71.7% |

⚠️ **Angka ini belum layak jadi dasar keputusan.** 396 dari 398 gambar CelebA sudah
ter-crop ketat sehingga deteksi wajah gagal dan seluruh gambar dipakai sebagai bbox.
MiniFASNet justru mengandalkan konteks di sekitar wajah (tepi layar, moiré, bingkai
foto) yang persis hilang pada crop ketat — jadi ini pengukuran yang menghukum model
di luar kondisi rancangannya. Pada sampel ground-truth milik repo resmi, dengan frame
penuh sebagaimana produksi mengirimnya, pemisahannya bersih: asli **1.000**, palsu
**0.18** dan **0.002**.

**Konsekuensi:** threshold liveness 0.5 **tidak boleh dianggap tervalidasi**. Kalau
angka 35,8% itu terbawa ke produksi, satu dari tiga absen sah akan gagal. Sebelum
go-live wajib diukur ulang dengan tangkapan webcam kantor sungguhan (asli) dan
percobaan spoof sungguhan (foto di layar HP, cetakan). Prosedurnya di bagian 9.

## 4. Skala Target

< 200 karyawan, 1–2 lokasi. Konsekuensi teknis yang mengikat seluruh dokumen ini:

- **Tanpa index vektor.** 200 karyawan x 5 embedding = 1.000 vektor. Sequential scan pgvector di bawah 5 ms. HNSW/IVFFlat ditambahkan hanya kalau jumlah karyawan lewat ~5.000.
- **Satu VPS cukup.** 4 vCPU / 8 GB. Tidak ada load balancer, tidak ada replica.
- **Puncak beban rendah.** ~200 absen tersebar 07:00–09:00, bukan 200 request bersamaan.
- **Tidak perlu antrian/worker.** Verifikasi sinkron di request cycle.

## 5. Pengguna

1. **Karyawan** — check-in/out dari laptop, lihat riwayat sendiri, ajukan koreksi.
2. **HR / Admin** — kelola karyawan, jadwal, lokasi, approve koreksi, export rekap.
3. **Manajer** — lihat kehadiran timnya.
4. **Super Admin** — konfigurasi sistem, threshold, audit log.

## 6. Alur Utama

### 6.1 Enrollment (pendaftaran wajah)
1. HR buat data karyawan → kirim link/token enrollment.
2. Karyawan buka link di laptop, ambil **3–5 foto** (hadap depan, kiri, kanan, dengan/tanpa kacamata).
3. Backend cek kualitas (blur, pencahayaan, tepat 1 wajah, lebar wajah min. 112px) → ekstrak embedding → simpan.
4. Cek duplikat: kalau embedding mirip karyawan lain, tolak dan flag ke HR.
5. Status karyawan jadi `active`.

### 6.2 Check-in / Check-out
1. Karyawan buka app → browser minta izin kamera.
2. Deteksi wajah live di browser, auto-capture saat wajah stabil, terang, dan menghadap depan.
3. Kirim frame + device fingerprint ke API (lokasi: lihat 7.4).
4. Backend: anti-spoof → ekstrak embedding → cari kandidat terdekat (pgvector) → cocokkan.
5. Validasi: jaringan/lokasi, jadwal kerja, duplikat absen, device binding.
6. Simpan record → tampilkan hasil (nama, jam, status tepat waktu / terlambat).

### 6.3 Koreksi
Gagal wajah 3x atau lupa absen → ajukan koreksi manual dengan alasan → antre approval HR → tercatat di audit log.

## 7. Kebutuhan Fungsional

### F1 — Autentikasi & Otorisasi
- Login email + password, JWT (access 15 menit + refresh token httpOnly cookie).
- RBAC: `employee`, `manager`, `hr`, `superadmin`.
- Rate limit login **per akun** (5 percobaan gagal / 15 menit), bukan per IP.
  Per-IP salah untuk produk ini: seluruh kantor keluar lewat satu IP NAT, jadi
  batas per-IP mengunci orang ke-11 di jam 08:00 sambil tidak menghambat penyerang
  yang berganti IP. Batas per-IP tetap ada tapi longgar, hanya untuk menahan flood.

### F2 — Master Data
- CRUD karyawan (NIK, nama, email, jabatan, departemen, atasan, tanggal masuk, status).
- CRUD departemen dan lokasi kerja.
- CRUD jadwal kerja: jam masuk, jam pulang, toleransi terlambat, hari kerja, hari libur.
- Import karyawan dari CSV.

### F3 — Face Enrollment
- Capture multi-foto dari webcam, preview, ambil ulang.
- Validasi kualitas + deteksi duplikat wajah.
- Re-enroll lewat `/enroll?ulang=1` — versi lama diarsip (`is_active = false`),
  bukan dihapus: audit yang bertanya "wajah mana yang tercatat saat absensi ini
  dibuat?" harus tetap bisa dijawab. Endpoint enroll pertama menolak bila sudah
  ada wajah terdaftar, supaya pengiriman kedua tidak diam-diam menumpuk wajah baru.
- HR bisa lihat status enrollment tiap karyawan.

### F4 — Absensi
- Check-in & check-out dengan verifikasi wajah.
- Status otomatis dari `work_schedules`: `on_time`, `late` (lewat jam masuk + toleransi),
  `early_leave` (pulang sebelum jam pulang). Dibandingkan pada waktu lokal (`TIMEZONE`,
  default `Asia/Jakarta`) -- `occurred_at` disimpan UTC, jadi tanpa ini penilaian
  terlambat meleset 7 jam. Karyawan baru otomatis mendapat jadwal default; tanpa
  jadwal, tidak ada yang bisa dinilai terlambat.
- **Identitas sesi mengikat.** Wajah yang cocok harus milik user yang sedang login.
  Mengidentifikasi dari wajah saja membuat siapa pun yang login bisa mengabsenkan
  rekannya cukup dengan mengarahkan webcam ke orang itu -- titip absen dengan
  langkah tambahan. Wajah orang lain dijawab "tidak dikenali", tidak pernah
  menyebut nama siapa yang cocok.
- Anti duplikat: tidak bisa check-in 2x dalam 1 hari kerja.
- Timestamp **selalu dari server**, tidak pernah dari klien.
- Retry otomatis kalau jaringan putus saat submit (maks. 3x / 60 detik).

### F5 — Anti-Spoofing (wajib rilis 1)
- Passive liveness dari 1 frame (MiniFASNet) — tolak foto cetak & layar HP.
- **Challenge gerakan** bila skor liveness ambigu (antara 0.15 dan 0.5): sistem
  meminta beberapa frame, bukan langsung menolak. Yang **diverifikasi**: setiap
  frame adalah user yang login, frame benar-benar berbeda satu sama lain (foto
  diam gagal), dan rata-rata skor anti-spoof di atas ambang.
- ⚠️ **Deteksi kedip tidak dijadikan gerbang.** Keterbukaan mata diukur dan
  dicatat, tapi model landmark 106-titik meregresi bentuk mata yang masuk akal
  alih-alih melacak kelopak — ditutupnya mata secara sintetis tidak menggerakkan
  angkanya sama sekali, dan kami tidak punya rekaman mata tertutup untuk
  memvalidasinya. Mengirimkannya sebagai gerbang berarti mengirim kontrol
  keamanan yang belum terbukti bekerja, yang lebih buruk daripada tidak ada.
- Simpan skor liveness di setiap record untuk audit.

### F6 — Dashboard & Laporan
- Karyawan: riwayat sendiri, rekap jam kerja bulan berjalan.
- Manajer: kehadiran tim hari ini + rekap bulanan.
- HR: monitoring realtime hari ini, rekap per departemen, daftar terlambat/absen.
- Export Excel & CSV; filter tanggal, departemen, karyawan.

### F7 — Koreksi & Audit
- Pengajuan koreksi absen dengan alasan.
- Approval/reject oleh HR dengan catatan.
- Audit log immutable: siapa, kapan, aksi apa, data lama → data baru.

### F8 — Notifikasi
- Email: koreksi menunggu approval (ke HR), hasil review koreksi (ke karyawan).
- SMTP opsional. Tanpa `SMTP_HOST`, pesan dicatat ke log dan alurnya tetap jalan:
  mail server yang mati tidak boleh menggagalkan pengajuan koreksi yang sudah tersimpan.
- Belum: peringatan gagal absen berulang, notifikasi perangkat baru.

## 7.4 — Validasi Lokasi (berubah dari v0.1)

**Geofence GPS dibatalkan.** Karena absen hanya dari laptop/desktop, browser tidak punya GPS — Geolocation API di desktop memakai posisi WiFi/IP dengan akurasi 20 m sampai beberapa kilometer, dan bisa dipalsukan dari devtools dalam 5 detik. Memakainya sebagai syarat lolos akan menolak karyawan yang jujur sekaligus meloloskan yang curang.

Penggantinya:

1. **IP allowlist jaringan kantor** (blocking). Absen `on_site` hanya diterima dari IP publik kantor. Sederhana, tidak bisa dipalsu dari browser, dan cocok dengan kenyataan "absen dari laptop kerja di meja".
2. **Flag `allow_remote` per karyawan** untuk WFH. Absen dari luar jaringan kantor diterima tapi ditandai `remote` di laporan.
3. **Device binding.** Maks. 2 perangkat *disetujui* per karyawan.
   ⚠️ Bug produksi yang pernah ada dan sudah diperbaiki: cookie perangkat dulu
   hanya dikirim di jalur sukses. Karyawan yang salah klik "pulang" sebelum
   "masuk" (atau kena retry jaringan yang justru diwajibkan poin F4 di bawah)
   kehabisan jatah 2 perangkat dalam dua kali gagal — dari laptop yang sama,
   tanpa konkurensi apa pun. Sekarang cookie dikirim begitu device diresolusi,
   terlepas dari sukses-tidaknya absensi itu sendiri. Diuji permanen di
   `test/e2e/e2e_api.py` bagian "COOKIE PERANGKAT PADA PERCOBAAN GAGAL". Perangkat dikenali
   lewat cookie `device_id` yang **diterbitkan server** (httpOnly, Secure, SameSite=Lax)
   -- bukan nilai yang dibuat browser, supaya tidak bisa dipalsukan dari JS. Perangkat
   dalam batas mendaftar sendiri; yang melebihi batas diparkir sebagai satu baris
   pending untuk disetujui HR lewat `GET/POST /admin/devices`. Hanya satu baris pending
   per karyawan: menerbitkan baris baru tiap percobaan akan membuat browser tanpa
   cookie membanjiri tabel.
4. **Geolocation tetap direkam** bila browser mengizinkan, sebagai sinyal audit — **tidak pernah sebagai penentu lolos**.

## 8. Kebutuhan Non-Fungsional

| Aspek | Target |
|---|---|
| Performa | Verifikasi wajah p95 < 1.5 detik |
| Konkurensi | 30 request absen bersamaan |
| Ketersediaan | 99.5%, wajib hidup 06:00–10:00 & 16:00–20:00 |
| Keamanan | HTTPS wajib (getUserMedia menuntutnya), password Argon2id |
| Privasi | Foto bukti **mati secara default** (`PHOTO_DIR` kosong). Bila diaktifkan, terhapus otomatis setelah `PHOTO_RETENTION_DAYS` (default 30) oleh binary itu sendiri, bukan cron yang bisa lupa dipasang |
| Kepatuhan | UU PDP: consent eksplisit saat enrollment, hak hapus data biometrik saat resign |
| Browser | Chrome/Edge/Firefox/Safari terbaru, desktop |
| Bahasa | UI Indonesia (struktur siap i18n) |

## 9. Arsitektur

```
Browser (Vue 3 + Ant Design Vue)
        │  HTTPS
        ▼
   Caddy  ── static SPA  +  reverse proxy /api
        │
        ▼
   Go API (systemd)  ──localhost:8000──►  Python Face Service (systemd)
        │                                        │
        ▼                                        ▼
  PostgreSQL + pgvector                   ONNX Runtime (CPU)
        │
        ▼
  /var/lib/hris/photos  (foto bukti, opsional)
```

### Kenapa 3 bahasa

- **Go** — API, auth, aturan bisnis absensi, transaksi, export. Satu binary, tanpa runtime, cocok untuk VPS tanpa Docker.
- **Python** — hanya ML: deteksi wajah, embedding, anti-spoof. Ekosistem model wajah matang di Python; padanan Go butuh binding dlib dan akurasinya lebih rendah. Service ini **stateless**: terima gambar, balikin vektor + skor. Bind ke `127.0.0.1`, tidak pernah terekspos publik.
- **Vue 3 + TypeScript + Ant Design Vue** — SPA, kamera via `getUserMedia`, deteksi wajah ringan di klien hanya untuk auto-capture/framing.

Keputusan yang tidak bisa ditawar: **semua verifikasi di server**. Klien tidak pernah menentukan lolos/tidak.

### Model
| Fungsi | Model | Catatan |
|---|---|---|
| Deteksi wajah | SCRFD (InsightFace) | + 5 landmark untuk alignment |
| Embedding | ArcFace `buffalo_l` (512-dim) | ONNX Runtime CPU |
| Anti-spoof | MiniFASNet (Silent-Face) | ringan, 1 frame |
| Klien (auto-capture) | MediaPipe Face Detection (WASM) | framing saja, bukan verifikasi |

### Pencocokan — ArcFace + Cosine Similarity (final)

Metode dikunci: **ArcFace `buffalo_l` (InsightFace, 512-dim) untuk embedding, cosine similarity untuk pencocokan.** Bukan Euclidean, bukan model lain — ArcFace dilatih memakai angular margin loss, jadi cosine similarity adalah metrik jarak yang metodenya sendiri dioptimalkan untuk itu.

- Similarity dihitung `1 - cosine_distance` antara embedding hasil `/face/embed` dan tiap embedding tersimpan milik karyawan; skala 0–1, makin besar makin mirip.
- Threshold **0.45**, sudah divalidasi terhadap LFW (bagian 3.1): FAR 0.000% dengan
  margin 2,3× di atas skor impostor terburuk. Dipertahankan meski 0.20–0.35 memberi
  FRR sedikit lebih baik, karena kesalahan di dua arah tidak setara: false accept =
  titip absen lolos (justru masalah yang produk ini ada untuk mencegah), false reject
  = karyawan mengulang atau lewat koreksi manual. Asimetri itu membenarkan margin lebih.
- Simpan semua 3–5 embedding hasil enrollment per karyawan (bukan hanya centroid) — cocok bila ada **minimal satu** yang lolos threshold. Variasi sudut/pencahayaan antar foto membuat satu embedding representatif tidak cukup.
- Identifikasi 1:N: hitung similarity terhadap seluruh embedding aktif di database (< 200 karyawan × 5 = ~1.000 vektor), ambil skor tertinggi. Kolom `vector(512)`, tanpa index (lihat bagian 4) — di skala ini sequential scan lebih cepat dan lebih sederhana daripada membangun index approximate.
- Kalau skor tertinggi < threshold: ditolak, tidak ada "kandidat terdekat dipaksa cocok".
- Kalau dua karyawan berbeda sama-sama di atas threshold (kembar/mirip): ambil skor tertinggi, tapi jarak keduanya < 0.05 memicu flag `low_confidence` ke HR untuk review manual, bukan auto-approve diam-diam.

**Kalibrasi threshold:** selama 2 minggu pertama produksi, simpan skor similarity mentah di setiap `attendances` (kolom sudah ada di skema) walau keputusan lolos/tolak tetap pakai 0.45. Setelah itu, plot distribusi skor genuine (match benar, dikonfirmasi manual oleh HR) vs impostor (percobaan gagal) dan pilih threshold di titik EER (equal error rate) sebagai nilai produksi. Ini pekerjaan satu kali di akhir Fase 1, dicatat di F5/anti-spoof sebagai bagian dari go-live checklist.

**Kalibrasi liveness — blocker go-live, bukan pekerjaan opsional.** Threshold
`minLivenessScore = 0.5` di `api/internal/attendance/service.go` belum tervalidasi
pada kondisi tangkapan sungguhan (bagian 3.1). Prosedur minimum sebelum dipakai
karyawan:

1. Kumpulkan ±100 tangkapan webcam kantor dari karyawan sungguhan (asli), pada
   pencahayaan pagi dan sore.
2. Kumpulkan ±100 percobaan spoof: foto wajah ditampilkan di layar HP, layar laptop,
   dan cetakan kertas — diambil dengan webcam yang sama.
3. Jalankan `test/e2e/benchmark_liveness.py` pada data itu, pilih threshold di mana
   penolakan wajah asli < 2% sambil menangkap spoof setinggi mungkin.
4. Kalau tidak ada threshold yang memenuhi keduanya, jangan paksakan angka: mode
   challenge (kedip/hadap samping) di F5 naik dari fallback jadi wajib.

**Jebakan implementasi:** MiniFASNet menuntut input **[0, 255]**, bukan [0, 1] —
repo aslinya sengaja membuang pembagian 255 di `to_tensor` mereka. Membaginya
membuat model mengeluarkan kelas nyaris konstan untuk input apa pun: service tetap
hidup, skor tetap keluar, tanpa error — dan semua spoof lolos diam-diam. Sudah
diperbaiki dan dikomentari di `face/app/liveness.py`; jangan "dirapikan" kembali.

### Skema data inti
`users`, `employees`, `departments`, `work_locations`, `work_schedules`,
`face_embeddings (employee_id, embedding vector(512), quality_score, is_active, created_at)`,
`attendances (employee_id, type, occurred_at, similarity, liveness_score, ip_address, lat, lng, device_id, photo_path, status, source)`,
`devices`, `attendance_corrections`, `audit_logs`.

### API (garis besar)
```
POST /api/v1/auth/login | refresh | logout
GET  /api/v1/me
POST /api/v1/enrollment/photos
POST /api/v1/attendance/check-in       # body: JPEG mentah; device_id dari cookie
POST /api/v1/attendance/check-out
GET  /api/v1/me                        # profil + absen hari ini
GET  /api/v1/attendance/me?month=
GET  /api/v1/admin/attendances
GET  /api/v1/admin/reports/export
POST /api/v1/corrections | PATCH /api/v1/corrections/:id
CRUD /api/v1/admin/employees|departments|locations|schedules
GET  /api/v1/admin/devices             # perangkat menunggu approval
POST /api/v1/admin/devices/:id/approve
```
Internal (127.0.0.1): `POST /face/embed`, `POST /face/analyze`.

---

## 10. Arah Desain Frontend

**Mode: Operate.** Karyawan sedang mengerjakan tugas, bukan sedang dibujuk. Ekspresi tidak boleh menutupi tugas, state, atau affordance yang sudah dikenal. Standar kualitasnya bukan keunikan, tapi **familiar yang dikerjakan dengan benar** — pengguna langsung percaya, tidak berhenti di tiap komponen yang terasa "agak beda".

### 10.1 Tesis layar absen

Layar absen adalah **satu viewport penuh berisi kamera dan jam** — bukan kartu widget yang menyempil di antara enam kartu statistik di dashboard. Ini menolak default kategori HRIS ("Attendance Card" di pojok kanan atas dashboard), karena absen adalah alasan orang membuka aplikasi ini, bukan catatan kaki dari alasan lain.

Kontrak lengkapnya ada di `.impeccable/surfaces/web-src-views-checkin-vue.md`.

### 10.2 Sistem visual

Ant Design Vue dipakai apa adanya sebagai bahasa komponen — tombol, form, tabel, notifikasi, tanggal. Yang di-override hanya token, lewat `ConfigProvider` + `theme.defaultAlgorithm`:

| Token | Nilai | Alasan |
|---|---|---|
| `colorPrimary` | biru institusional gelap (bukan `#1677FF` bawaan) | AntD default terlihat seperti demo AntD; satu perubahan ini yang paling menentukan identitas |
| `colorBgLayout` | `#F5F6F8` | ground abu-dingin, memisahkan panel dari latar |
| `borderRadius` | `6` | lebih tenang dari default `6`–`8` campuran |
| `fontFamily` | system sans stack | Operate surface; tidak perlu display font |
| `colorSuccess/Warning/Error` | AntD default | **dikunci hanya untuk status kehadiran** |

Aturan warna: **Restrained**. Netral + satu aksen. Aksen hanya untuk aksi primer, item terpilih, dan indikator state — tidak pernah dekorasi. Hijau/kuning/merah tidak boleh dipakai di luar makna kehadiran (hijau = tepat waktu, kuning = terlambat, merah = tidak hadir/gagal), supaya pemindaian tabel HR bisa dilakukan lewat warna saja.

**Angka waktu selalu `font-variant-numeric: tabular-nums`.** Jam, durasi, kolom waktu di tabel. Ini produk yang isinya waktu; digit yang bergoyang saat jam berjalan adalah cacat yang terlihat setiap detik.

Density: `comfortable` di layar absen (satu tugas, satu fokus), `compact` di tabel admin (banyak baris, layar lebar).

### 10.3 Layar absen — komposisi

Satu kolom terpusat, lebar maks 720px, di dalam viewport penuh tanpa sidebar.

```
        Selamat pagi, Budi Santoso
        Senin, 7 September 2026

              07:42:19            ← 64px, tabular-nums, elemen terbesar

        ┌──────────────────────┐
        │                      │   ← preview kamera 4:3
        │     [ wajah ]        │     ring frame = indikator state
        │                      │
        └──────────────────────┘
          Wajah terdeteksi ✓        ← satu baris status, berubah

        [      Absen Masuk      ]   ← satu tombol, lebar penuh kolom

        Masuk 07:42  ·  Pulang —  ·  Durasi —
```

**Ring frame kamera adalah indikator state utama**, bukan teks:

| State | Ring | Teks |
|---|---|---|
| Mencari wajah | abu, statis | "Posisikan wajah di dalam bingkai" |
| Wajah terdeteksi | biru, ring menebal | "Wajah terdeteksi — jangan bergerak" |
| Mengirim | biru, sweep berputar | "Memverifikasi…" |
| Berhasil | hijau, sekali pulse | "Budi Santoso · 07:42 · Tepat waktu" |
| Gagal | merah | pesan spesifik (lihat 10.5) |

Satu tempat untuk melihat state berarti mata tidak berpindah antara kamera dan teks di detik yang menentukan.

### 10.4 Layar admin

Layout AntD standar: `Layout.Sider` (menu) + `Layout.Header` (breadcrumb, user) + konten. Tidak ada penemuan navigasi baru — di Operate mode, konsistensi mengalahkan kejutan.

- Tabel absensi: `a-table` compact, kolom status pakai `a-tag` berwarna semantik, sticky header, pagination server-side.
- Filter di atas tabel dalam satu baris `a-form layout="inline"`, tidak di dalam drawer. Filter yang tersembunyi tidak dipakai.
- Monitoring hari ini: 4 statistik ringkas (hadir / terlambat / belum absen / izin) lalu tabelnya. Statistik ini bukan hiasan — angka "belum absen" jam 09:00 adalah kerjaan HR hari itu.
- Export: tombol di kanan atas tabel, mewarisi filter yang aktif.

### 10.5 State yang wajib ada (bukan opsional)

Ini bagian yang paling sering dilewat dan paling sering jadi keluhan:

- **Izin kamera ditolak** — halaman harus menjelaskan cara mengembalikannya di Chrome/Firefox/Safari, bukan sekadar "kamera tidak tersedia".
- **Tidak ada webcam terdeteksi** — arahkan ke jalur koreksi manual, jangan buntu.
- **Terlalu gelap / wajah terlalu kecil / lebih dari satu wajah** — pesan spesifik per kasus, dicek di klien sebelum submit supaya tidak buang trip ke server.
- **Wajah tidak dikenali** — hitung percobaan; setelah 3x tawarkan tombol "Ajukan koreksi manual". Jangan biarkan orang mencoba tanpa akhir.
- **Liveness gagal** — "Gunakan wajah asli, bukan foto atau layar." Jelas, tanpa menuduh.
- **Di luar jaringan kantor** — sebut alasannya dan siapa yang harus dihubungi.
- **Sudah absen hari ini** — tampilkan record yang ada, bukan error.
- **Loading** — skeleton di tabel, bukan spinner di tengah konten.
- **Empty state** — HR yang belum punya karyawan melihat ajakan import CSV, bukan "Tidak ada data".

### 10.6 Motion

150–250 ms untuk semua transisi. Motion hanya menyampaikan state: perubahan ring frame, sweep verifikasi, pulse keberhasilan. Tidak ada animasi masuk halaman, tidak ada hover effect dekoratif. Pengguna sedang dalam alur; mereka tidak mau menunggu koreografi.

### 10.7 Aksesibilitas & i18n

- Setiap perubahan state kamera diumumkan lewat `aria-live="polite"` — indikator warna saja tidak cukup, dan ini satu-satunya cara pengguna screen reader tahu absennya berhasil.
- Kontras teks minimal 4.5:1; status jangan hanya dibedakan warna (tag AntD punya teks, pertahankan).
- Seluruh alur absen bisa diselesaikan dengan keyboard.
- Semua string lewat `vue-i18n` sejak awal, meski hanya ada `id`. Menambahkan i18n setelah 40 komponen jadi adalah pekerjaan seminggu; melakukannya di hari pertama gratis.

### 10.8 Dependency frontend

`vue`, `vue-router`, `pinia`, `ant-design-vue`, `@ant-design/icons-vue`, `axios`, `vue-i18n`, `@mediapipe/tasks-vision`, `dayjs` (sudah jadi dependensi AntD — jangan tambah date library lain).

Tidak ada Tailwind. AntD sudah membawa sistem spacing dan token sendiri; menumpuk dua sistem tata letak menghasilkan dua sumber kebenaran untuk pertanyaan yang sama.

---

## 11. Deployment — VPS Tanpa Docker

Target: 1 VPS Ubuntu 22.04/24.04, 4 vCPU / 8 GB / 80 GB SSD.

### 11.1 Komponen

| Komponen | Cara jalan | Port |
|---|---|---|
| Caddy | `systemd` (paket resmi) | 80, 443 |
| Go API | `systemd` unit, binary di `/opt/hris/api` | 127.0.0.1:8080 |
| Python face service | `systemd` unit, `uvicorn` dalam venv `/opt/hris/face/.venv` | 127.0.0.1:8000 |
| PostgreSQL 16 + pgvector | `apt` (`postgresql-16-pgvector`) | 127.0.0.1:5432 |
| SPA | file statis hasil `vite build` di `/var/www/hris` | — |

Hanya Caddy yang mendengarkan di interface publik. Semua service lain bind ke `127.0.0.1`.

### 11.2 Caddy

Caddy dipilih daripada nginx karena HTTPS otomatis dalam tiga baris, dan HTTPS bukan opsional di sini — `getUserMedia` mati tanpa secure context, jadi seluruh produk mati tanpa TLS.

```
hris.perusahaan.com {
    root * /var/www/hris
    handle /api/* { reverse_proxy 127.0.0.1:8080 }
    handle { try_files {path} /index.html; file_server }
    encode gzip zstd
}
```

### 11.3 systemd

Dua unit, pola yang sama: `User=hris`, `Restart=always`, `EnvironmentFile=/etc/hris/api.env` (mode 600, berisi DSN dan JWT secret). Hardening minimum yang sepadan: `NoNewPrivileges=true`, `ProtectSystem=strict`, `ReadWritePaths=/var/lib/hris`.

Python service: `WorkingDirectory=/opt/hris/face`, `ExecStart=/opt/hris/face/.venv/bin/uvicorn main:app --host 127.0.0.1 --port 8000 --workers 2`. Model ONNX diunduh sekali saat provisioning ke `/opt/hris/face/models`, tidak saat runtime.

### 11.4 Deploy

Satu script `deploy.sh` di repo:
```
build   → GOOS=linux go build, vite build, pip install -r requirements.txt
upload  → rsync binary + dist/ + kode python ke VPS
migrate → jalankan migrasi SQL
restart → systemctl restart hris-api hris-face && reload caddy
```
Migrasi database: file SQL bernomor dijalankan oleh binary Go saat start (`golang-migrate` sebagai library). Tidak perlu tool migrasi terpisah untuk skala ini.

### 11.5 Backup & operasional

- `pg_dump` harian via cron ke `/var/backups/hris`, retensi 14 hari, sinkron ke object storage eksternal (backup yang hanya ada di VPS yang sama bukan backup).
- Foto bukti: retensi 30 hari, cron pembersih harian.
- Log: `journalctl` + `logrotate` bawaan; belum perlu stack logging terpisah.
- Monitoring rilis 1: endpoint `/healthz` (cek DB + face service) dipantau uptime checker gratis. Alert ke email/WhatsApp kalau mati — jam 07:00 adalah jam yang tidak boleh mati.

### 11.6 Risiko VPS satu mesin

Satu VPS berarti satu titik kegagalan. Untuk < 200 karyawan ini trade-off yang benar, dengan syarat backup off-site jalan dan ada **jalur cadangan absen manual** yang bisa dipakai HR kalau server mati di jam masuk. Tanpa jalur itu, mati 30 menit di pagi hari = data kehadiran satu hari hilang.

---

## 12. Risiko Produk

| Risiko | Mitigasi |
|---|---|
| Kembar identik / saudara mirip | Threshold ketat + flag manual review |
| Wajah berubah (jenggot, kacamata, berat badan) | Re-enroll mudah; multi-embedding per orang |
| Webcam laptop murah / pencahayaan buruk | Cek kualitas di klien sebelum submit + panduan pencahayaan di UI |
| Spoof video HD | Passive liveness + challenge; foto bukti disimpan untuk audit |
| Karyawan tanpa webcam | Jalur koreksi manual + approval HR; data dicatat siapa yang sering pakai jalur ini |
| Regulasi biometrik (UU PDP) | Consent tercatat, retensi jelas, hak hapus saat resign |
| Karyawan merasa diawasi | Transparansi di UI: sebutkan apa yang disimpan dan berapa lama, di halaman enrollment |

## 13. Rencana Rilis

**Fase 1 — MVP (4–6 minggu)**
Auth + RBAC, CRUD karyawan/jadwal, enrollment, check-in/out + anti-spoof, IP allowlist, dashboard HR dasar, export CSV, deploy VPS.

**Fase 2 — Operasional (3–4 minggu)**
Koreksi + approval, device binding, notifikasi email, dashboard manajer, export Excel, audit log, empty & error state lengkap.

**Fase 3 — Opsional**
Mobile web (kalau ternyata dibutuhkan), lembur & shift, integrasi payroll, dark mode, kalibrasi threshold otomatis dari data produksi.

## 14. Pertanyaan Terbuka

Sudah terjawab: skala (< 200, 1–2 lokasi), device (laptop/desktop), karakter visual (kalem & institusional), dark mode (tidak di rilis 1).

Masih terbuka:
1. Foto bukti tiap absen disimpan atau tidak? (privasi vs audit — mempengaruhi F4 dan retensi)
2. WFH diizinkan? Kalau ya, berapa persen karyawan yang perlu flag `allow_remote`?
3. Kantor punya IP publik statis? Kalau dinamis, validasi jaringan harus pakai cara lain.
4. Ada sistem payroll existing yang harus menerima data ini?
5. Domain dan VPS sudah ada, atau perlu diadakan?
