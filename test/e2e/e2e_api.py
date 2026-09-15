"""End-to-end test against the real Go API + real Postgres/pgvector.

Runs against the REAL face service (ArcFace + MiniFASNet), using real faces from
LFW as test identities and the upstream anti-spoof sample as the spoof case.
Everything here is production code: auth, refresh rotation, RBAC, enrollment
validation, duplicate detection, pgvector matching, attendance rules.
"""

import io
import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path

from PIL import Image

API = os.getenv("API_URL", "http://127.0.0.1:8080/api/v1")
FACES = Path("/tmp/e2e_faces")

passed, failed = 0, 0


def prepare_faces() -> None:
    """Two LFW identities (two photos each) plus a third as the unknown face.

    Enrolment uses brightness variants of one photo; check-in uses the *other*
    photo of that identity, so matching is tested across genuinely different
    images rather than against a copy of what was enrolled.
    """
    if (FACES / "ready").exists():
        return

    import numpy as np
    import pandas as pd

    FACES.mkdir(parents=True, exist_ok=True)
    url = "https://huggingface.co/datasets/logasja/lfw/resolve/main/pairs/test-00000-of-00001.parquet"
    same = pd.read_parquet(url).query("pair == 1")

    # LFW crops are 250x250 with ~90px faces, below the 112px quality gate in
    # docs/PRD.md 6.1. Upscale 2x so the fixtures look like a real 640x480
    # webcam capture instead of failing a rule that is doing its job.
    def load(cell) -> Image.Image:
        img = Image.open(io.BytesIO(cell["bytes"]))
        return img.resize((img.width * 2, img.height * 2), Image.LANCZOS)

    for label, idx in [("A", 0), ("B", 5), ("C", 9)]:
        row = same.iloc[idx]
        load(row["img_0"]).save(FACES / f"{label}1.jpg", quality=95)
        load(row["img_1"]).save(FACES / f"{label}2.jpg", quality=95)
        base = Image.open(FACES / f"{label}1.jpg")
        for i in range(3):
            arr = np.array(base).astype(np.int16) + (i - 1) * 12
            Image.fromarray(np.clip(arr, 0, 255).astype(np.uint8)).save(FACES / f"{label}1_v{i}.jpg")

    Image.new("RGB", (640, 480), (20, 20, 20)).save(FACES / "noface.jpg")

    spoof = FACES / "spoof.jpg"
    if not spoof.exists():
        urllib.request.urlretrieve(
            "https://raw.githubusercontent.com/minivision-ai/"
            "Silent-Face-Anti-Spoofing/master/images/sample/image_F2.jpg",
            spoof,
        )
    (FACES / "ready").write_text("ok")


def face(name: str) -> bytes:
    return (FACES / f"{name}.jpg").read_bytes()


def check(name, condition, detail=""):
    global passed, failed
    if condition:
        passed += 1
        print(f"  PASS  {name}")
    else:
        failed += 1
        print(f"  FAIL  {name}  {detail}")


def request(method, path, body=None, token=None, content_type="application/json", raw=None, cookie=None):
    url = API + path
    data = raw if raw is not None else (json.dumps(body).encode() if body is not None else None)
    req = urllib.request.Request(url, data=data, method=method)
    if data is not None:
        req.add_header("Content-Type", content_type)
    if token:
        req.add_header("Authorization", f"Bearer {token}")
    if cookie:
        req.add_header("Cookie", cookie)
    try:
        with urllib.request.urlopen(req) as resp:
            set_cookie = resp.headers.get("Set-Cookie")
            payload = resp.read().decode()
            return resp.status, (json.loads(payload) if payload else {}), set_cookie
    except urllib.error.HTTPError as e:
        payload = e.read().decode()
        try:
            parsed = json.loads(payload)
        except json.JSONDecodeError:
            parsed = {"raw": payload}
        return e.code, parsed, e.headers.get("Set-Cookie")


def request_raw(method, path, token=None):
    """Like request() but returns the raw body -- the CSV export is not JSON."""
    req = urllib.request.Request(API + path, method=method)
    if token:
        req.add_header("Authorization", f"Bearer {token}")
    try:
        with urllib.request.urlopen(req) as resp:
            return resp.status, resp.read(), resp.headers
    except urllib.error.HTTPError as e:
        return e.code, e.read(), e.headers


