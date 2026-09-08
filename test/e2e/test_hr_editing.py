"""Covers the editing powers HR asked for: full employee editing (including
NIK, quota, status and a password reset), editing a leave request in place,
editing departments and locations, and every account editing its own profile.
"""

import datetime
import json
import sys
import urllib.error
import urllib.request

API = "http://127.0.0.1:8080/api/v1"
passed = failed = 0


def check(name, ok, detail=""):
    global passed, failed
    if ok:
        passed += 1
        print(f"  PASS  {name}")
    else:
        failed += 1
        print(f"  FAIL  {name}  {detail}")


def req(method, path, body=None, token=None):
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(API + path, data=data, method=method)
    if data is not None:
        r.add_header("Content-Type", "application/json")
    if token:
        r.add_header("Authorization", f"Bearer {token}")
    try:
        with urllib.request.urlopen(r) as resp:
            payload = resp.read().decode()
            return resp.status, (json.loads(payload) if payload else {})
    except urllib.error.HTTPError as e:
        payload = e.read().decode()
        try:
            return e.code, json.loads(payload)
        except json.JSONDecodeError:
            return e.code, {"raw": payload}


def next_weekday(offset=7):
    d = datetime.date.today() + datetime.timedelta(days=offset)
    while d.isoweekday() > 5:
        d += datetime.timedelta(days=1)
    return d


print("\n=== SETUP ===")
status, body = req("POST", "/auth/login", {"email": "hr@perusahaan.com", "password": "rahasia-sekali-123"})
hr_token = body.get("access_token", "")
check("login HR", status == 200, f"got {status} {body}")

status, dept = req("POST", "/admin/departments", {"name": "Divisi Uji"}, token=hr_token)
check("departemen dibuat", status == 201, f"got {status} {dept}")
status, dept2 = req("POST", "/admin/departments", {"name": "Divisi Kedua"}, token=hr_token)
status, loc = req("POST", "/admin/locations", {"name": "Lokasi Uji"}, token=hr_token)
check("lokasi dibuat", status == 201, f"got {status} {loc}")
status, sched = req("GET", "/admin/schedules", token=hr_token)
schedule_id = (sched.get("data") or [{}])[0].get("id")

status, body = req("POST", "/admin/employees",
                    {"nik": "EDIT-1", "full_name": "Editable Satu", "email": "edit1@perusahaan.com",
                     "department_id": dept.get("id"), "location_id": loc.get("id")}, token=hr_token)
emp_id = body.get("employee", {}).get("id", "")
emp_password = body.get("temp_password", "")
check("karyawan dibuat", status == 201, f"got {status} {body}")

status, body = req("POST", "/admin/employees",
                    {"nik": "EDIT-2", "full_name": "Editable Dua", "email": "edit2@perusahaan.com"}, token=hr_token)
emp2_id = body.get("employee", {}).get("id", "")

print("\n=== HR EDIT SELURUH DATA KARYAWAN ===")
status, body = req("PUT", f"/admin/employees/{emp_id}", {
    "nik": "EDIT-1-BARU",
    "full_name": "Editable Satu Diubah",
    "email": "edit1baru@perusahaan.com",
    "department_id": dept2.get("id"),
    "schedule_id": schedule_id,
    "manager_id": emp2_id,
    "annual_leave_quota": 20,
    "status": "active",
}, token=hr_token)
check("HR bisa mengubah seluruh field karyawan", status == 200, f"got {status} {body}")

status, body = req("GET", "/admin/employees", token=hr_token)
row = next((e for e in body.get("data", []) if e["id"] == emp_id), None)
check("NIK tersimpan", row and row["nik"] == "EDIT-1-BARU", f"got {row}")
check("nama tersimpan", row and row["full_name"] == "Editable Satu Diubah", f"got {row}")
check("email tersimpan", row and row["email"] == "edit1baru@perusahaan.com", f"got {row}")
check("departemen berpindah", row and row.get("department_id") == dept2.get("id"), f"got {row}")
check("kuota cuti tersimpan", row and row.get("annual_leave_quota") == 20, f"got {row}")
check("status tersimpan", row and row["status"] == "active", f"got {row}")
check("atasan tersimpan", row and row.get("manager_id") == emp2_id, f"got {row}")
check("lokasi TIDAK hilang walau tidak dikirim", row and row.get("location_id") == loc.get("id"), f"got {row}")

