"""Drives the MCP server over stdio the way an AI client does: real JSON-RPC on
stdin/stdout, handshake first, then tools/list and tools/call.

    python3 test/e2e/test_mcp.py

Needs the API running and test/e2e/seed_demo.py already applied.
"""

import json
import os
import subprocess
import sys

BINARY = os.path.join(os.path.dirname(__file__), "..", "..", "api", "bin", "mcp")
PROTOCOL_VERSION = "2025-06-18"

passed = failed = 0


def check(name, ok, detail=""):
    global passed, failed
    if ok:
        passed += 1
        print(f"  PASS  {name}")
    else:
        failed += 1
        print(f"  FAIL  {name}  {detail}")


class MCP:
    """Minimal MCP stdio client: newline-delimited JSON-RPC 2.0."""

    def __init__(self, env):
        self.proc = subprocess.Popen(
            [os.path.abspath(BINARY)],
            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
            text=True, bufsize=1, env={**os.environ, **env},
        )
        self._id = 0

    def send(self, method, params=None, notify=False):
        msg = {"jsonrpc": "2.0", "method": method}
        if params is not None:
            msg["params"] = params
        if not notify:
            self._id += 1
            msg["id"] = self._id
        self.proc.stdin.write(json.dumps(msg) + "\n")
        self.proc.stdin.flush()
        if notify:
            return None
        # Skip any server-initiated traffic that is not our reply.
        while True:
            line = self.proc.stdout.readline()
            if not line:
                raise RuntimeError(f"server menutup koneksi. stderr: {self.stderr()}")
            resp = json.loads(line)
            if resp.get("id") == self._id:
                return resp

    def stderr(self):
        try:
            self.proc.stderr.flush()
        except Exception:
            pass
        return "(lihat log stderr server)"

    def close(self):
        try:
            self.proc.stdin.close()
            self.proc.wait(timeout=5)
        except Exception:
            self.proc.kill()


def call_tool(mcp, name, args=None):
    resp = mcp.send("tools/call", {"name": name, "arguments": args or {}})
    result = resp.get("result", {})
    text = "".join(c.get("text", "") for c in result.get("content", []))
    return result, text


print("\n=== HANDSHAKE ===")
mcp = MCP({
    "HRIS_API_URL": "http://127.0.0.1:8080",
    "HRIS_MCP_EMAIL": "hr@perusahaan.com",
    "HRIS_MCP_PASSWORD": "rahasia-sekali-123",
})

resp = mcp.send("initialize", {
    "protocolVersion": PROTOCOL_VERSION,
    "capabilities": {},
    "clientInfo": {"name": "test-harness", "version": "1.0"},
})
check("initialize dijawab", "result" in resp, str(resp)[:200])
info = resp.get("result", {}).get("serverInfo", {})
check("serverInfo menyebut nama server", info.get("name") == "hris-face", str(info))
check("server mengumumkan kemampuan tools",
      "tools" in resp.get("result", {}).get("capabilities", {}), str(resp.get("result", {}).get("capabilities")))

mcp.send("notifications/initialized", {}, notify=True)

print("\n=== DAFTAR TOOL ===")
resp = mcp.send("tools/list")
tools = {t["name"]: t for t in resp.get("result", {}).get("tools", [])}
expected = [
    "attendance_today", "list_attendance", "list_leave_requests",
    "leave_balance", "list_employees", "list_corrections", "list_master_data",
]
for name in expected:
    check(f"tool '{name}' terdaftar", name in tools, f"tersedia: {sorted(tools)}")

check("setiap tool punya deskripsi",
      all(tools[n].get("description") for n in tools), "ada tool tanpa deskripsi")
check("tool berparameter punya inputSchema",
      "properties" in (tools.get("list_attendance", {}).get("inputSchema") or {}),
      str(tools.get("list_attendance", {}).get("inputSchema"))[:200])

# Read-only is the guarantee that matters most: no tool may look like a mutation.
mutating = [n for n in tools if any(
    v in n for v in ("create", "update", "delete", "approve", "reject", "reset", "edit", "set_"))]
check("tidak ada tool yang bisa mengubah data", not mutating, f"ditemukan: {mutating}")