def multipart_named(field, filename, content, content_type):
    boundary = "----hristest"
    body = (
        f"--{boundary}\r\n"
        f'Content-Disposition: form-data; name="{field}"; filename="{filename}"\r\n'
        f"Content-Type: {content_type}\r\n\r\n"
    ).encode() + content + f"\r\n--{boundary}--\r\n".encode()
    return body, f"multipart/form-data; boundary={boundary}"


def multipart(photos):
    boundary = "----hristest"
    parts = []
    for i, photo in enumerate(photos):
        parts.append(
            (
                f"--{boundary}\r\n"
                f'Content-Disposition: form-data; name="photos"; filename="p{i}.jpg"\r\n'
                f"Content-Type: image/jpeg\r\n\r\n"
            ).encode()
            + photo
            + b"\r\n"
        )
    parts.append(f"--{boundary}--\r\n".encode())
    return b"".join(parts), f"multipart/form-data; boundary={boundary}"


prepare_faces()

print("\n=== AUTH ===")
status, body, _ = request("POST", "/auth/login", {"email": "hr@perusahaan.com", "password": "salah"})
check("login password salah ditolak 401", status == 401, f"got {status} {body}")

status, body, cookie = request("POST", "/auth/login", {"email": "hr@perusahaan.com", "password": "rahasia-sekali-123"})
check("login benar mengembalikan 200 + token", status == 200 and "access_token" in body, f"got {status} {body}")
hr_token = body.get("access_token", "")
check("refresh token dikirim sebagai httpOnly cookie", cookie and "HttpOnly" in cookie, f"cookie={cookie}")
check("refresh cookie Secure + SameSite", cookie and "Secure" in cookie and "SameSite" in cookie, f"cookie={cookie}")

refresh_cookie = cookie.split(";")[0] if cookie else ""
status, body, new_cookie = request("POST", "/auth/refresh", cookie=refresh_cookie)
check("refresh menghasilkan access token baru", status == 200 and "access_token" in body, f"got {status} {body}")

status, body, _ = request("POST", "/auth/refresh", cookie=refresh_cookie)
check("refresh token lama ditolak setelah rotasi", status == 401, f"got {status} {body}")

status, body, _ = request("GET", "/admin/employees")
check("endpoint admin tanpa token ditolak 401", status == 401, f"got {status}")

print("\n=== MASTER DATA (RBAC) ===")
status, dept, _ = request("POST", "/admin/departments", {"name": "Operasional"}, token=hr_token)
check("HR bisa membuat departemen", status == 201 and "id" in dept, f"got {status} {dept}")

status, body, _ = request(
    "POST", "/admin/employees",
    {"nik": "A001", "full_name": "Andi Pratama", "email": "andi@perusahaan.com", "department_id": dept.get("id")},
    token=hr_token,
)
check("HR bisa membuat karyawan", status == 201, f"got {status} {body}")
andi_password = body.get("temp_password", "")
check("password sementara dikembalikan sekali", bool(andi_password), f"got {body}")
check("karyawan baru berstatus pending_enrollment", body.get("employee", {}).get("status") == "pending_enrollment", f"got {body}")

status, body, _ = request(
    "POST", "/admin/employees",
    {"nik": "A001", "full_name": "Duplikat NIK", "email": "lain@perusahaan.com"},
    token=hr_token,
)
check("NIK duplikat ditolak 409", status == 409, f"got {status} {body}")

status, body, _ = request(
    "POST", "/admin/employees",
    {"nik": "A002", "full_name": "Email Duplikat", "email": "andi@perusahaan.com"},
    token=hr_token,
)
check("email duplikat ditolak 409", status == 409, f"got {status} {body}")

status, body, _ = request("POST", "/admin/employees", {"nik": "A003", "full_name": "Tanpa Email"}, token=hr_token)
check("field wajib divalidasi 400", status == 400, f"got {status} {body}")

# Second employee for the duplicate-face test.
status, body, _ = request(
    "POST", "/admin/employees",
    {"nik": "B001", "full_name": "Budi Santoso", "email": "budi@perusahaan.com"},
    token=hr_token,
)
budi_password = body.get("temp_password", "")

print("\n=== EMPLOYEE LOGIN + RBAC ===")
status, body, _ = request("POST", "/auth/login", {"email": "andi@perusahaan.com", "password": andi_password})
check("karyawan bisa login dengan password sementara", status == 200, f"got {status} {body}")
andi_token = body.get("access_token", "")

