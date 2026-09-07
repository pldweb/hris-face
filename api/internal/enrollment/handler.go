package enrollment

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	r.POST("/enrollment/photos", uploadHandler(svc, false))
	r.POST("/enrollment/re-enroll", uploadHandler(svc, true))
}

func uploadHandler(svc *Service, replace bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "form tidak valid"})
			return
		}

		files := form.File["photos"]
		photos := make([][]byte, 0, len(files))
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca foto"})
				return
			}
			data, err := io.ReadAll(io.LimitReader(f, 5<<20))
			f.Close()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca foto"})
				return
			}
			photos = append(photos, data)
		}

		userID := c.GetString("user_id")
		if replace {
			err = svc.ReEnroll(c.Request.Context(), userID, photos)
		} else {
			err = svc.Enroll(c.Request.Context(), userID, photos)
		}
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"status": "active"})
			return
		}

		var rejected *RejectedError
		if errors.As(err, &rejected) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"rejections": rejected.Rejections})
			return
		}

		switch {
		case errors.Is(err, ErrTooFewPhotos), errors.Is(err, ErrTooManyPhotos):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrEmployeeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrDuplicateFace), errors.Is(err, ErrAlreadyEnrolled):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memproses enrollment"})
		}
	}
}
