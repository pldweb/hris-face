package report

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"github.com/hris-face/api/internal/middleware"
)

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	hr := middleware.RequireRole("hr", "superadmin")
	r.GET("/admin/attendances", hr, listHandler(svc))
	r.GET("/admin/attendances/today", hr, todayHandler(svc))
	r.DELETE("/admin/attendances/:id", hr, deleteHandler(svc))
	r.PUT("/admin/attendances/:id", hr, updateHandler(svc))
	r.GET("/admin/reports/export", hr, exportHandler(svc))
	r.GET("/admin/reports/export.xlsx", hr, exportXlsxHandler(svc))
	r.GET("/attendance/me", meHistoryHandler(svc))
	r.GET("/team/attendances", middleware.RequireRole("manager", "hr", "superadmin"), teamHandler(svc))
	r.GET("/team/today", middleware.RequireRole("manager", "hr", "superadmin"), teamTodayHandler(svc))
}

func deleteHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := svc.Delete(c.Request.Context(), c.Param("id"), c.GetString("user_id"))
		if errors.Is(err, ErrAttendanceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			log.Printf("delete attendance %s: %v", c.Param("id"), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus absensi"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}

type updateAttendanceRequest struct {
	OccurredAt string `json:"occurred_at" binding:"required"`
}

func updateHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req updateAttendanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "occurred_at wajib diisi"})
			return
		}
		when, err := time.Parse(time.RFC3339, req.OccurredAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format waktu tidak valid"})
			return
		}

		row, err := svc.Update(c.Request.Context(), c.Param("id"), c.GetString("user_id"), when)
		if errors.Is(err, ErrAttendanceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			log.Printf("update attendance %s: %v", c.Param("id"), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui absensi"})
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

func filterFrom(c *gin.Context) Filter {
	f := Filter{
		DepartmentID: c.Query("department_id"),
		EmployeeID:   c.Query("employee_id"),
		Status:       c.Query("status"),
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.From = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			// Inclusive end date: the caller means "through this day".
			end := t.AddDate(0, 0, 1)
			f.To = &end
		}
	}
	f.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
	f.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	return f
}

func listHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, total, err := svc.List(c.Request.Context(), filterFrom(c))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data absensi"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": rows, "total": total})
	}
}

func todayHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		sum, err := svc.Today(c.Request.Context(), "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil ringkasan hari ini"})
			return
		}
		c.JSON(http.StatusOK, sum)
	}
}

// exportHandler streams CSV honouring the same filter as the table on screen.
func exportHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := svc.All(c.Request.Context(), filterFrom(c))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data absensi"})
			return
		}

		filename := fmt.Sprintf("absensi-%s.csv", time.Now().In(svc.timezone).Format("2006-01-02"))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)

		// Excel in Indonesian locales opens UTF-8 CSV as mojibake without a BOM.
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) //nolint:errcheck

		w := csv.NewWriter(c.Writer)
		defer w.Flush()
		w.Write([]string{"NIK", "Nama", "Departemen", "Tipe", "Tanggal", "Jam", "Status", "Sumber", "Perlu Review"}) //nolint:errcheck
		for _, r := range rows {
			local := r.OccurredAt.In(svc.timezone)
			w.Write([]string{ //nolint:errcheck
				r.NIK, r.FullName, r.Department, typeLabel(r.Type),
				local.Format("2006-01-02"), local.Format("15:04:05"),
				statusLabel(r.Status), r.Source, boolLabel(r.LowConfidence),
			})
		}
	}
}

func meHistoryHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		month := c.DefaultQuery("month", time.Now().In(svc.timezone).Format("2006-01"))
		from, err := time.ParseInLocation("2006-01", month, svc.timezone)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format bulan harus YYYY-MM"})
			return
		}
		to := from.AddDate(0, 1, 0)
		if now := time.Now().In(svc.timezone); to.After(now) {
			to = now.AddDate(0, 0, 1)
		}

		records, err := svc.MyHistory(c.Request.Context(), c.GetString("user_id"), from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil riwayat"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": records, "month": month})
	}
}

func teamHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		managerID, err := svc.employeeIDForUser(c.Request.Context(), c.GetString("user_id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "akun ini tidak terhubung ke data karyawan"})
			return
		}
		f := filterFrom(c)
		f.ManagerID = managerID
		rows, total, err := svc.List(c.Request.Context(), f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data tim"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": rows, "total": total})
	}
}

func teamTodayHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		managerID, err := svc.employeeIDForUser(c.Request.Context(), c.GetString("user_id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "akun ini tidak terhubung ke data karyawan"})
			return
		}
		sum, err := svc.Today(c.Request.Context(), managerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil ringkasan tim"})
			return
		}
		c.JSON(http.StatusOK, sum)
	}
}

func typeLabel(t string) string {
	if t == "check_out" {
		return "Pulang"
	}
	return "Masuk"
}

func statusLabel(s string) string {
	switch s {
	case "on_time":
		return "Tepat waktu"
	case "late":
		return "Terlambat"
	case "early_leave":
		return "Pulang cepat"
	case "absent":
		return "Tidak hadir"
	}
	return s
}

func boolLabel(b bool) string {
	if b {
		return "Ya"
	}
	return "Tidak"
}

// exportXlsxHandler produces a real spreadsheet rather than CSV. It matters for
// the monthly recap: dates and times land as typed cells that Excel can sort
// and filter, instead of text that silently misorders 09:00 against 10:00.
func exportXlsxHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := svc.All(c.Request.Context(), filterFrom(c))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data absensi"})
			return
		}

		f := excelize.NewFile()
		defer f.Close()
		const sheet = "Absensi"
		f.SetSheetName(f.GetSheetName(0), sheet)

		headers := []string{"NIK", "Nama", "Departemen", "Tipe", "Tanggal", "Jam", "Status", "Sumber", "Perlu Review"}
		bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		dateStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 14}) // dd/mm/yyyy
		timeStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 21}) // hh:mm:ss

		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h) //nolint:errcheck
		}
		f.SetRowStyle(sheet, 1, 1, bold)   //nolint:errcheck
		f.SetPanes(sheet, &excelize.Panes{ //nolint:errcheck
			Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft",
		})
		f.AutoFilter(sheet, "A1:I1", nil) //nolint:errcheck

		for i, r := range rows {
			row := i + 2
			local := r.OccurredAt.In(svc.timezone)
			values := []any{
				r.NIK, r.FullName, r.Department, typeLabel(r.Type),
				local, local, statusLabel(r.Status), r.Source, boolLabel(r.LowConfidence),
			}
			for col, v := range values {
				cell, _ := excelize.CoordinatesToCellName(col+1, row)
				f.SetCellValue(sheet, cell, v) //nolint:errcheck
			}
			dateCell, _ := excelize.CoordinatesToCellName(5, row)
			timeCell, _ := excelize.CoordinatesToCellName(6, row)
			f.SetCellStyle(sheet, dateCell, dateCell, dateStyle) //nolint:errcheck
			f.SetCellStyle(sheet, timeCell, timeCell, timeStyle) //nolint:errcheck
		}

		for col, width := range map[string]float64{"A": 14, "B": 26, "C": 18, "D": 10, "E": 12, "F": 10, "G": 14, "H": 18, "I": 13} {
			f.SetColWidth(sheet, col, col, width) //nolint:errcheck
		}

		filename := fmt.Sprintf("absensi-%s.xlsx", time.Now().In(svc.timezone).Format("2006-01-02"))
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
		if err := f.Write(c.Writer); err != nil {
			log.Printf("export xlsx: %v", err)
		}
	}
}