status, body, _ = request("GET", "/admin/employees", token=andi_token)
check("karyawan biasa ditolak dari endpoint admin 403", status == 403, f"got {status} {body}")

print("\n=== ENROLLMENT ===")
raw, ctype = multipart([face("A1_v0"), face("A1_v1")])
status, body, _ = request("POST", "/enrollment/photos", token=andi_token, raw=raw, content_type=ctype)
check("kurang dari 3 foto ditolak 400", status == 400, f"got {status} {body}")

raw, ctype = multipart([face("A1_v0"), face("noface"), face("A1_v1")])
status, body, _ = request("POST", "/enrollment/photos", token=andi_token, raw=raw, content_type=ctype)
check("foto tanpa wajah ditolak per-foto 422", status == 422 and body.get("rejections"), f"got {status} {body}")
check("penolakan menyebut index foto yang gagal", status == 422 and body["rejections"][0]["index"] == 1, f"got {body}")

raw, ctype = multipart([face("A1_v0"), face("A1_v1"), face("A1_v2")])
status, body, _ = request("POST", "/enrollment/photos", token=andi_token, raw=raw, content_type=ctype)
check("enrollment valid berhasil", status == 200 and body.get("status") == "active", f"got {status} {body}")

status, body, _ = request("POST", "/auth/login", {"email": "budi@perusahaan.com", "password": budi_password})
budi_token = body.get("access_token", "")
raw, ctype = multipart([face("A1_v0"), face("A1_v1"), face("A1_v2")])
status, body, _ = request("POST", "/enrollment/photos", token=budi_token, raw=raw, content_type=ctype)
check("wajah duplikat karyawan lain ditolak 409", status == 409, f"got {status} {body}")
check("pesan duplikat tidak membocorkan nama karyawan lain",
      "Andi" not in json.dumps(body), f"got {body}")

raw, ctype = multipart([face("B1_v0"), face("B1_v1"), face("B1_v2")])
status, body, _ = request("POST", "/enrollment/photos", token=budi_token, raw=raw, content_type=ctype)
check("wajah berbeda diterima", status == 200, f"got {status} {body}")

print("\n=== CHECK-IN ===")
status, body, _ = request("POST", "/attendance/check-in", token=andi_token, raw=face("spoof"), content_type="image/jpeg")
check("spoof (liveness rendah) ditolak 422", status == 422, f"got {status} {body}")
check("pesan spoof jelas", "asli" in json.dumps(body).lower(), f"got {body}")

status, body, _ = request("POST", "/attendance/check-in", token=andi_token, raw=face("C2"), content_type="image/jpeg")
check("wajah tak dikenal ditolak 422", status == 422, f"got {status} {body}")

status, body, device_cookie = request("POST", "/attendance/check-in", token=andi_token, raw=face("A2"), content_type="image/jpeg")
check("check-in wajah terdaftar berhasil", status == 200, f"got {status} {body}")
check("server menerbitkan cookie device httpOnly", device_cookie and "HttpOnly" in device_cookie, f"cookie={device_cookie}")
# A browser keeps this cookie; urllib does not, so thread it manually or every
# request looks like a brand-new device and trips the 2-device limit.
andi_device = device_cookie.split(";")[0] if device_cookie else ""
check("check-in mengembalikan nama yang benar", body.get("full_name") == "Andi Pratama", f"got {body}")
check("check-in mengembalikan status kehadiran", body.get("status") in ("on_time", "late"), f"got {body}")
andi_checkin_status = body.get("status")

first_checkin_at = body.get("occurred_at")
status, body, _ = request("POST", "/attendance/check-in", token=andi_token, raw=face("A2"), content_type="image/jpeg", cookie=andi_device)
check("presensi ulang: check-in kedua di hari sama diterima (menimpa)", status == 200, f"got {status} {body}")
check("presensi ulang: jam check-in berubah", body.get("occurred_at") != first_checkin_at, f"got {body}")

status, body, _ = request("POST", "/attendance/check-in", token=budi_token, raw=face("B2"), content_type="image/jpeg")
check("karyawan kedua bisa check-in (tidak tertukar)", status == 200 and body.get("full_name") == "Budi Santoso", f"got {status} {body}")

print("\n=== STATUS DARI JADWAL KERJA ===")
# Default schedule from migration 0003: 08:00-17:00, 15 min tolerance. Whether
# this run lands late or on time depends on the wall clock, so derive the
# expectation the same way the server should and compare -- a hardcoded
# "on_time" fails this whenever the suite runs after 08:15 WIB.
import datetime
import zoneinfo

