package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const refreshCookieName = "refresh_token"
const refreshCookiePath = "/api/v1/auth"

// Five failed attempts per account per 15 minutes. Slow enough to make guessing
// pointless, loose enough that a person who forgot which password they set is
// not locked out for the day.
var loginThrottle = newThrottle(5, 15*time.Minute)

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	r.POST("/auth/login", loginHandler(svc))
	r.POST("/auth/refresh", refreshHandler(svc))
	r.POST("/auth/logout", logoutHandler(svc))
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func loginHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email dan password wajib diisi"})
			return
		}

		if loginThrottle.blocked(req.Email) {
			c.JSON(http.StatusTooManyRequests,
				gin.H{"error": "Terlalu banyak percobaan gagal untuk akun ini. Coba lagi dalam 15 menit."})
			return
		}

		pair, err := svc.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, ErrInvalidCredentials) {
				loginThrottle.recordFailure(req.Email)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "email atau password salah"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "terjadi kesalahan"})
			return
		}
		loginThrottle.reset(req.Email)

		setRefreshCookie(c, pair.RefreshToken, pair.RefreshTTL)
		c.JSON(http.StatusOK, gin.H{"access_token": pair.AccessToken})
	}
}

func refreshHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		presented, err := c.Cookie(refreshCookieName)
		if err != nil || presented == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "sesi berakhir, silakan masuk kembali"})
			return
		}

		pair, err := svc.Refresh(c.Request.Context(), presented)
		if err != nil {
			clearRefreshCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "sesi berakhir, silakan masuk kembali"})
			return
		}

		setRefreshCookie(c, pair.RefreshToken, pair.RefreshTTL)
		c.JSON(http.StatusOK, gin.H{"access_token": pair.AccessToken})
	}
}

func logoutHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if presented, err := c.Cookie(refreshCookieName); err == nil && presented != "" {
			_ = svc.Logout(c.Request.Context(), presented)
		}
		clearRefreshCookie(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func setRefreshCookie(c *gin.Context, token string, ttl time.Duration) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, token, int(ttl.Seconds()), refreshCookiePath, "", true, true)
}

func clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, refreshCookiePath, "", true, true)
}
