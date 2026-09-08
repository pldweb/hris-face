"""Fills a freshly-seeded database with realistic demo data so every screen has
something to show: departments, work locations, employees, attendance rows,
a pending and an approved leave request, and a pending correction.

Assumes the API is running and the superadmin from `api/bin/seed` exists.

    python3 test/e2e/seed_demo.py

Attendance rows are created through the correction-approval path rather than
the face pipeline, because the face service is not needed (or wanted) just to
get demo data on screen.
"""

import datetime
import json
import sys
import urllib.error
import urllib.request

API = "http://127.0.0.1:8080/api/v1"
HR = {"email": "hr@perusahaan.com", "password": "rahasia-sekali-123"}


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


def workday(offset_days):
    """A date `offset_days` back from today, skipped back off weekends."""
    d = datetime.date.today() - datetime.timedelta(days=offset_days)
    while d.isoweekday() > 5:
        d -= datetime.timedelta(days=1)
    return d


status, body = req("POST", "/auth/login", HR)
if status != 200:
    print(f"gagal login sebagai HR ({status}): {body}")
    print("pastikan API jalan dan `api/bin/seed` sudah dijalankan.")
    sys.exit(1)
hr_token = body["access_token"]
print("login HR OK")

departments = {}
for name in ["Produksi", "Keuangan", "IT"]:
    status, d = req("POST", "/admin/departments", {"name": name}, token=hr_token)
    if status == 201:
        departments[name] = d["id"]
        print(f"  departemen: {name}")

locations = {}
for name in ["Kantor Pusat", "Cabang Bandung"]:
    status, l = req("POST", "/admin/locations", {"name": name}, token=hr_token)
    if status == 201:
        locations[name] = l["id"]
        print(f"  lokasi: {name}")

PEOPLE = [
    ("2024001", "Budi Santoso", "budi@perusahaan.com", "Produksi", "Kantor Pusat"),
    ("2024002", "Siti Rahayu", "siti@perusahaan.com", "Keuangan", "Kantor Pusat"),
    ("2024003", "Agus Wijaya", "agus@perusahaan.com", "IT", "Cabang Bandung"),
    ("2024004", "Dewi Lestari", "dewi@perusahaan.com", "Produksi", "Cabang Bandung"),
]

employees = []
for nik, name, email, dept, loc in PEOPLE:
    status, body = req(
        "POST", "/admin/employees",
        {"nik": nik, "full_name": name, "email": email,
         "department_id": departments.get(dept), "location_id": locations.get(loc)},
        token=hr_token,
    )
    if status != 201:
        print(f"  lewati {name}: {status} {body}")
        continue
    emp = body["employee"]
    password = body["temp_password"]
    _, login = req("POST", "/auth/login", {"email": email, "password": password})
    employees.append({"id": emp["id"], "name": name, "email": email,
                      "password": password, "token": login.get("access_token", "")})
    print(f"  karyawan: {name} ({email} / {password})")

if not employees:
    print("tidak ada karyawan dibuat -- mungkin data demo sudah ada. selesai.")
    sys.exit(0)

# Attendance, via correction requests HR then approves: gives Monitoring and
# the employee history real rows without needing the face service.
print("membuat data kehadiran...")
for i, emp in enumerate(employees[:3]):
    day = workday(i + 1)
    for kind, hour, minute in (("check_in", 8, 5 + i * 20), ("check_out", 17, 10)):
        at = datetime.datetime.combine(day, datetime.time(hour, minute)).astimezone()
        status, corr = req(
            "POST", "/corrections",
            {"requested_type": kind, "requested_time": at.isoformat(),
             "reason": "absen demo untuk pengisian data awal"},
            token=emp["token"],
        )
        if status == 201:
            req("PATCH", f"/admin/corrections/{corr['id']}", {"decision": "approve"}, token=hr_token)
    print(f"  kehadiran: {emp['name']} pada {day}")

# One correction left pending so the Koreksi Absen screen is not empty.
pending_day = workday(6)
at = datetime.datetime.combine(pending_day, datetime.time(8, 0)).astimezone()
status, _ = req(
    "POST", "/corrections",
    {"requested_type": "check_in", "requested_time": at.isoformat(),
     "reason": "kamera tidak mengenali wajah, sudah coba 3 kali"},
    token=employees[-1]["token"],
)
if status == 201:
    print(f"  koreksi menunggu: {employees[-1]['name']}")

print("membuat pengajuan cuti...")
# Approved annual leave (shows as on_leave in history + reduces balance).
start = workday(3)
status, lv = req(
    "POST", "/leave-requests",
    {"type": "annual", "start_date": str(start), "end_date": str(start),
     "reason": "acara keluarga"},
    token=employees[0]["token"],
)
if status == 201:
    req("PATCH", f"/admin/leave-requests/{lv['id']}", {"decision": "approve"}, token=hr_token)
    print(f"  cuti disetujui: {employees[0]['name']} pada {start}")

# Pending requests so the admin approval queue has something to act on.
future = datetime.date.today() + datetime.timedelta(days=7)
while future.isoweekday() > 5:
    future += datetime.timedelta(days=1)
for emp, kind, reason in (
    (employees[1], "annual", "cuti tahunan, rencana pulang kampung"),
    (employees[2], "sick", "demam, surat dokter menyusul"),
):
    status, _ = req(
        "POST", "/leave-requests",
        {"type": kind, "start_date": str(future), "end_date": str(future),
         "reason": reason},
        token=emp["token"],
    )
    if status == 201:
        print(f"  cuti menunggu: {emp['name']} ({kind}) pada {future}")
    future += datetime.timedelta(days=1)
    while future.isoweekday() > 5:
        future += datetime.timedelta(days=1)

# Promote one person to manager with direct reports, so the Tim screen has a
# user it actually works for. There is no API for this yet (roles and
# manager_id are not editable through the admin UI), so it goes through psql.
import subprocess

manager = employees[1]
reports = [e["id"] for e in employees if e["id"] != manager["id"]]
sql = (
    f"UPDATE users SET role = 'manager' "
    f"WHERE email = '{manager['email']}'; "
    f"UPDATE employees SET manager_id = '{manager['id']}' "
    f"WHERE id IN ({', '.join(chr(39) + r + chr(39) for r in reports)});"
)
try:
    done = subprocess.run(
        ["docker", "exec", "hris-local-db", "psql", "-U", "postgres", "-d", "hris", "-c", sql],
        capture_output=True, text=True, timeout=15,
    )
    if done.returncode == 0:
        print(f"  manager: {manager['name']} membawahi {len(reports)} karyawan")
    else:
        print(f"  (lewati promosi manager: {done.stderr.strip()[:80]})")
except (OSError, subprocess.SubprocessError) as err:
    print(f"  (lewati promosi manager, psql tidak tersedia: {err})")

print("\nselesai. akun untuk login:")
print(f"  HR/admin : {HR['email']} / {HR['password']}")
for emp in employees:
    role = " (manager)" if emp["id"] == manager["id"] else ""
    print(f"  karyawan : {emp['email']} / {emp['password']}{role}")