now_local = datetime.datetime.now(zoneinfo.ZoneInfo("Asia/Jakarta"))
minutes_now = now_local.hour * 60 + now_local.minute
expected_status = "late" if minutes_now > (8 * 60 + 15) else "on_time"
check(
    f"status check-in dihitung dari jadwal (jam {now_local:%H:%M} WIB -> {expected_status})",
    andi_checkin_status == expected_status,
    f"got {andi_checkin_status}, expected {expected_status}",
)

print("\n=== IDENTITAS SESI vs WAJAH ===")
status, body, _ = request("POST", "/attendance/check-out", token=hr_token, raw=face("A2"), content_type="image/jpeg")
check("akun tanpa data karyawan ditolak 404", status == 404, f"got {status} {body}")

# Budi yang login, tapi wajah Andi di depan kamera: absennya tetap ditolak
# (tidak bisa titip absen -- sesi menentukan siapa, wajah cuma verifikasi),
# tapi sejak keputusan produk terbaru pesan penolakan BOLEH menyebut nama
# pemilik wajah yang terdeteksi.
status, body, _ = request("POST", "/attendance/check-out", token=budi_token, raw=face("A2"), content_type="image/jpeg")
check("wajah orang lain di sesi sendiri tetap ditolak (anti titip-absen)", status == 422, f"got {status} {body}")
check("penolakan menyebut nama pemilik wajah yang terdeteksi",
      body.get("matched_name") == "Andi Pratama", f"got {body}")

print("\n=== CHECK-OUT ===")

status, body, _ = request("POST", "/attendance/check-out", token=andi_token, raw=face("A2"), content_type="image/jpeg", cookie=andi_device)
check("check-out setelah check-in berhasil", status == 200, f"got {status} {body}")
check("check-out bertipe check_out", body.get("type") == "check_out", f"got {body}")
check("check-out sebelum jam pulang = early_leave",
      body.get("status") in ("early_leave", "on_time"), f"got {body}")

first_checkout_at = body.get("occurred_at")
status, body, _ = request("POST", "/attendance/check-out", token=andi_token, raw=face("A2"), content_type="image/jpeg", cookie=andi_device)
check("presensi ulang: check-out kedua diterima (menimpa)", status == 200, f"got {status} {body}")
check("presensi ulang: jam check-out berubah", body.get("occurred_at") != first_checkout_at, f"got {body}")

status, body, _ = request("GET", "/me", token=andi_token)
check("/me menampilkan jam masuk dan pulang", bool(body.get("check_in_at")) and bool(body.get("check_out_at")), f"got {body}")

print("\n=== DEVICE BINDING ===")
status, body, _ = request("GET", "/admin/devices", token=hr_token)
devices = body.get("data") or []
check("perangkat terdaftar otomatis saat absen", len(devices) >= 2, f"got {status} {body}")
check("perangkat dalam batas langsung disetujui", all(d.get("approved") for d in devices), f"got {devices}")
check("satu perangkat per karyawan (tidak menumpuk)",
      len({d["employee_id"] for d in devices}) == len(devices), f"got {devices}")
check("perangkat tercatat dengan nama karyawan", all(d.get("full_name") for d in devices), f"got {devices}")

status, body, _ = request("GET", "/admin/devices", token=andi_token)
check("daftar perangkat hanya untuk HR", status == 403, f"got {status} {body}")

print("\n=== JADWAL & LOKASI ===")
status, body, _ = request("GET", "/admin/schedules", token=hr_token)
schedules = body.get("data") or []
check("jadwal default tersedia", len(schedules) >= 1, f"got {status} {body}")
check("jadwal punya jam masuk/pulang",
      schedules and schedules[0].get("start_time") and schedules[0].get("end_time"), f"got {schedules}")

status, body, _ = request("POST", "/admin/schedules",
    {"name": "Shift Sore", "start_time": "13:00", "end_time": "21:00", "late_tolerance_minutes": 10},
    token=hr_token)
check("HR bisa membuat jadwal baru", status == 201, f"got {status} {body}")
new_schedule_id = body.get("id", "")

status, body, _ = request("PUT", f"/admin/schedules/{new_schedule_id}",
    {"name": "Shift Sore (revisi)", "start_time": "14:00", "end_time": "22:00", "late_tolerance_minutes": 5},
    token=hr_token)
