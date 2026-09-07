package employee

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hris-face/api/internal/middleware"
)

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	r.GET("/me", meHandler(svc))
	r.GET("/admin/employees", middleware.RequireRole("hr", "superadmin"), listHandler(svc))
	r.POST("/admin/employees", middleware.RequireRole("hr", "superadmin"), createHandler(svc))
	r.PUT("/admin/employees/:id", middleware.RequireRole("hr", "superadmin"), updateHandler(svc))
	r.DELETE("/admin/employees/:id", middleware.RequireRole("hr", "superadmin"), deactivateHandler(svc))
	r.POST("/admin/employees/import", middleware.RequireRole("hr", "superadmin"), importHandler(svc))
	r.GET("/admin/departments", middleware.RequireRole("hr", "superadmin"), listDepartmentsHandler(svc))
	r.POST("/admin/departments", middleware.RequireRole("hr", "superadmin"), createDepartmentHandler(svc))
	r.GET("/admin/devices", middleware.RequireRole("hr", "superadmin"), listDevicesHandler(svc))
	r.POST("/admin/devices/:id/approve", middleware.RequireRole("hr", "superadmin"), approveDeviceHandler(svc))
}

func meHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		me, err := svc.Me(c.Request.Context(), c.GetString("user_id"))
		if err != nil {
			if errors.Is(err, ErrNoEmployeeRecord) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil profil"})
			return
		}
		c.JSON(http.StatusOK, me)
	}
}

func listHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		employees, err := svc.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data karyawan"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": employees})
	}
}

type createEmployeeRequest struct {
	NIK          string  `json:"nik" binding:"required"`
	FullName     string  `json:"full_name" binding:"required"`
	Email        string  `json:"email" binding:"required,email"`
	DepartmentID *string `json:"department_id"`
	LocationID   *string `json:"location_id"`
}

func createHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createEmployeeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "data karyawan tidak lengkap atau tidak valid"})
			return
		}

		result, err := svc.Create(c.Request.Context(), CreateEmployeeInput{
			NIK: req.NIK, FullName: req.FullName, Email: req.Email,
			DepartmentID: req.DepartmentID, LocationID: req.LocationID,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrEmailTaken):
				c.JSON(http.StatusConflict, gin.H{"error": "email sudah terdaftar"})
			case errors.Is(err, ErrNIKTaken):
				c.JSON(http.StatusConflict, gin.H{"error": "NIK sudah terdaftar"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat karyawan"})
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"employee":      result.Employee,
			"temp_password": result.TempPassword,
		})
	}
}

type updateEmployeeRequest struct {
	FullName     string  `json:"full_name" binding:"required"`
	Email        string  `json:"email" binding:"required,email"`
	DepartmentID *string `json:"department_id"`
	LocationID   *string `json:"location_id"`
}

func updateHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req updateEmployeeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nama dan email wajib diisi"})
			return
		}

		err := svc.Update(c.Request.Context(), c.Param("id"), UpdateEmployeeInput{
			FullName: req.FullName, Email: req.Email,
			DepartmentID: req.DepartmentID, LocationID: req.LocationID,
		})
		switch {
		case err == nil:
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		case errors.Is(err, ErrEmployeeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui karyawan"})
		}
	}
}

// deactivateHandler is the "Hapus" action on the employee list. It sets the
// employee inactive rather than deleting the row -- see Service.Deactivate.
func deactivateHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := svc.Deactivate(c.Request.Context(), c.Param("id"))
		if errors.Is(err, ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menonaktifkan karyawan"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "inactive"})
	}
}

func listDepartmentsHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		departments, err := svc.ListDepartments(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data departemen"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": departments})
	}
}

type createDepartmentRequest struct {
	Name string `json:"name" binding:"required"`
}

func createDepartmentHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createDepartmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nama departemen wajib diisi"})
			return
		}
		dept, err := svc.CreateDepartment(c.Request.Context(), req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat departemen"})
			return
		}
		c.JSON(http.StatusCreated, dept)
	}
}

func listDevicesHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		devices, err := svc.ListDevices(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data perangkat"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": devices})
	}
}

func approveDeviceHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := svc.ApproveDevice(c.Request.Context(), c.Param("id"))
		if errors.Is(err, ErrDeviceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyetujui perangkat"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "approved"})
	}
}

func importHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file CSV wajib diunggah"})
			return
		}
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca file"})
			return
		}
		defer f.Close()

		result, err := svc.ImportCSV(c.Request.Context(), f)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
