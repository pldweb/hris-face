package masterdata

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hris-face/api/internal/middleware"
)

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	hr := middleware.RequireRole("hr", "superadmin")
	r.GET("/admin/schedules", hr, listSchedulesHandler(svc))
	r.POST("/admin/schedules", hr, createScheduleHandler(svc))
	r.PUT("/admin/schedules/:id", hr, updateScheduleHandler(svc))
	r.GET("/admin/locations", hr, listLocationsHandler(svc))
	r.POST("/admin/locations", hr, createLocationHandler(svc))
	r.PUT("/admin/locations/:id", hr, updateLocationHandler(svc))
	r.DELETE("/admin/locations/:id", hr, deleteLocationHandler(svc))
	r.PUT("/admin/employees/:id/schedule", hr, assignScheduleHandler(svc))
}

func listSchedulesHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := svc.ListSchedules(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil jadwal"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func createScheduleHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in Schedule
		if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" || in.StartTime == "" || in.EndTime == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nama, jam masuk, dan jam pulang wajib diisi"})
			return
		}
		out, err := svc.CreateSchedule(c.Request.Context(), in)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat jadwal"})
			return
		}
		c.JSON(http.StatusCreated, out)
	}
}

func updateScheduleHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in Schedule
		if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "data jadwal tidak valid"})
			return
		}
		err := svc.UpdateSchedule(c.Request.Context(), c.Param("id"), in)
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui jadwal"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func listLocationsHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := svc.ListLocations(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil lokasi"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

type nameRequest struct {
	Name string `json:"name" binding:"required"`
}

type locationRequest struct {
	Name string `json:"name" binding:"required"`
	// SetGeofence must be true for Lat/Lng/RadiusMeters to be written at all --
	// see Service.UpdateLocation. A caller that only cares about the name (the
	// employee-form quick-add) omits this and the three fields entirely.
	SetGeofence  bool     `json:"set_geofence"`
	Lat          *float64 `json:"lat"`
	Lng          *float64 `json:"lng"`
	RadiusMeters *int     `json:"radius_meters"`
}

func createLocationHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req locationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nama lokasi wajib diisi"})
			return
		}
		out, err := svc.CreateLocation(c.Request.Context(), req.Name, req.Lat, req.Lng, req.RadiusMeters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat lokasi"})
			return
		}
		c.JSON(http.StatusCreated, out)
	}
}

func updateLocationHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req locationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nama lokasi wajib diisi"})
			return
		}
		err := svc.UpdateLocation(c.Request.Context(), c.Param("id"), req.Name, req.SetGeofence, req.Lat, req.Lng, req.RadiusMeters)
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui lokasi"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func deleteLocationHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := svc.DeleteLocation(c.Request.Context(), c.Param("id"))
		switch {
		case err == nil:
			c.JSON(http.StatusOK, gin.H{"status": "deleted"})
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrLocationInUse):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus lokasi"})
		}
	}
}

type assignRequest struct {
	ScheduleID string `json:"schedule_id" binding:"required"`
}

func assignScheduleHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req assignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_id wajib diisi"})
			return
		}
		err := svc.AssignSchedule(c.Request.Context(), c.Param("id"), req.ScheduleID)
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menetapkan jadwal"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