print("\n=== PANGGIL TOOL ===")
result, text = call_tool(mcp, "attendance_today")
check("attendance_today berhasil", not result.get("isError"), text[:200])
today = json.loads(text) if text.startswith("{") else {}
check("ringkasan memuat total_active", "total_active" in today, text[:200])
check("ringkasan memuat on_leave", "on_leave" in today, text[:200])

result, text = call_tool(mcp, "list_employees")
check("list_employees berhasil", not result.get("isError"), text[:200])
data = json.loads(text)
check("mengembalikan karyawan hasil seed", data.get("jumlah", 0) >= 4, text[:200])
check("departemen tampil sebagai nama, bukan uuid",
      any(r.get("departemen") in ("Produksi", "Keuangan", "IT") for r in data.get("data", [])),
      text[:300])

result, text = call_tool(mcp, "list_employees", {"department": "IT"})
check("saring per departemen berhasil", not result.get("isError"), text[:200])
data = json.loads(text)
check("hasil saring hanya departemen itu",
      all(r["departemen"] == "IT" for r in data.get("data", [])), text[:300])

result, text = call_tool(mcp, "list_employees", {"department": "Tidak Ada"})
check("departemen tak dikenal -> error yang menjelaskan pilihan",
      result.get("isError") and "tersedia" in text.lower(), text[:200])

result, text = call_tool(mcp, "list_leave_requests", {"status": "pending"})
check("list_leave_requests berhasil", not result.get("isError"), text[:200])
check("ada pengajuan cuti pending", json.loads(text).get("jumlah", 0) >= 1, text[:200])

result, text = call_tool(mcp, "leave_balance", {"employee": "Budi"})
check("leave_balance per nama berhasil", not result.get("isError"), text[:200])
bal = json.loads(text)
check("saldo memuat kuota/terpakai/sisa",
      all(k in bal for k in ("kuota", "terpakai", "sisa")), text[:200])
check("sisa = kuota - terpakai", bal.get("sisa") == bal.get("kuota", 0) - bal.get("terpakai", 0), text[:200])

result, text = call_tool(mcp, "leave_balance", {"employee": "Tidak Ada Orang"})
check("karyawan tak ditemukan -> pesan jelas",
      result.get("isError") and "tidak ditemukan" in text.lower(), text[:200])

result, text = call_tool(mcp, "list_attendance", {"from": "2026-09-01", "to": "2026-09-30"})
check("list_attendance dengan rentang berhasil", not result.get("isError"), text[:200])
check("periode ikut dikembalikan", "periode" in json.loads(text), text[:200])

result, text = call_tool(mcp, "list_attendance", {"from": "2026-09-30", "to": "2026-09-01"})
check("rentang terbalik -> error, bukan hasil kosong",
      result.get("isError") and "terbalik" in text.lower(), text[:200])

result, text = call_tool(mcp, "list_attendance", {"from": "30-09-2026"})
check("format tanggal salah -> error yang menyebut format benar",
      result.get("isError") and "YYYY-MM-DD" in text, text[:200])

result, text = call_tool(mcp, "list_master_data")
check("list_master_data berhasil", not result.get("isError"), text[:200])
md = json.loads(text)
check("master data memuat tiga bagian",
      all(k in md for k in ("departemen", "lokasi_kerja", "jadwal_kerja")), text[:200])

result, text = call_tool(mcp, "list_corrections", {"status": "pending"})
check("list_corrections berhasil", not result.get("isError"), text[:200])

mcp.close()

print("\n=== KREDENSIAL SALAH ===")
bad = MCP({
    "HRIS_API_URL": "http://127.0.0.1:8080",
    "HRIS_MCP_EMAIL": "hr@perusahaan.com",
    "HRIS_MCP_PASSWORD": "password-salah",
})
bad.send("initialize", {"protocolVersion": PROTOCOL_VERSION, "capabilities": {},
                        "clientInfo": {"name": "t", "version": "1"}})
bad.send("notifications/initialized", {}, notify=True)
result, text = call_tool(bad, "attendance_today")
check("kredensial salah -> tool error, server tidak crash",
      result.get("isError") and "ditolak" in text.lower(), text[:200])
bad.close()

print(f"\n{'='*46}\nPASS: {passed}   FAIL: {failed}\n{'='*46}")
sys.exit(1 if failed else 0)
