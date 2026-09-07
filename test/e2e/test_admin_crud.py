"""Verifies the HR-side CRUD added on top of the base build: employee
edit/deactivate, work-location edit/delete (incl. the in-use conflict), and
attendance edit with schedule-derived status re-computation.
"""

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


print("\n=== SETUP ===")
status, body = req("POST", "/auth/login", {"email": "hr@perusahaan.com", "password": "rahasia-sekali-123"})
hr_token = body.get("access_token", "")
check("login HR", status == 200, f"got {status} {body}")

status, dept, _ = (*req("POST", "/admin/departments", {"name": "Produksi"}, token=hr_token), None)
check("departemen dibuat", status == 201, f"got {status} {dept}")

status, loc_a = req("POST", "/admin/locations", {"name": "Kantor Cabang A"}, token=hr_token)
check("lokasi A dibuat", status == 201, f"got {status} {loc_a}")
status, loc_b = req("POST", "/admin/locations", {"name": "Kantor Cabang B"}, token=hr_token)
check("lokasi B dibuat", status == 201, f"got {status} {loc_b}")

status, body = req(
    "POST", "/admin/employees",
    {"nik": "CRUD-1", "full_name": "Fajar Nugraha", "email": "fajar@perusahaan.com",
     "department_id": dept.get("id"), "location_id": loc_a.get("id")},
    token=hr_token,
)
check("karyawan dibuat dengan lokasi", status == 201, f"got {status} {body}")
emp_id = body.get("employee", {}).get("id", "")
check("employee.location_id ikut dikembalikan saat create", body.get("employee", {}).get("location_id") == loc_a.get("id"),
      f"got {body.get('employee')}")

print("\n=== EDIT KARYAWAN ===")
status, body = req(
    "PUT", f"/admin/employees/{emp_id}",
    {"full_name": "Fajar Nugraha Wijaya", "email": "fajar.w@perusahaan.com",
     "department_id": dept.get("id"), "location_id": loc_b.get("id")},
    token=hr_token,
)
check("HR bisa mengubah karyawan", status == 200, f"got {status} {body}")

status, body = req("GET", "/admin/employees", token=hr_token)
updated = next((e for e in body.get("data", []) if e["id"] == emp_id), None)
check("nama tersimpan setelah edit", updated and updated["full_name"] == "Fajar Nugraha Wijaya", f"got {updated}")
check("email tersimpan setelah edit", updated and updated["email"] == "fajar.w@perusahaan.com", f"got {updated}")
check("lokasi berpindah ke B setelah edit", updated and updated.get("location_id") == loc_b.get("id"), f"got {updated}")

# Deliberately OMITS department_id/location_id -- this is the exact shape that
# exposed the COALESCE bug (a partial payload used to silently wipe both
# fields to NULL). Re-check afterward that they survived untouched.
status, body = req("PUT", f"/admin/employees/{emp_id}",
                    {"full_name": "Fajar Nugraha Wijaya", "email": "fajar.w@perusahaan.com"}, token=hr_token)
check("edit dengan email milik diri sendiri tidak dianggap konflik", status == 200, f"got {status} {body}")

status, body = req("GET", "/admin/employees", token=hr_token)
still_there = next((e for e in body.get("data", []) if e["id"] == emp_id), None)
check("department_id TIDAK terhapus walau tidak dikirim di payload",
      still_there and still_there.get("department_id") == dept.get("id"), f"got {still_there}")
check("location_id TIDAK terhapus walau tidak dikirim di payload",
      still_there and still_there.get("location_id") == loc_b.get("id"), f"got {still_there}")

status, body = req("PUT", "/admin/employees/00000000-0000-0000-0000-000000000000",
                    {"full_name": "Tidak Ada", "email": "tidakada@perusahaan.com"}, token=hr_token)
check("edit karyawan tak ada -> 404", status == 404, f"got {status} {body}")

andi_status, andi_body = req(
    "POST", "/admin/employees",
    {"nik": "CRUD-2", "full_name": "Andi Kedua", "email": "andikedua@perusahaan.com"}, token=hr_token)
andi_id = andi_body.get("employee", {}).get("id", "")
status, body = req("PUT", f"/admin/employees/{andi_id}",
                    {"full_name": "Andi Kedua", "email": "fajar.w@perusahaan.com"}, token=hr_token)
check("edit ke email milik orang lain -> 409", status == 409, f"got {status} {body}")

print("\n=== NONAKTIFKAN KARYAWAN (SOFT DELETE) ===")
status, body = req("DELETE", f"/admin/employees/{andi_id}", token=hr_token)
check("HR bisa menonaktifkan karyawan", status == 200 and body.get("status") == "inactive", f"got {status} {body}")

status, body = req("GET", "/admin/employees", token=hr_token)
andi_after = next((e for e in body.get("data", []) if e["id"] == andi_id), None)
check("status jadi inactive, baris TETAP ada (bukan hard delete)",
      andi_after is not None and andi_after["status"] == "inactive", f"got {andi_after}")