# "" means clear, which is different from omitting the field entirely.
status, body = req("PUT", f"/admin/employees/{emp_id}", {
    "full_name": "Editable Satu Diubah", "email": "edit1baru@perusahaan.com", "manager_id": "",
}, token=hr_token)
status, body = req("GET", "/admin/employees", token=hr_token)
row = next((e for e in body.get("data", []) if e["id"] == emp_id), None)
check("kirim manager_id kosong -> atasan dihapus", row and not row.get("manager_id"), f"got {row}")
check("departemen tetap ada setelah hapus atasan", row and row.get("department_id") == dept2.get("id"), f"got {row}")

status, body = req("PUT", f"/admin/employees/{emp_id}", {
    "full_name": "X", "email": "edit1baru@perusahaan.com", "nik": "EDIT-2"}, token=hr_token)
check("NIK bentrok -> 409", status == 409, f"got {status} {body}")

status, body = req("PUT", f"/admin/employees/{emp_id}", {
    "full_name": "X", "email": "edit1baru@perusahaan.com", "manager_id": emp_id}, token=hr_token)
check("atasan = diri sendiri -> 400", status == 400, f"got {status} {body}")

status, body = req("PUT", f"/admin/employees/{emp_id}", {
    "full_name": "X", "email": "edit1baru@perusahaan.com", "status": "ngawur"}, token=hr_token)
check("status tidak valid -> 400", status == 400, f"got {status} {body}")

print("\n=== HR RESET PASSWORD KARYAWAN ===")
status, body = req("PUT", f"/admin/employees/{emp_id}", {
    "full_name": "Editable Satu Diubah", "email": "edit1baru@perusahaan.com", "password": "pendek"},
    token=hr_token)
check("password terlalu pendek -> 400", status == 400, f"got {status} {body}")

status, body = req("PUT", f"/admin/employees/{emp_id}", {
    "full_name": "Editable Satu Diubah", "email": "edit1baru@perusahaan.com",
    "password": "PasswordBaru123"}, token=hr_token)
check("HR bisa reset password karyawan", status == 200, f"got {status} {body}")

status, body = req("POST", "/auth/login", {"email": "edit1baru@perusahaan.com", "password": "PasswordBaru123"})
check("karyawan bisa login dengan password baru", status == 200, f"got {status} {body}")
emp_token = body.get("access_token", "")

status, body = req("POST", "/auth/login", {"email": "edit1baru@perusahaan.com", "password": emp_password})
check("password lama sudah tidak berlaku", status == 401, f"got {status} {body}")

print("\n=== PROFIL SENDIRI ===")
status, body = req("GET", "/me/profile", token=hr_token)
check("HR bisa membaca profilnya sendiri", status == 200, f"got {status} {body}")
check("profil HR menandai bukan karyawan", body.get("is_employee") is False, f"got {body}")
check("role ikut dikembalikan", body.get("role") in ("hr", "superadmin"), f"got {body}")

status, body = req("PATCH", "/me/profile", {"full_name": "Ratna Admin"}, token=hr_token)
check("HR bisa mengubah namanya sendiri", status == 200, f"got {status} {body}")
status, body = req("GET", "/me/profile", token=hr_token)
check("nama HR tersimpan walau tanpa data karyawan", body.get("full_name") == "Ratna Admin", f"got {body}")

status, body = req("GET", "/me/profile", token=emp_token)
check("karyawan membaca profilnya sendiri", status == 200 and body.get("is_employee") is True, f"got {body}")

status, body = req("PATCH", "/me/profile", {"full_name": "Nama Pilihan Sendiri"}, token=emp_token)
check("karyawan bisa mengubah namanya sendiri", status == 200, f"got {status} {body}")
status, body = req("GET", "/admin/employees", token=hr_token)
row = next((e for e in body.get("data", []) if e["id"] == emp_id), None)
check("perubahan nama karyawan terlihat oleh HR", row and row["full_name"] == "Nama Pilihan Sendiri", f"got {row}")

status, body = req("PATCH", "/me/profile",
                    {"current_password": "salah-sekali", "new_password": "PasswordLagi123"}, token=emp_token)
check("ganti password tanpa password lama yang benar -> 400", status == 400, f"got {status} {body}")

status, body = req("PATCH", "/me/profile",
                    {"current_password": "PasswordBaru123", "new_password": "pendek"}, token=emp_token)
check("password baru terlalu pendek -> 400", status == 400, f"got {status} {body}")

status, body = req("PATCH", "/me/profile",
                    {"current_password": "PasswordBaru123", "new_password": "PasswordLagi123"}, token=emp_token)
