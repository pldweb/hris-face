package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const dateLayout = "2006-01-02"

// Every tool here is read-only by construction: the client exposes no write
// verb at all, so a mis-parsed instruction cannot approve leave or change a
// record. Widening that is a deliberate change, not a slip.
func New(c *Client, version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "hris-face",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: "Data kehadiran, cuti, dan karyawan dari HRIS Face Attendance. " +
			"Semua tool hanya membaca; tidak ada yang bisa mengubah data. " +
			"Tanggal memakai format YYYY-MM-DD dan zona waktu Asia/Jakarta.",
	})

	registerAttendanceTools(s, c)
	registerLeaveTools(s, c)
	registerPeopleTools(s, c)
	return s
}

// textResult renders a value as pretty JSON. Returning JSON rather than prose
// keeps the model's input unambiguous when a caller asks it to compute totals.
func textResult(v any) (*mcp.CallToolResult, any, error) {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(body)}},
	}, nil, nil
}

func errResult(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}, nil, nil
}

// ---------------------------------------------------------------- attendance

type todayInput struct{}

type attendanceInput struct {
	From       string `json:"from,omitempty" jsonschema:"awal periode, format YYYY-MM-DD. Kosong berarti awal bulan berjalan."`
	To         string `json:"to,omitempty" jsonschema:"akhir periode, format YYYY-MM-DD. Kosong berarti hari ini."`
	Status     string `json:"status,omitempty" jsonschema:"saring status: on_time, late, atau early_leave."`
	Department string `json:"department,omitempty" jsonschema:"nama departemen, dicocokkan tanpa membedakan huruf besar/kecil."`
	Limit      int    `json:"limit,omitempty" jsonschema:"jumlah baris maksimum (default 100, maksimal 500)."`
}

func registerAttendanceTools(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "attendance_today",
		Description: "Ringkasan kehadiran hari ini: jumlah karyawan aktif, tepat waktu, terlambat, " +
			"belum absen, sudah pulang, dan sedang cuti. Pakai ini untuk pertanyaan seperti " +
			"'berapa yang telat hari ini?' atau 'siapa saja yang belum absen?'.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ todayInput) (*mcp.CallToolResult, any, error) {
		var summary map[string]any
		if err := c.get("/admin/attendances/today", nil, &summary); err != nil {
			return errResult(err)
		}
		return textResult(summary)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "list_attendance",
		Description: "Daftar catatan kehadiran (absen masuk/pulang) pada rentang tanggal, " +
			"bisa disaring per status dan departemen. Untuk rekap bulanan, isi from = tanggal 1 " +
			"dan to = hari ini.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in attendanceInput) (*mcp.CallToolResult, any, error) {
		params := url.Values{}

		from, to, err := resolveRange(in.From, in.To)
		if err != nil {
			return errResult(err)
		}
		params.Set("from", from)
		params.Set("to", to)

		if in.Status != "" {
			params.Set("status", in.Status)
		}
		if in.Department != "" {
			id, err := departmentIDByName(c, in.Department)
			if err != nil {
				return errResult(err)
			}
			params.Set("department_id", id)
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 100
		}
		if limit > 500 {
			limit = 500
		}
		params.Set("limit", strconv.Itoa(limit))

		var out struct {
			Data []map[string]any `json:"data"`
		}
		if err := c.get("/admin/attendances", params, &out); err != nil {
			return errResult(err)
		}
		return textResult(map[string]any{
			"periode": map[string]string{"from": from, "to": to},
			"jumlah":  len(out.Data),
			"data":    out.Data,
		})
	})
}

// resolveRange defaults to month-to-date, which is the period HR asks about
// most, and validates order so a reversed range fails loudly instead of
// silently returning nothing.
func resolveRange(fromStr, toStr string) (string, string, error) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := now

	if fromStr != "" {
		parsed, err := time.Parse(dateLayout, fromStr)
		if err != nil {
			return "", "", fmt.Errorf("tanggal 'from' tidak valid: %q, pakai format YYYY-MM-DD", fromStr)
		}
		from = parsed
	}
	if toStr != "" {
		parsed, err := time.Parse(dateLayout, toStr)
		if err != nil {
			return "", "", fmt.Errorf("tanggal 'to' tidak valid: %q, pakai format YYYY-MM-DD", toStr)
		}
		to = parsed
	}
	if to.Before(from) {
		return "", "", fmt.Errorf("rentang terbalik: 'from' (%s) setelah 'to' (%s)",
			from.Format(dateLayout), to.Format(dateLayout))
	}
	return from.Format(dateLayout), to.Format(dateLayout), nil
}