status, body = req("DELETE", "/admin/employees/00000000-0000-0000-0000-000000000000", token=hr_token)
check("nonaktifkan karyawan tak ada -> 404", status == 404, f"got {status} {body}")

print("\n=== EDIT & HAPUS LOKASI ===")
status, body = req("PUT", f"/admin/locations/{loc_b.get('id')}", {"name": "Kantor Cabang B (Renov)"}, token=hr_token)
check("HR bisa mengubah nama lokasi", status == 200, f"got {status} {body}")

status, body = req("GET", "/admin/locations", token=hr_token)
loc_b_after = next((l for l in body.get("data", []) if l["id"] == loc_b.get("id")), None)
check("nama lokasi tersimpan", loc_b_after and loc_b_after["name"] == "Kantor Cabang B (Renov)", f"got {loc_b_after}")

status, body = req("DELETE", f"/admin/locations/{loc_b.get('id')}", token=hr_token)
check("hapus lokasi yang masih dipakai karyawan aktif -> 409, bukan crash",
      status == 409 and "dipakai" in body.get("error", ""), f"got {status} {body}")

status, unused_loc = req("POST", "/admin/locations", {"name": "Gudang (tidak dipakai)"}, token=hr_token)
status, body = req("DELETE", f"/admin/locations/{unused_loc.get('id')}", token=hr_token)
check("hapus lokasi yang TIDAK dipakai -> berhasil", status == 200, f"got {status} {body}")

status, body = req("PUT", "/admin/locations/00000000-0000-0000-0000-000000000000", {"name": "X"}, token=hr_token)
check("edit lokasi tak ada -> 404", status == 404, f"got {status} {body}")

print("\n=== EDIT TANGGAL ABSENSI (status dihitung ulang) ===")
# No face pipeline available in this test run, so seed a real attendance row
# through the already-covered correction-approval path instead (it writes a
# genuine attendances row with source=manual_correction, status hardcoded
# on_time regardless of the time chosen -- see correction/service.go). That
# gives Update() something real to re-derive status FROM, which is the actual
# thing under test here.
import datetime

status, seed_body = req("POST", "/admin/employees",
                         {"nik": "CRUD-3", "full_name": "Edit Waktu Tester", "email": "edittest@perusahaan.com"},
                         token=hr_token)
seed_password = seed_body.get("temp_password", "")
check("setup: karyawan untuk uji edit waktu dibuat", status == 201 and seed_password, f"got {status} {seed_body}")

status, login_body = req("POST", "/auth/login", {"email": "edittest@perusahaan.com", "password": seed_password})
seed_token = login_body.get("access_token", "")

early_morning = datetime.datetime.now().astimezone().replace(hour=7, minute=30, second=0, microsecond=0).isoformat()
status, corr = req("POST", "/corrections",
                    {"requested_type": "check_in", "requested_time": early_morning,
                     "reason": "lupa absen pagi ini, baru ingat sekarang"},
                    token=seed_token)
check("setup: pengajuan koreksi dibuat", status == 201, f"got {status} {corr}")

status, body = req("PATCH", f"/admin/corrections/{corr.get('id')}", {"decision": "approve"}, token=hr_token)
check("setup: koreksi disetujui HR (menghasilkan baris absensi nyata)", status == 200, f"got {status} {body}")

status, body = req("GET", "/admin/attendances?status=on_time", token=hr_token)
seeded = next((r for r in body.get("data", []) if r["full_name"] == "Edit Waktu Tester"), None)
check("baris absensi hasil koreksi ditemukan, status awal on_time", seeded is not None, f"got {body}")

if seeded:
    att_id = seeded["id"]
    late_time = datetime.datetime.now().astimezone().replace(hour=11, minute=0, second=0, microsecond=0).isoformat()
    status, body = req("PUT", f"/admin/attendances/{att_id}", {"occurred_at": late_time}, token=hr_token)
    check("HR bisa mengubah waktu absensi", status == 200, f"got {status} {body}")
    check("status dihitung ulang dari jadwal (check-in jam 11 -> late, bukan tetap on_time)",
          body.get("status") == "late", f"got {body}")
    check("occurred_at tersimpan sesuai yang diedit",
          body.get("occurred_at", "").startswith(late_time[:16]), f"got {body.get('occurred_at')}")
    check("source tetap manual_correction setelah diedit", body.get("source") == "manual_correction", f"got {body}")

    status, body = req("GET", "/admin/attendances?status=late", token=hr_token)
    check("perubahan status terlihat lewat filter status=late",
          any(r["id"] == att_id for r in body.get("data", [])), f"got {body}")
else:
    check("(dilewati) HR bisa mengubah waktu absensi", False, "tidak ada baris untuk diuji -- setup gagal di atas")

status, body = req("PUT", "/admin/attendances/00000000-0000-0000-0000-000000000000",
                    {"occurred_at": "2026-01-01T08:00:00+07:00"}, token=hr_token)
check("edit absensi tak ada -> 404", status == 404, f"got {status} {body}")

print(f"\n{'='*46}\nPASS: {passed}   FAIL: {failed}\n{'='*46}")
sys.exit(1 if failed else 0)
