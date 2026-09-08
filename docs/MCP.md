# MCP Server — HRIS Face Attendance

Server MCP (Model Context Protocol) yang membuka data HRIS ke asisten AI seperti
Claude Desktop atau Claude Code, supaya bisa ditanyai hal seperti:

> "Berapa orang yang telat hari ini?"
> "Siapa saja yang cutinya masih menunggu persetujuan?"
> "Sisa cuti Budi berapa?"
> "Rekap kehadiran divisi Produksi bulan ini dong"

## Sifatnya: hanya membaca

Server ini **tidak punya satu pun tool yang bisa mengubah data.** HTTP client di
dalamnya cuma mengimplementasikan `GET` — tidak ada `POST`, `PUT`, `PATCH`, atau
`DELETE` sama sekali. Jadi kalaupun AI salah menafsirkan perintah, kemungkinan
terburuknya adalah salah menjawab, bukan menyetujui cuti orang atau mengubah
absensi.

Datanya diambil lewat REST API yang sudah ada, bukan query database langsung.
Konsekuensinya: role check dan logika laporan tetap satu sumber di API, dan
server MCP tidak bisa diam-diam melewati izin yang diberlakukan API.

## Build

```bash
cd api
go build -o bin/mcp ./cmd/mcp
```

## Konfigurasi

Tiga environment variable:

| Variable | Wajib | Keterangan |
|---|---|---|
| `HRIS_MCP_EMAIL` | ya | Akun HR atau superadmin |
| `HRIS_MCP_PASSWORD` | ya | Password akun tersebut |
| `HRIS_API_URL` | tidak | Default `http://127.0.0.1:8080` |

Server login sendiri ke API dan memperbarui token saat kedaluwarsa (access token
umurnya 15 menit), jadi tidak perlu menyiapkan token manual.

### Claude Code

```bash
claude mcp add hris \
  --env HRIS_MCP_EMAIL=hr@perusahaan.com \
  --env HRIS_MCP_PASSWORD='rahasia-sekali-123' \
  --env HRIS_API_URL=http://127.0.0.1:8080 \
  -- /path/ke/hris-face/api/bin/mcp
```

### Claude Desktop

Tambahkan ke `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "hris": {
      "command": "/path/ke/hris-face/api/bin/mcp",
      "env": {
        "HRIS_API_URL": "http://127.0.0.1:8080",
        "HRIS_MCP_EMAIL": "hr@perusahaan.com",
        "HRIS_MCP_PASSWORD": "rahasia-sekali-123"
      }
    }
  }
}
```

Pakai **path absolut** ke binary. Setelah itu restart Claude Desktop.

## Tool yang tersedia

| Tool | Kegunaan |
|---|---|
| `attendance_today` | Ringkasan hari ini: aktif, tepat waktu, telat, belum absen, sudah pulang, sedang cuti |
| `list_attendance` | Catatan kehadiran per rentang tanggal, bisa disaring status & departemen. Default: awal bulan s.d. hari ini |
| `list_employees` | Daftar karyawan + departemen, lokasi, status, kuota cuti. Bisa disaring/dicari |
| `list_leave_requests` | Pengajuan cuti/izin/sakit, bisa disaring status & jenis |
| `leave_balance` | Kuota, terpakai, dan sisa cuti tahunan seorang karyawan (cari per nama atau NIK) |
| `list_corrections` | Pengajuan koreksi absen |
| `list_master_data` | Departemen, lokasi kerja, jadwal kerja |

Semua tanggal memakai format `YYYY-MM-DD`, zona waktu Asia/Jakarta.

Id internal (uuid) diterjemahkan jadi nama yang bisa dibaca di tempat yang
memungkinkan, supaya jawaban AI tidak berisi uuid yang tidak berarti bagi orang.

## Catatan operasional

- **Stdio, bukan network.** Server jalan sebagai proses lokal yang bicara lewat
  stdin/stdout. Tidak ada port yang dibuka, jadi tidak menambah permukaan
  serangan di VPS.
- **Semua log ke stderr.** Stdout dipakai untuk protokol JSON-RPC; menulis apa
  pun ke sana akan merusak sesi.
- **API harus jalan.** Server MCP adalah client dari API HRIS. Kalau API mati,
  tool akan mengembalikan pesan error yang menyebutkan alamat yang gagal
  dihubungi, bukan diam-diam mengembalikan data kosong.
- **Nama ambigu ditolak, bukan ditebak.** Kalau "Budi" cocok dengan dua orang,
  `leave_balance` mengembalikan error berisi daftar NIK yang cocok. Menebak
  salah satu berarti menjawab pertanyaan tentang orang yang keliru.

## Tes

```bash
python3 test/e2e/test_mcp.py
```

Skrip ini bicara JSON-RPC sungguhan ke binary (handshake, `tools/list`,
`tools/call`), termasuk memastikan tidak ada tool yang bersifat mengubah data
dan bahwa kredensial salah menghasilkan error yang jelas, bukan crash.