// --------------------------------------------------------------------- leave

type leaveInput struct {
	Status string `json:"status,omitempty" jsonschema:"saring status: pending, approved, atau rejected."`
	Type   string `json:"type,omitempty" jsonschema:"saring jenis: annual (cuti tahunan), sick (sakit), atau permit (izin)."`
}

type balanceInput struct {
	Employee string `json:"employee" jsonschema:"nama atau NIK karyawan."`
	Year     int    `json:"year,omitempty" jsonschema:"tahun kuota, default tahun berjalan."`
}

func registerLeaveTools(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "list_leave_requests",
		Description: "Daftar pengajuan cuti/izin/sakit beserta statusnya. Pakai status=pending " +
			"untuk melihat yang masih menunggu persetujuan HR.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in leaveInput) (*mcp.CallToolResult, any, error) {
		params := url.Values{}
		if in.Status != "" {
			params.Set("status", in.Status)
		}
		if in.Type != "" {
			params.Set("type", in.Type)
		}
		var out struct {
			Data []map[string]any `json:"data"`
		}
		if err := c.get("/admin/leave-requests", params, &out); err != nil {
			return errResult(err)
		}
		return textResult(map[string]any{"jumlah": len(out.Data), "data": out.Data})
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "leave_balance",
		Description: "Sisa kuota cuti tahunan seorang karyawan: kuota, terpakai, dan sisa. " +
			"Hanya cuti tahunan yang memotong kuota; sakit dan izin tidak.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in balanceInput) (*mcp.CallToolResult, any, error) {
		emp, err := findEmployee(c, in.Employee)
		if err != nil {
			return errResult(err)
		}
		year := in.Year
		if year == 0 {
			year = time.Now().Year()
		}

		// There is no admin-side balance endpoint, so derive it the same way
		// the API does: quota on the employee minus approved annual days.
		params := url.Values{}
		params.Set("status", "approved")
		params.Set("type", "annual")
		var out struct {
			Data []struct {
				EmployeeID string `json:"employee_id"`
				StartDate  string `json:"start_date"`
				DaysCount  int    `json:"days_count"`
			} `json:"data"`
		}
		if err := c.get("/admin/leave-requests", params, &out); err != nil {
			return errResult(err)
		}
		used := 0
		for _, r := range out.Data {
			if r.EmployeeID == emp.ID && strings.HasPrefix(r.StartDate, strconv.Itoa(year)) {
				used += r.DaysCount
			}
		}
		return textResult(map[string]any{
			"karyawan":   emp.FullName,
			"nik":        emp.NIK,
			"tahun":      year,
			"kuota":      emp.AnnualLeaveQuota,
			"terpakai":   used,
			"sisa":       emp.AnnualLeaveQuota - used,
			"keterangan": "Hanya cuti tahunan (annual) yang memotong kuota.",
		})
	})
}

// -------------------------------------------------------------------- people

type employeesInput struct {
	Status     string `json:"status,omitempty" jsonschema:"saring status: active, inactive, atau pending_enrollment."`
	Department string `json:"department,omitempty" jsonschema:"nama departemen."`
	Search     string `json:"search,omitempty" jsonschema:"cari berdasarkan potongan nama atau NIK."`
}

type emptyInput struct{}

