package correction

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hris-face/api/internal/middleware"
)

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	r.POST("/corrections", createHandler(svc))
	r.GET("/corrections/me", myListHandler(svc))
	hr := middleware.RequireRole("hr", "superadmin")
	r.GET("/admin/corrections", hr, listHandler(svc))
	r.PATCH("/admin/corrections/:id", hr, reviewHandler(svc))
}

type createRequest struct {
	RequestedType string `json:"requested_type" binding:"required"`
	RequestedTime string `json:"requested_time" binding:"required"`
	Reason        string `json:"reason" binding:"required,min=5"`
}

func createHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tipe, waktu, dan alasan (min. 5 karakter) wajib diisi"})
			return
		}
		when, err := time.Parse(time.RFC3339, req.RequestedTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format waktu tidak valid"})
			return
		}

		out, err := svc.Create(c.Request.Context(), c.GetString("user_id"), req.RequestedType, when, req.Reason)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidType):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, ErrNoEmployeeRecord):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengajukan koreksi"})
			}
			return
		}
		c.JSON(http.StatusCreated, out)
	}
}

func myListHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := svc.List(c.Request.Context(), c.Query("status"), c.GetString("user_id"), true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil pengajuan"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func listHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := svc.List(c.Request.Context(), c.Query("status"), "", false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil pengajuan"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

type reviewRequest struct {
	Decision string `json:"decision" binding:"required,oneof=approve reject"`
	Note     string `json:"note"`
}

func reviewHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req reviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "decision harus approve atau reject"})
			return
		}

		err := svc.Review(c.Request.Context(), c.Param("id"), c.GetString("user_id"), req.Decision, req.Note)
		switch {
		case err == nil:
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrAlreadyReviewed):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memproses pengajuan"})
		}
	}
}