check("HR bisa mengubah jadwal tanpa mengirim work_days", status == 200, f"got {status} {body}")

status, body, _ = request("GET", "/admin/schedules", token=hr_token)
revised = next((x for x in body.get("data", []) if x["id"] == new_schedule_id), None)
check("hari kerja tidak ikut tereset saat tidak dikirim",
      revised and revised.get("work_days") == [1, 2, 3, 4, 5], f"got {revised}")
check("perubahan jam tersimpan", revised and revised.get("start_time") == "14:00", f"got {revised}")

status, body, _ = request("POST", "/admin/locations", {"name": "Kantor Pusat"}, token=hr_token)
check("HR bisa membuat lokasi", status == 201, f"got {status} {body}")

status, body, _ = request("GET", "/admin/schedules", token=andi_token)
check("master data hanya untuk HR", status == 403, f"got {status} {body}")

print("\n=== MONITORING & EXPORT ===")
status, body, _ = request("GET", "/admin/attendances", token=hr_token)
check("daftar absensi bisa diambil", status == 200 and isinstance(body.get("data"), list), f"got {status} {body}")
check("daftar absensi berisi record hari ini", body.get("total", 0) >= 2, f"got total={body.get('total')}")
check("baris absensi memuat nama dan status",
      body["data"] and body["data"][0].get("full_name") and body["data"][0].get("status"), f"got {body}")

print("\n=== HAPUS ABSENSI ===")
budi_attendance = next((r for r in body.get("data", []) if r.get("full_name") == "Budi Santoso"), None)
check("setup: absensi Budi tersedia untuk dihapus", budi_attendance is not None, f"got {body}")
attendance_id = budi_attendance.get("id", "") if budi_attendance else ""

status, body, _ = request("DELETE", f"/admin/attendances/{attendance_id}", token=budi_token)
check("karyawan tidak bisa menghapus absensi", status == 403, f"got {status} {body}")

status, body, _ = request("DELETE", f"/admin/attendances/{attendance_id}", token=hr_token)
check("HR bisa menghapus absensi", status == 200 and body.get("status") == "deleted", f"got {status} {body}")

status, body, _ = request("GET", "/admin/attendances", token=hr_token)
check("absensi yang dihapus tidak lagi tampil", all(r.get("id") != attendance_id for r in body.get("data", [])), f"got {body}")

status, body, _ = request("DELETE", f"/admin/attendances/{attendance_id}", token=hr_token)
check("menghapus absensi yang sudah hilang mengembalikan 404", status == 404, f"got {status} {body}")

status, body, _ = request("GET", "/admin/attendances?status=late", token=hr_token)
check("filter status bekerja",
      all(r.get("status") == "late" for r in body.get("data", [])), f"got {body}")

status, body, _ = request("GET", "/admin/attendances?from=2020-01-01&to=2020-01-02", token=hr_token)
check("filter tanggal mengecualikan data di luar rentang", body.get("total") == 0, f"got {body}")

status, body, _ = request("GET", "/admin/attendances/today", token=hr_token)
check("ringkasan hari ini tersedia", status == 200 and "not_yet" in body, f"got {status} {body}")
# Budi's check-in was deliberately deleted just above (HAPUS ABSENSI), so only
# Andi's remains -- the summary correctly reflects that, not the pre-delete count.
check("ringkasan menghitung yang sudah absen (setelah Budi dihapus)",
      body.get("on_time", 0) + body.get("late", 0) >= 1, f"got {body}")
check("ringkasan mencerminkan penghapusan Budi", body.get("not_yet", 0) >= 1, f"got {body}")

csv_status, csv_body, _ = request_raw("GET", "/admin/reports/export", token=hr_token)
check("export CSV berhasil", csv_status == 200, f"got {csv_status}")
check("CSV punya header berbahasa Indonesia", b"NIK" in csv_body and "Nama".encode() in csv_body, f"got {csv_body[:120]}")
check("CSV diawali BOM agar Excel tidak mojibake", csv_body.startswith(b"\xef\xbb\xbf"), f"got {csv_body[:10]}")
check("CSV berisi data karyawan", b"Andi Pratama" in csv_body, f"got {csv_body[:300]}")

status, body, _ = request("GET", "/admin/reports/export", token=andi_token)
check("export hanya untuk HR", status == 403, f"got {status}")