type employee struct {
	ID               string `json:"id"`
	NIK              string `json:"nik"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	DepartmentID     string `json:"department_id"`
	LocationID       string `json:"location_id"`
	AnnualLeaveQuota int    `json:"annual_leave_quota"`
	Status           string `json:"status"`
}

func registerPeopleTools(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "list_employees",
		Description: "Daftar karyawan beserta NIK, email, departemen, lokasi, status, dan kuota cuti. " +
			"Bisa disaring per status/departemen atau dicari per nama/NIK.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in employeesInput) (*mcp.CallToolResult, any, error) {
		employees, err := listEmployees(c)
		if err != nil {
			return errResult(err)
		}
		departments, locations, err := lookups(c)
		if err != nil {
			return errResult(err)
		}

		var wantDept string
		if in.Department != "" {
			id, err := departmentIDByName(c, in.Department)
			if err != nil {
				return errResult(err)
			}
			wantDept = id
		}

		rows := make([]map[string]any, 0, len(employees))
		for _, e := range employees {
			if in.Status != "" && e.Status != in.Status {
				continue
			}
			if wantDept != "" && e.DepartmentID != wantDept {
				continue
			}
			if in.Search != "" {
				q := strings.ToLower(in.Search)
				if !strings.Contains(strings.ToLower(e.FullName), q) && !strings.Contains(strings.ToLower(e.NIK), q) {
					continue
				}
			}
			rows = append(rows, map[string]any{
				"nik":                e.NIK,
				"nama":               e.FullName,
				"email":              e.Email,
				"departemen":         departments[e.DepartmentID],
				"lokasi":             locations[e.LocationID],
				"status":             e.Status,
				"kuota_cuti_tahunan": e.AnnualLeaveQuota,
			})
		}
		return textResult(map[string]any{"jumlah": len(rows), "data": rows})
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_corrections",
		Description: "Daftar pengajuan koreksi absen dari karyawan (misal lupa absen atau wajah tidak terdeteksi).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in leaveInput) (*mcp.CallToolResult, any, error) {
		params := url.Values{}
		if in.Status != "" {
			params.Set("status", in.Status)
		}
		var out struct {
			Data []map[string]any `json:"data"`
		}
		if err := c.get("/admin/corrections", params, &out); err != nil {
			return errResult(err)
		}
		return textResult(map[string]any{"jumlah": len(out.Data), "data": out.Data})
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_master_data",
		Description: "Daftar departemen, lokasi kerja, dan jadwal kerja yang terdaftar di sistem.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		var depts struct {
			Data []map[string]any `json:"data"`
		}
		if err := c.get("/admin/departments", nil, &depts); err != nil {
			return errResult(err)
		}
		var locs struct {
			Data []map[string]any `json:"data"`
		}
		if err := c.get("/admin/locations", nil, &locs); err != nil {
			return errResult(err)
		}
		var scheds struct {
			Data []map[string]any `json:"data"`
		}
		if err := c.get("/admin/schedules", nil, &scheds); err != nil {
			return errResult(err)
		}
		return textResult(map[string]any{
			"departemen":   depts.Data,
			"lokasi_kerja": locs.Data,
			"jadwal_kerja": scheds.Data,
		})
	})
}

// ------------------------------------------------------------------- helpers

func listEmployees(c *Client) ([]employee, error) {
	var out struct {
		Data []employee `json:"data"`
	}
	if err := c.get("/admin/employees", nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// findEmployee resolves a human-typed name or NIK. An ambiguous match is an
// error rather than a guess: silently picking one of two "Budi"s would answer
// a question about the wrong person's leave balance.
func findEmployee(c *Client, query string) (*employee, error) {
	employees, err := listEmployees(c)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, fmt.Errorf("nama atau NIK karyawan wajib diisi")
	}

	var matches []employee
	for _, e := range employees {
		if strings.EqualFold(e.NIK, q) {
			return &e, nil
		}
		if strings.Contains(strings.ToLower(e.FullName), q) {
			matches = append(matches, e)
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("karyawan %q tidak ditemukan", query)
	case 1:
		return &matches[0], nil
	default:
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, fmt.Sprintf("%s (NIK %s)", m.FullName, m.NIK))
		}
		return nil, fmt.Errorf("%q cocok dengan lebih dari satu karyawan: %s -- sebutkan NIK-nya",
			query, strings.Join(names, ", "))
	}
}

func departmentIDByName(c *Client, name string) (string, error) {
	var out struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := c.get("/admin/departments", nil, &out); err != nil {
		return "", err
	}
	available := make([]string, 0, len(out.Data))
	for _, d := range out.Data {
		if strings.EqualFold(d.Name, name) {
			return d.ID, nil
		}
		available = append(available, d.Name)
	}
	return "", fmt.Errorf("departemen %q tidak ada. Yang tersedia: %s", name, strings.Join(available, ", "))
}

// lookups maps ids to names so tool output shows "Produksi" instead of a uuid
// the caller cannot interpret.
func lookups(c *Client) (map[string]string, map[string]string, error) {
	var depts struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := c.get("/admin/departments", nil, &depts); err != nil {
		return nil, nil, err
	}
	var locs struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := c.get("/admin/locations", nil, &locs); err != nil {
		return nil, nil, err
	}

	d := make(map[string]string, len(depts.Data))
	for _, x := range depts.Data {
		d[x.ID] = x.Name
	}
	l := make(map[string]string, len(locs.Data))
	for _, x := range locs.Data {
		l[x.ID] = x.Name
	}
	return d, l, nil
}
