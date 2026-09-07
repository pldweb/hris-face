"""Verifies the leave/permission/sick (cuti/izin/sakit) module: create, list,
balance, overlap/quota rejection, HR review, and the report/history integration
that shows approved leave days as 'on_leave' instead of 'absent'.
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


def prev_weekday(d, weekday):
    """Most recent date (<= d) with the given Python weekday (Mon=0..Sun=6)."""
    days_back = (d.weekday() - weekday) % 7
    return d - datetime.timedelta(days=days_back)


print("\n=== SETUP ===")
status, body = req("POST", "/auth/login", {"email": "hr@perusahaan.com", "password": "rahasia-sekali-123"})
hr_token = body.get("access_token", "")
check("login HR", status == 200, f"got {status} {body}")

suffix = str(int(datetime.datetime.now().timestamp()))
status, seed_body = req(
    "POST", "/admin/employees",
    {"nik": f"LEAVE-{suffix}", "full_name": "Leave Tester", "email": f"leavetest{suffix}@perusahaan.com"},
    token=hr_token,
)
check("setup: karyawan untuk uji cuti dibuat", status == 201, f"got {status} {seed_body}")
seed_password = seed_body.get("temp_password", "")

status, login_body = req("POST", "/auth/login", {"email": f"leavetest{suffix}@perusahaan.com", "password": seed_password})
emp_token = login_body.get("access_token", "")
check("login karyawan", status == 200, f"got {status} {login_body}")

# Friday -> Monday: 4 calendar days, but only Fri+Mon are workdays (Sat/Sun excluded) -> days_count=2.
# Picked in the recent past (not future) so /attendance/me?month=... -- which caps its
# range at "tomorrow" -- actually returns these days for the history-integration check below.
today = datetime.date.today()
friday = prev_weekday(today - datetime.timedelta(days=14), 4)  # a Friday ~2+ weeks ago
monday = friday + datetime.timedelta(days=3)

print("\n=== KARYAWAN: BUAT PENGAJUAN CUTI ===")
status, leave1 = req(
    "POST", "/leave-requests",
    {"type": "annual", "start_date": friday.isoformat(), "end_date": monday.isoformat(),
     "reason": "liburan keluarga ke luar kota"},
    token=emp_token,
)
check("cuti tahunan dibuat (201)", status == 201, f"got {status} {leave1}")
check("days_count dihitung hanya hari kerja (Jumat+Senin=2, bukan 4 hari kalender)",
      leave1.get("days_count") == 2, f"got {leave1}")
check("full_name ikut dikembalikan", leave1.get("full_name") == "Leave Tester", f"got {leave1}")
check("status awal pending", leave1.get("status") == "pending", f"got {leave1}")

status, mine = req("GET", "/leave-requests/me", token=emp_token)
check("GET /leave-requests/me menampilkan pengajuan", any(r["id"] == leave1.get("id") for r in mine.get("data", [])), f"got {mine}")

status, bal = req("GET", "/leave-requests/me/balance", token=emp_token)
check("balance: used=0 selama belum disetujui", bal.get("used") == 0, f"got {bal}")
check("balance: remaining=quota selama belum disetujui", bal.get("remaining") == bal.get("quota"), f"got {bal}")
quota = bal.get("quota", 12)

print("\n=== VALIDASI: OVERLAP & KUOTA ===")
status, overlap_body = req(
    "POST", "/leave-requests",
    {"type": "annual", "start_date": friday.isoformat(), "end_date": friday.isoformat(),
     "reason": "tumpang tindih dengan pengajuan sebelumnya"},
    token=emp_token,
)
check("pengajuan bertabrakan tanggal -> 409", status == 409, f"got {status} {overlap_body}")

far_start = monday + datetime.timedelta(days=30)
far_end = far_start + datetime.timedelta(days=quota + 10)
status, quota_body = req(
    "POST", "/leave-requests",
    {"type": "annual", "start_date": far_start.isoformat(), "end_date": far_end.isoformat(),
     "reason": "cuti melebihi sisa kuota tahunan"},
    token=emp_token,
)
check("pengajuan melebihi kuota -> 400", status == 400, f"got {status} {quota_body}")
check("pesan error menyebut kuota", "kuota" in quota_body.get("error", ""), f"got {quota_body}")

print("\n=== HR: REVIEW PENGAJUAN ===")
status, admin_list = req("GET", "/admin/leave-requests?status=pending", token=hr_token)
check("HR melihat pengajuan pending", any(r["id"] == leave1.get("id") for r in admin_list.get("data", [])), f"got {admin_list}")

status, approved = req("PATCH", f"/admin/leave-requests/{leave1.get('id')}", {"decision": "approve"}, token=hr_token)
check("HR approve pengajuan -> 200", status == 200, f"got {status} {approved}")
check("status jadi approved", approved.get("status") == "approved", f"got {approved}")

status, bal2 = req("GET", "/leave-requests/me/balance", token=emp_token)
check("balance: used bertambah sesuai days_count setelah disetujui", bal2.get("used") == 2, f"got {bal2}")
check("balance: remaining berkurang sesuai", bal2.get("remaining") == quota - 2, f"got {bal2}")

status, redo = req("PATCH", f"/admin/leave-requests/{leave1.get('id')}", {"decision": "approve"}, token=hr_token)
check("review ulang pengajuan yang sudah diproses -> 409", status == 409, f"got {status} {redo}")

status, notfound = req("PATCH", "/admin/leave-requests/00000000-0000-0000-0000-000000000000",
                        {"decision": "approve"}, token=hr_token)
check("review pengajuan tak ada -> 404", status == 404, f"got {status} {notfound}")

print("\n=== SICK/PERMIT TIDAK MEMPENGARUHI SALDO TAHUNAN ===")
sick_start = monday + datetime.timedelta(days=60)
sick_end = sick_start + datetime.timedelta(days=1)
status, sick = req(
    "POST", "/leave-requests",
    {"type": "sick", "start_date": sick_start.isoformat(), "end_date": sick_end.isoformat(),
     "reason": "demam tinggi, surat dokter menyusul"},
    token=emp_token,
)
check("pengajuan sakit dibuat", status == 201, f"got {status} {sick}")
status, sick_approved = req("PATCH", f"/admin/leave-requests/{sick.get('id')}", {"decision": "approve"}, token=hr_token)
check("pengajuan sakit disetujui", status == 200 and sick_approved.get("status") == "approved", f"got {status} {sick_approved}")

status, bal3 = req("GET", "/leave-requests/me/balance", token=emp_token)
check("saldo tahunan tidak berubah oleh cuti sakit", bal3.get("used") == 2 and bal3.get("remaining") == quota - 2, f"got {bal3}")

print("\n=== INTEGRASI RIWAYAT ABSENSI (on_leave) ===")
month = friday.strftime("%Y-%m")
status, history = req("GET", f"/attendance/me?month={month}", token=emp_token)
by_date = {d["date"]: d for d in history.get("data", [])}
fri_rec = by_date.get(friday.isoformat())
mon_rec = by_date.get(monday.isoformat()) if monday.strftime("%Y-%m") == month else None

check("hari Jumat cuti tampil status on_leave", fri_rec is not None and fri_rec.get("status") == "on_leave", f"got {fri_rec}")
check("leave_type tampil 'annual' untuk hari itu", fri_rec is not None and fri_rec.get("leave_type") == "annual", f"got {fri_rec}")

if monday.strftime("%Y-%m") != month:
    status, history2 = req("GET", f"/attendance/me?month={monday.strftime('%Y-%m')}", token=emp_token)
    mon_rec = next((d for d in history2.get("data", []) if d["date"] == monday.isoformat()), None)
check("hari Senin cuti tampil status on_leave", mon_rec is not None and mon_rec.get("status") == "on_leave", f"got {mon_rec}")

# Weekend day inside the range was never a workday and has no leave row covering
# it as a *counted* day, but the leave_requests row spans start..end inclusive,
# so the approved range still covers Sat/Sun -- confirm that's on_leave too,
# consistent with the LEFT JOIN LATERAL matching any date in [start,end].
saturday = friday + datetime.timedelta(days=1)
sat_rec = by_date.get(saturday.isoformat())
check("hari Sabtu (dalam rentang, bukan hari kerja) tetap tampil on_leave",
      sat_rec is not None and sat_rec.get("status") == "on_leave", f"got {sat_rec}")

print(f"\n{'='*46}\nPASS: {passed}   FAIL: {failed}\n{'='*46}")
sys.exit(1 if failed else 0)
