"""Verifies the movement challenge (docs/PRD.md F5) against the real stack.

The property under test is the one that can actually be verified: a still image
repeated N times must NOT pass, while genuinely different frames of the same
person must. Blink detection itself is deliberately NOT a gate -- see the note
in api/internal/attendance/challenge.go.
"""

import io
import json
import sys
import urllib.error
import urllib.request
from pathlib import Path

import numpy as np
from PIL import Image

API = "http://127.0.0.1:8080/api/v1"
FACES = Path("/tmp/e2e_faces")

passed = failed = 0


def check(name, ok, detail=""):
    global passed, failed
    if ok:
        passed += 1
        print(f"  PASS  {name}")
    else:
        failed += 1
        print(f"  FAIL  {name}  {detail}")


def post(path, body=None, token=None, raw=None, content_type="application/json"):
    data = raw if raw is not None else (json.dumps(body).encode() if body else None)
    req = urllib.request.Request(API + path, data=data, method="POST")
    if data is not None:
        req.add_header("Content-Type", content_type)
    if token:
        req.add_header("Authorization", f"Bearer {token}")
    try:
        with urllib.request.urlopen(req) as resp:
            return resp.status, json.loads(resp.read().decode() or "{}")
    except urllib.error.HTTPError as e:
        payload = e.read().decode()
        try:
            return e.code, json.loads(payload)
        except json.JSONDecodeError:
            return e.code, {"raw": payload}


def frames_multipart(images):
    boundary = "----hrischallenge"
    parts = []
    for i, img in enumerate(images):
        parts.append(
            (
                f"--{boundary}\r\n"
                f'Content-Disposition: form-data; name="frames"; filename="f{i}.jpg"\r\n'
                f"Content-Type: image/jpeg\r\n\r\n"
            ).encode()
            + img
            + b"\r\n"
        )
    parts.append(f"--{boundary}--\r\n".encode())
    return b"".join(parts), f"multipart/form-data; boundary={boundary}"


def jpeg_bytes(img: Image.Image) -> bytes:
    buf = io.BytesIO()
    img.save(buf, "JPEG", quality=92)
    return buf.getvalue()


def shifted(img: Image.Image, dx: int, dy: int) -> Image.Image:
    """Simulates a person moving slightly between frames."""
    arr = np.array(img)
    return Image.fromarray(np.roll(np.roll(arr, dy, axis=0), dx, axis=1))


print("\n=== SETUP ===")
status, body = post("/auth/login", {"email": "hr@perusahaan.com", "password": "rahasia-sekali-123"})
hr_token = body.get("access_token", "")
check("login HR", status == 200, f"got {status} {body}")

status, body = post(
    "/admin/employees",
    {"nik": "CH-1", "full_name": "Challenge Tester", "email": "challenge@perusahaan.com"},
    token=hr_token,
)
temp_password = body.get("temp_password", "")
check("karyawan uji dibuat", status == 201, f"got {status} {body}")

status, body = post("/auth/login", {"email": "challenge@perusahaan.com", "password": temp_password})
emp_token = body.get("access_token", "")
check("login karyawan uji", status == 200, f"got {status} {body}")

base = Image.open(FACES / "C2.jpg").convert("RGB")
enrol = [jpeg_bytes(base), jpeg_bytes(shifted(base, 2, 1)), jpeg_bytes(shifted(base, -2, -1))]
boundary = "----hrisenrol"
parts = []
for i, p in enumerate(enrol):
    parts.append(
        (
            f"--{boundary}\r\n"
            f'Content-Disposition: form-data; name="photos"; filename="p{i}.jpg"\r\n'
            f"Content-Type: image/jpeg\r\n\r\n"
        ).encode()
        + p
        + b"\r\n"
    )
parts.append(f"--{boundary}--\r\n".encode())
status, body = post(
    "/enrollment/photos",
    token=emp_token,
    raw=b"".join(parts),
    content_type=f"multipart/form-data; boundary={boundary}",
)
check("enrollment karyawan uji", status == 200, f"got {status} {body}")

print("\n=== CHALLENGE: FOTO DIAM HARUS GAGAL ===")
still = jpeg_bytes(base)
raw, ctype = frames_multipart([still, still, still, still])
status, body = post("/attendance/challenge/check-in", token=emp_token, raw=raw, content_type=ctype)
check("frame identik ditolak", status == 422, f"got {status} {body}")
check(
    "pesan menyebut tidak ada gerakan",
    "gerakan" in json.dumps(body).lower(),
    f"got {body}",
)

print("\n=== CHALLENGE: FRAME TERLALU SEDIKIT ===")
raw, ctype = frames_multipart([still, still])
status, body = post("/attendance/challenge/check-in", token=emp_token, raw=raw, content_type=ctype)
check("kurang dari 3 frame ditolak 400", status == 400, f"got {status} {body}")

print("\n=== CHALLENGE: GERAKAN ASLI HARUS LOLOS ===")
moving = [
    jpeg_bytes(base),
    jpeg_bytes(shifted(base, 9, 5)),
    jpeg_bytes(shifted(base, -9, -6)),
    jpeg_bytes(shifted(base, 5, -9)),
]
raw, ctype = frames_multipart(moving)
status, body = post("/attendance/challenge/check-in", token=emp_token, raw=raw, content_type=ctype)
check("frame bergerak diterima", status == 200, f"got {status} {body}")
if status == 200:
    ch = body.get("challenge", {})
    check("challenge melaporkan jumlah frame", ch.get("frames") == 4, f"got {ch}")
    check("challenge melaporkan variasi terukur", ch.get("variation", 0) > 4, f"got {ch}")
    check("challenge melaporkan rata-rata liveness", ch.get("mean_liveness", 0) > 0, f"got {ch}")
    check("absensi tercatat atas nama yang benar",
          body.get("result", {}).get("full_name") == "Challenge Tester", f"got {body.get('result')}")

print("\n=== CHALLENGE: WAJAH ORANG LAIN HARUS DITOLAK ===")
other = Image.open(FACES / "B1.jpg").convert("RGB")
mixed = [
    jpeg_bytes(base),
    jpeg_bytes(shifted(other, 9, 5)),
    jpeg_bytes(shifted(base, -9, -6)),
]
raw, ctype = frames_multipart(mixed)
status, body = post("/attendance/challenge/check-out", token=emp_token, raw=raw, content_type=ctype)
check("urutan dengan wajah orang lain ditolak", status == 422, f"got {status} {body}")

print(f"\n{'='*46}\nPASS: {passed}   FAIL: {failed}\n{'='*46}")
sys.exit(1 if failed else 0)
