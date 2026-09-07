package leave

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hris-face/api/internal/middleware"
)

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	r.POST("/leave-requests", createHandler(svc))
	r.GET("/leave-requests/me", myListHandler(svc))
	r.GET("/leave-requests/me/balance", myBalanceHandler(svc))

	hr := middleware.RequireRole("hr", "superadmin")
	r.GET("/admin/leave-requests", hr, listHandler(svc))
	r.PATCH("/admin/leave-requests/:id", hr, reviewHandler(svc))
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNoEmployeeRecord), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidType), errors.Is(err, ErrInvalidRange), errors.Is(err, ErrNoWorkdays), errors.Is(err, ErrQuotaExceeded):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrOverlap), errors.Is(err, ErrAlreadyReviewed):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memproses pengajuan cuti"})
	}
}

type createRequest struct {
	Type      string `json:"type" binding:"required,oneof=annual sick permit"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
}

func createHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type, start_date, end_date, dan reason wajib diisi"})
			return
		}
		out, err := svc.Create(c.Request.Context(), c.GetString("user_id"), req.Type, req.StartDate, req.EndDate, req.Reason)
		if err != nil {
			writeErr(c, err)
			return
		}
		c.JSON(http.StatusCreated, out)
	}
}

func myListHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := svc.ListMine(c.Request.Context(), c.GetString("user_id"), c.Query("status"))
		if err != nil {
			writeErr(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func myBalanceHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		year := time.Now().Year()
		if v := c.Query("year"); v != "" {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "year tidak valid"})
				return
			}
			year = parsed
		}

		employeeID, err := svc.employeeIDForUser(c.Request.Context(), c.GetString("user_id"))
		if err != nil {
			writeErr(c, err)
			return
		}
		bal, err := svc.Balance(c.Request.Context(), employeeID, year)
		if err != nil {
			writeErr(c, err)
			return
		}
		c.JSON(http.StatusOK, bal)
	}
}

func listHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := svc.ListAdmin(c.Request.Context(), c.Query("status"), c.Query("type"),
			c.Query("employee_id"), c.Query("from"), c.Query("to"))
		if err != nil {
			writeErr(c, err)
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
		out, err := svc.Review(c.Request.Context(), c.Param("id"), c.GetString("user_id"), req.Decision, req.Note)
		if err != nil {
			writeErr(c, err)
			return
		}
		c.JSON(http.StatusOK, out)
	}
}