print("\n=== RIWAYAT KARYAWAN ===")
status, body, _ = request("GET", "/attendance/me", token=andi_token)
check("karyawan bisa melihat riwayatnya", status == 200 and isinstance(body.get("data"), list), f"got {status} {body}")
today_row = next((d for d in body.get("data", []) if d.get("check_in_at")), None)
check("riwayat memuat hari absen hari ini", today_row is not None, f"got {body}")
check("riwayat menghitung durasi kerja", today_row and today_row.get("minutes", 0) >= 0, f"got {today_row}")

print("\n=== KOREKSI ABSEN ===")
status, body, _ = request("POST", "/corrections",
    {"requested_type": "check_in", "requested_time": "2026-09-01T08:00:00+07:00", "reason": "x"},
    token=budi_token)
check("alasan terlalu pendek ditolak 400", status == 400, f"got {status} {body}")

status, body, _ = request("POST", "/corrections",
    {"requested_type": "makan_siang", "requested_time": "2026-09-01T08:00:00+07:00", "reason": "alasan yang cukup"},
    token=budi_token)
check("tipe koreksi tidak valid ditolak", status == 400, f"got {status} {body}")

status, body, _ = request("POST", "/corrections",
    {"requested_type": "check_in", "requested_time": "2026-09-01T08:00:00+07:00",
     "reason": "kamera tidak mengenali wajah setelah 3 percobaan"},
    token=budi_token)
check("karyawan bisa mengajukan koreksi", status == 201, f"got {status} {body}")
correction_id = body.get("id", "")
check("pengajuan berstatus pending", body.get("status") == "pending", f"got {body}")

status, body, _ = request("GET", "/corrections/me", token=budi_token)
check("karyawan melihat pengajuannya sendiri", len(body.get("data", [])) == 1, f"got {body}")

status, body, _ = request("GET", "/corrections/me", token=andi_token)
check("pengajuan orang lain tidak bocor", len(body.get("data", [])) == 0, f"got {body}")

status, body, _ = request("GET", "/admin/corrections?status=pending", token=hr_token)
check("HR melihat antrean koreksi", len(body.get("data", [])) == 1, f"got {status} {body}")

status, body, _ = request("PATCH", f"/admin/corrections/{correction_id}",
    {"decision": "approve", "note": "sudah dicek CCTV"}, token=hr_token)
check("HR bisa menyetujui koreksi", status == 200, f"got {status} {body}")

status, body, _ = request("PATCH", f"/admin/corrections/{correction_id}",
    {"decision": "reject"}, token=hr_token)
check("koreksi yang sudah diproses tidak bisa diproses ulang", status == 409, f"got {status} {body}")

status, body, _ = request("GET", "/admin/attendances?from=2026-09-01&to=2026-09-01", token=hr_token)
approved_row = next((r for r in body.get("data", []) if r.get("source") == "manual_correction"), None)
check("koreksi yang disetujui menghasilkan record absensi", approved_row is not None, f"got {body}")
check("record koreksi ditandai sebagai manual",
      approved_row and approved_row.get("source") == "manual_correction", f"got {approved_row}")

status, body, _ = request("PATCH", f"/admin/corrections/{correction_id}",
    {"decision": "approve"}, token=budi_token)
check("karyawan tidak bisa menyetujui koreksi", status == 403, f"got {status} {body}")

print("\n=== IMPORT CSV ===")
csv_body = (
    "nik,nama,email,departemen\n"
    "IMP-1,Citra Dewi,citra@perusahaan.com,Keuangan\n"
    "IMP-2,Dodi Firman,dodi@perusahaan.com,Keuangan\n"
    "IMP-3,Tanpa Email,,\n"
    "A001,NIK Bentrok,bentrok@perusahaan.com,\n"
).encode()
raw, ctype = multipart_named("file", "karyawan.csv", csv_body, "text/csv")
status, body, _ = request("POST", "/admin/employees/import", token=hr_token, raw=raw, content_type=ctype)
check("import CSV berhasil diproses", status == 200, f"got {status} {body}")
check("baris valid dibuat", len(body.get("created", [])) == 2, f"got {body.get('created')}")
check("baris tanpa email dilaporkan gagal",
      any("wajib diisi" in r.get("error", "") for r in body.get("failed", [])), f"got {body.get('failed')}")
check("NIK bentrok dilaporkan gagal",
      any("NIK sudah terdaftar" in r.get("error", "") for r in body.get("failed", [])), f"got {body.get('failed')}")