check("karyawan bisa ganti password sendiri", status == 200, f"got {status} {body}")
status, body = req("POST", "/auth/login", {"email": "edit1baru@perusahaan.com", "password": "PasswordLagi123"})
check("login dengan password hasil ganti sendiri", status == 200, f"got {status} {body}")
emp_token = body.get("access_token", "")

status, body = req("PATCH", "/me/profile", {"email": "hr@perusahaan.com"}, token=emp_token)
check("ganti email ke milik orang lain -> 409", status == 409, f"got {status} {body}")

print("\n=== EDIT DEPARTEMEN ===")
status, body = req("PUT", f"/admin/departments/{dept.get('id')}", {"name": "Divisi Uji (Ubah)"}, token=hr_token)
check("HR bisa mengubah nama departemen", status == 200, f"got {status} {body}")
status, body = req("GET", "/admin/departments", token=hr_token)
d = next((x for x in body.get("data", []) if x["id"] == dept.get("id")), None)
check("nama departemen tersimpan", d and d["name"] == "Divisi Uji (Ubah)", f"got {d}")

status, body = req("DELETE", f"/admin/departments/{dept2.get('id')}", token=hr_token)
check("hapus departemen yang masih dipakai -> 409", status == 409, f"got {status} {body}")

status, body = req("DELETE", f"/admin/departments/{dept.get('id')}", token=hr_token)
check("hapus departemen yang tidak dipakai -> 200", status == 200, f"got {status} {body}")

status, body = req("PUT", "/admin/departments/00000000-0000-0000-0000-000000000000", {"name": "X"}, token=hr_token)
check("edit departemen tak ada -> 404", status == 404, f"got {status} {body}")

print("\n=== HR EDIT PENGAJUAN CUTI ===")
start = next_weekday(7)
status, lv = req("POST", "/leave-requests",
                  {"type": "annual", "start_date": str(start), "end_date": str(start),
                   "reason": "acara keluarga"}, token=emp_token)
check("karyawan mengajukan cuti", status == 201, f"got {status} {lv}")
leave_id = lv.get("id", "")

new_end = start + datetime.timedelta(days=1)
while new_end.isoweekday() > 5:
    new_end += datetime.timedelta(days=1)
status, body = req("PUT", f"/admin/leave-requests/{leave_id}",
                    {"type": "sick", "start_date": str(start), "end_date": str(new_end),
                     "reason": "diralat HR: sakit, ada surat dokter"}, token=hr_token)
check("HR bisa mengubah pengajuan cuti", status == 200, f"got {status} {body}")
check("jenis cuti berubah", body.get("type") == "sick", f"got {body}")
check("tanggal selesai berubah", body.get("end_date") == str(new_end), f"got {body}")
check("hari kerja dihitung ulang", body.get("days_count") == 2, f"got {body}")
check("alasan berubah", "diralat HR" in body.get("reason", ""), f"got {body}")

status, body = req("PUT", f"/admin/leave-requests/{leave_id}",
                    {"type": "sick", "start_date": str(new_end), "end_date": str(start),
                     "reason": "tanggal terbalik"}, token=hr_token)
check("tanggal terbalik -> 400", status == 400, f"got {status} {body}")

# Editing a request without changing its dates must not trip the overlap check
# against itself.
status, body = req("PUT", f"/admin/leave-requests/{leave_id}",
                    {"type": "sick", "start_date": str(start), "end_date": str(new_end),
                     "reason": "disimpan ulang tanpa ubah tanggal"}, token=hr_token)
check("simpan ulang tanpa ubah tanggal tidak dianggap bentrok", status == 200, f"got {status} {body}")

status, body = req("PUT", f"/admin/leave-requests/{leave_id}",
                    {"type": "annual", "start_date": str(start), "end_date": str(start),
                     "reason": "melebihi kuota"}, token=hr_token)
check("edit ke cuti tahunan masih dalam kuota -> 200", status == 200, f"got {status} {body}")

status, body = req("PUT", "/admin/leave-requests/00000000-0000-0000-0000-000000000000",
                    {"type": "sick", "start_date": str(start), "end_date": str(start), "reason": "x"},
                    token=hr_token)
check("edit pengajuan tak ada -> 404", status == 404, f"got {status} {body}")

status, body = req("PUT", f"/admin/leave-requests/{leave_id}",
                    {"type": "sick", "start_date": str(start), "end_date": str(start), "reason": "x"},
                    token=emp_token)
check("karyawan biasa tidak boleh mengedit pengajuan -> 403", status == 403, f"got {status} {body}")

print(f"\n{'='*46}\nPASS: {passed}   FAIL: {failed}\n{'='*46}")
sys.exit(1 if failed else 0)
