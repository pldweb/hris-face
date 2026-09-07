---
version: 1
slug: "web-src-views-checkin-vue"
primary_target: "web/src/views/CheckIn.vue"
related_targets: []
---

Scope: layar Check-in/Check-out karyawan (web/src/views/CheckIn.vue). Mode: Operate.

Audience: karyawan tetap, 2x sehari, di meja kerja, laptop, indoor.
Task: buktikan ini benar saya, lalu selesai. Target < 5 detik.
Constraints: Ant Design Vue (pinned), light only, desktop only, bahasa Indonesia, verifikasi selalu di server.

## Direction contract

THESIS: Layar absen adalah satu viewport penuh yang isinya kamera dan jam — bukan kartu widget di dalam grid dashboard. Menolak default kategori "Attendance Card" yang menyempil di antara enam kartu statistik.

OWN-WORLD: Ant Design light dengan palet dipersempit — ground netral abu-dingin (#F5F6F8), permukaan putih, satu aksen biru institusional sebagai colorPrimary. Hijau/kuning/merah AntD dikunci hanya untuk status kehadiran, tidak pernah dekorasi. System sans stack; semua angka waktu pakai font-variant-numeric: tabular-nums. Radius 6px, border 1px #E5E7EB, shadow hanya pada overlay. Kepadatan: comfortable di layar absen, compact di tabel admin.

STORY: Karyawan buka halaman, langsung lihat wajahnya sendiri dan jam berjalan, tahu ia belum absen, menghadap kamera, dan pergi dengan bukti tercatat jam berapa dan statusnya apa.

FIRST VIEWPORT: Satu kolom terpusat, lebar maks 720px. Atas: sapaan + tanggal, jam berjalan besar (64px, tabular-nums) sebagai elemen terbesar di layar. Tengah: preview kamera 4:3 dengan frame panduan wajah; ring frame adalah indikator state (abu = mencari, biru = terdeteksi, hijau = terverifikasi, merah = gagal). Bawah frame: satu baris status teks yang berubah. Aksi primer satu tombol lebar di bawah kamera. Di kanan/bawah: ringkasan hari ini (masuk, pulang, durasi) — read-only.

FORM: Composition dipilih langsung oleh user dari preview terminal (opsi "Kalem & institusional"); brief-pinned, tidak lewat concept-seed. Seed key: n/a (pinned brief).

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance.