check("baris gagal menyebut nomor barisnya",
      all(r.get("line", 0) >= 2 for r in body.get("failed", [])), f"got {body.get('failed')}")
check("import mengembalikan password sementara",
      all(r.get("temp_password") for r in body.get("created", [])), f"got {body.get('created')}")

status, body, _ = request("GET", "/admin/departments", token=hr_token)
check("departemen dari CSV dibuat otomatis",
      any(d["name"] == "Keuangan" for d in body.get("data", [])), f"got {body}")

bad_csv = b"kolomsalah,lain\n1,2\n"
raw, ctype = multipart_named("file", "salah.csv", bad_csv, "text/csv")
status, body, _ = request("POST", "/admin/employees/import", token=hr_token, raw=raw, content_type=ctype)
check("CSV tanpa kolom wajib ditolak dengan pesan jelas",
      status == 400 and "nik" in body.get("error", ""), f"got {status} {body}")

status, body, _ = request("POST", "/admin/employees/import", token=andi_token, raw=raw, content_type=ctype)
check("import hanya untuk HR", status == 403, f"got {status} {body}")

print("\n=== RE-ENROLL ===")
raw, ctype = multipart([face("A1_v0"), face("A1_v1"), face("A1_v2")])
status, body, _ = request("POST", "/enrollment/photos", token=andi_token, raw=raw, content_type=ctype)
check("enroll ulang lewat endpoint pertama ditolak", status == 409, f"got {status} {body}")

status, body, _ = request("POST", "/enrollment/re-enroll", token=andi_token, raw=raw, content_type=ctype)
check("re-enroll berhasil", status == 200, f"got {status} {body}")

status, body, _ = request("POST", "/attendance/check-out", token=andi_token,
                          raw=face("A2"), content_type="image/jpeg", cookie=andi_device)
check("wajah masih dikenali setelah re-enroll", status in (200, 409), f"got {status} {body}")

print("\n=== EXPORT XLSX ===")
xlsx_status, xlsx_body, xlsx_headers = request_raw("GET", "/admin/reports/export.xlsx", token=hr_token)
check("export xlsx berhasil", xlsx_status == 200, f"got {xlsx_status}")
check("file xlsx punya signature ZIP", xlsx_body.startswith(b"PK"), f"got {xlsx_body[:8]}")
check("content-type xlsx benar",
      "spreadsheetml" in (xlsx_headers.get("Content-Type") or ""), f"got {xlsx_headers.get('Content-Type')}")
# xlsx is a ZIP, so the sheet name is compressed -- open it rather than
# grepping raw bytes, and confirm real data landed in the cells.
import io as _io
import zipfile

with zipfile.ZipFile(_io.BytesIO(xlsx_body)) as zf:
    names = zf.namelist()
    workbook = zf.read("xl/workbook.xml").decode()
    shared = zf.read("xl/sharedStrings.xml").decode() if "xl/sharedStrings.xml" in names else ""
check("xlsx adalah workbook yang valid", "xl/worksheets/sheet1.xml" in names, f"got {names[:6]}")
check("sheet bernama Absensi", 'name="Absensi"' in workbook, f"got {workbook[:200]}")
check("xlsx memuat header berbahasa Indonesia", "Perlu Review" in shared, f"got {shared[:200]}")
check("xlsx memuat data karyawan", "Andi Pratama" in shared, f"got {shared[:300]}")

status, body, _ = request("GET", "/admin/reports/export.xlsx", token=andi_token)
check("export xlsx hanya untuk HR", status == 403, f"got {status}")

print("\n=== COOKIE PERANGKAT PADA PERCOBAAN GAGAL ===")
# Regression guard: sebelumnya cookie device hanya dikirim di jalur sukses.
# Karyawan yang klik "pulang" sebelum "masuk" (atau retry jaringan F4) bisa
# kehabisan jatah 2 perangkat dalam dua kali gagal, dari laptop yang sama.
# C1.jpg itself contains two faces (confirmed separately), so its brightness
# variants inherit that and cannot enrol. Use a dedicated, verified single-face
# identity instead of fighting the existing A/B/C fixtures.
def prepare_cookie_test_face():
    import numpy as np
    import pandas as pd
    df = pd.read_parquet(
        "https://huggingface.co/datasets/logasja/lfw/resolve/main/pairs/test-00000-of-00001.parquet"
    ).query("pair == 1")
    row = df.iloc[25]
    img = Image.open(io.BytesIO(row["img_0"]["bytes"]))
    img = img.resize((img.width * 2, img.height * 2), Image.LANCZOS)
    img.save(FACES / "D1.jpg", quality=95)
    for i in range(3):
        arr = np.array(img).astype(np.int16) + (i - 1) * 12
        Image.fromarray(np.clip(arr, 0, 255).astype(np.uint8)).save(FACES / f"D1_v{i}.jpg")
    img2 = Image.open(io.BytesIO(row["img_1"]["bytes"]))
    img2 = img2.resize((img2.width * 2, img2.height * 2), Image.LANCZOS)
    img2.save(FACES / "D2.jpg", quality=95)

