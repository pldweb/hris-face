# PRODUCT.md — HRIS Absensi Face Recognition

## Apa ini
Web app absensi karyawan yang memverifikasi identitas lewat wajah dari webcam laptop/desktop. Menggantikan mesin fingerprint dan rekap manual.

## Mekanisme unik
Karyawan membuktikan kehadirannya dengan wajahnya sendiri, di depan laptop kerjanya, dalam hitungan detik — tanpa kartu, tanpa PIN, tanpa antre di mesin.

## Pengguna & scene nyata
- **Karyawan** (< 200 orang, 1–2 lokasi). Duduk di meja, laptop kantor, pencahayaan indoor. Absen 2x sehari, ± 30 detik total. Frekuensi tinggi, perhatian rendah — ini rutinitas, bukan tugas.
- **HR/Admin.** Layar lebar, monitoring pagi hari (jam 07:00–09:00) dan tutup bulan. Butuh kepadatan data dan export.
- **Manajer.** Sesekali, cek tim.

## Batasan yang mengikat
- **Desktop/laptop only.** Tidak ada mobile web, tidak ada kios tablet di rilis 1.
- **Komitmen brand: Ant Design Vue** sebagai bahasa komponen. Bukan preferensi — sudah dipilih.
- **Light mode saja** di rilis 1. Token disiapkan untuk dark, tapi tidak diimplementasi.
- **Karakter: kalem & institusional.** Terasa seperti sistem perusahaan resmi. Ini alat kerja, bukan produk konsumen.
- **Bahasa Indonesia** di seluruh UI.
- **Deploy: satu VPS, tanpa Docker.** systemd + nginx/Caddy.

## Yang bikin hasil poles terasa salah
Absensi wajah gampang terasa seperti pengawasan. Kalau UI-nya main-main (emoji, animasi lucu, copy kasual) itu terasa meremehkan sesuatu yang menyangkut gaji orang. Kalau UI-nya dingin dan penuh peringatan, terasa seperti alat curiga. Nada yang benar: tenang, tegas, dan jelas soal apa yang direkam.

## Bukan tujuan (rilis 1)
Payroll, cuti/approval workflow, mobile app, shift rotasi, integrasi mesin fingerprint lama.