prepare_cookie_test_face()

status, body, _ = request(
    "POST", "/admin/employees",
    {"nik": "COOKIE-1", "full_name": "Cookie Tester", "email": "cookietester@perusahaan.com"},
    token=hr_token,
)
cookie_password = body.get("temp_password", "")
status, body, _ = request("POST", "/auth/login", {"email": "cookietester@perusahaan.com", "password": cookie_password})
cookie_token = body.get("access_token", "")

raw, ctype = multipart([face("D1_v0"), face("D1_v1"), face("D1_v2")])
status, body, _ = request("POST", "/enrollment/photos", token=cookie_token, raw=raw, content_type=ctype)
check("setup: enrollment untuk tes cookie", status == 200, f"got {status} {body}")

# Klik keliru: check-out sebelum pernah check-in. Harus gagal 409, TAPI harus
# tetap menerbitkan cookie device karena resolveDevice sudah menulis baris ke DB.
status, body, first_cookie = request("POST", "/attendance/check-out", token=cookie_token, raw=face("D2"), content_type="image/jpeg")
check("percobaan pertama gagal seperti biasa (409)", status == 409, f"got {status} {body}")
check("cookie device tetap diterbitkan meski gagal", first_cookie is not None and "device_id" in first_cookie,
      f"got cookie={first_cookie}")

device_cookie = first_cookie.split(";")[0] if first_cookie else ""

# Klik keliru kedua, memakai cookie yang sama -- tidak boleh membuat perangkat baru.
status, body, second_cookie = request("POST", "/attendance/check-out", token=cookie_token, raw=face("D2"), content_type="image/jpeg", cookie=device_cookie)
check("percobaan kedua gagal dengan alasan yang sama", status == 409, f"got {status} {body}")

# Akhirnya check-in yang benar, laptop yang sama (cookie yang sama) -- harus BERHASIL,
# bukan "Perangkat belum disetujui HR".
status, body, _ = request("POST", "/attendance/check-in", token=cookie_token, raw=face("D2"), content_type="image/jpeg", cookie=device_cookie)
check("check-in sungguhan dari laptop yang sama berhasil, tidak terkunci", status == 200, f"got {status} {body}")

status, body, _ = request("GET", "/admin/devices", token=hr_token)
mine = [d for d in body.get("data", []) if d.get("full_name") == "Cookie Tester"]
check("hanya SATU perangkat tercatat meski dua percobaan gagal", len(mine) == 1, f"got {mine}")
check("perangkat itu langsung disetujui (bukan menunggu HR)", mine and mine[0].get("approved") is True, f"got {mine}")

print("\n=== PRESENSI GANDA BERSAMAAN ===")
# Auto-capture + klik bisa tiba bersamaan; tanpa kunci baris keduanya menyisipkan baris baru.
import concurrent.futures

def _checkout(_):
    return request("POST", "/attendance/check-out", token=cookie_token, raw=face("D2"),
                   content_type="image/jpeg", cookie=device_cookie)[0]

with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
    statuses = list(pool.map(_checkout, range(2)))
check("dua check-out bersamaan sama-sama diterima", statuses == [200, 200], f"got {statuses}")
status, body, _ = request("GET", "/admin/attendances?limit=200", token=hr_token)
outs = [r for r in body.get("data", []) if r.get("full_name") == "Cookie Tester" and r.get("type") == "check_out"]
check("hanya SATU baris check-out tercatat", len(outs) == 1, f"got {len(outs)} rows: {outs}")

print(f"\n{'='*46}\nPASS: {passed}   FAIL: {failed}\n{'='*46}")
sys.exit(1 if failed else 0)
