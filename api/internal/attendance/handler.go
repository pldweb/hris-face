package attendance

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hris-face/api/internal/middleware"
)

const deviceCookieName = "device_id"
const deviceCookieMaxAge = 5 * 365 * 24 * 3600 // device binding should outlive sessions

// parseLatLng returns nil for either value when it's missing or not a valid
// float -- a malformed coordinate is treated the same as "no location sent"
// (checkGeofence's ErrLocationRequired), not a 400: the browser's geolocation
// API failing silently must not read as a client bug.
func parseLatLng(latStr, lngStr string) (*float64, *float64) {
	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)
	if errLat != nil || errLng != nil {
		return nil, nil
	}
	return &lat, &lng
}

func RegisterRoutes(r gin.IRoutes, svc *Service) {
	r.POST("/attendance/check-in", markHandler(svc, "check_in"))
	r.POST("/attendance/check-out", markHandler(svc, "check_out"))
	r.POST("/attendance/challenge/check-in", challengeHandler(svc, "check_in"))
	r.POST("/attendance/challenge/check-out", challengeHandler(svc, "check_out"))
}

func RegisterScanRoute(r gin.IRoutes, svc *Service) {
	r.POST("/admin/face-scan", middleware.RequireRole("hr", "superadmin"), scanHandler(svc))
}

func scanHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		imageJPEG, err := io.ReadAll(io.LimitReader(c.Request.Body, 5<<20))
		if err != nil || len(imageJPEG) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gambar tidak valid"})
			return
		}
		result, err := svc.Scan(c.Request.Context(), imageJPEG)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memeriksa wajah"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// challengeHandler takes a multi-frame sequence for the case where the passive
// anti-spoof model was unsure about a single frame.
func challengeHandler(svc *Service, kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "form tidak valid"})
			return
		}

		files := form.File["frames"]
		frames := make([][]byte, 0, len(files))
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca frame"})
				return
			}
			data, err := io.ReadAll(io.LimitReader(f, 5<<20))
			f.Close()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membaca frame"})
				return
			}
			frames = append(frames, data)
		}

		deviceKey, _ := c.Cookie(deviceCookieName)
		lat, lng := parseLatLng(c.PostForm("lat"), c.PostForm("lng"))
		result, challenge, issuedKey, err := svc.RecordWithChallenge(
			c.Request.Context(), kind, c.GetString("user_id"), frames,
			deviceKey, c.Request.UserAgent(), c.ClientIP(), lat, lng)

		// The cookie is set whenever a device was resolved, BEFORE checking err:
		// resolveDevice can commit a new approved device row even when the
		// attendance write that follows it then fails a business rule. Setting
		// the cookie only on the success path silently orphaned that device --
		// the next request had no cookie to present, so it registered ANOTHER
		// device, burning through the two-device limit in a couple of failed
		// attempts even on the employee's own, unchanged laptop.
		if issuedKey != "" {
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(deviceCookieName, issuedKey, deviceCookieMaxAge, "/api/v1", "", true, true)
		}

		if err != nil {
			body := gin.H{"error": messageFor(err)}
			var wrongPerson *ErrWrongPerson
			if errors.As(err, &wrongPerson) {
				body["matched_name"] = wrongPerson.Name
			}
			c.JSON(statusFor(err), body)
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": result, "challenge": challenge})
	}
}

func markHandler(svc *Service, kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		imageJPEG, err := io.ReadAll(io.LimitReader(c.Request.Body, 5<<20)) // 5MB cap
		if err != nil || len(imageJPEG) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gambar tidak valid"})
			return
		}

		deviceKey, _ := c.Cookie(deviceCookieName)
		userAgent := c.Request.UserAgent()

		userID := c.GetString("user_id")
		lat, lng := parseLatLng(c.Query("lat"), c.Query("lng"))

		var result *Result
		var issuedKey string
		if kind == "check_in" {
			result, issuedKey, err = svc.CheckIn(c.Request.Context(), userID, imageJPEG, deviceKey, userAgent, c.ClientIP(), lat, lng)
		} else {
			result, issuedKey, err = svc.CheckOut(c.Request.Context(), userID, imageJPEG, deviceKey, userAgent, c.ClientIP(), lat, lng)
		}

		// Set before checking err -- see the comment in challengeHandler. A device
		// can be resolved and committed even when the attendance write that
		// follows then fails on a business rule (already marked, no check-in
		// yet, ...); the client must still learn its key or it silently loses
		// that device slot on every subsequent request.
		if issuedKey != "" {
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(deviceCookieName, issuedKey, deviceCookieMaxAge, "/api/v1", "", true, true)
		}

		if err != nil {
			body := gin.H{"error": messageFor(err)}
			var wrongPerson *ErrWrongPerson
			if errors.As(err, &wrongPerson) {
				body["matched_name"] = wrongPerson.Name
			}
			c.JSON(statusFor(err), body)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func statusFor(err error) int {
	var wrongPerson *ErrWrongPerson
	var outsideRadius *ErrOutsideRadius
	switch {
	case errors.As(err, &wrongPerson):
		return http.StatusUnprocessableEntity
	case errors.As(err, &outsideRadius):
		return http.StatusForbidden
	case errors.Is(err, ErrNoFaceMatch), errors.Is(err, ErrLivenessFailed),
		errors.Is(err, ErrStillImage), errors.Is(err, ErrChallengeFailed):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrChallengeRequired):
		return http.StatusPreconditionRequired
	case errors.Is(err, ErrTooFewFrames):
		return http.StatusBadRequest
	case errors.Is(err, ErrNoCheckInYet):
		return http.StatusConflict
	case errors.Is(err, ErrOutsideOfficeNet), errors.Is(err, ErrDeviceNotApproved), errors.Is(err, ErrLocationRequired):
		return http.StatusForbidden
	case errors.Is(err, ErrNoEmployeeRecord):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func messageFor(err error) string {
	var wrongPerson *ErrWrongPerson
	var outsideRadius *ErrOutsideRadius
	switch {
	case errors.As(err, &wrongPerson):
		return "Wajah cocok dengan " + wrongPerson.Name + ", bukan akun yang sedang login."
	case errors.As(err, &outsideRadius):
		return outsideRadius.Error()
	case errors.Is(err, ErrLocationRequired):
		return "Izinkan akses lokasi di browser untuk absen di lokasi kerja ini."
	case errors.Is(err, ErrNoFaceMatch):
		return "Wajah tidak dikenali. Coba lagi atau ajukan koreksi manual."
	case errors.Is(err, ErrLivenessFailed):
		return "Gunakan wajah asli, bukan foto atau layar."
	case errors.Is(err, ErrChallengeRequired):
		return "Perlu verifikasi gerakan. Hadap kamera dan gerakkan kepala sedikit."
	case errors.Is(err, ErrStillImage):
		return "Tidak terdeteksi gerakan. Pastikan Anda menghadap kamera langsung."
	case errors.Is(err, ErrChallengeFailed):
		return "Verifikasi gerakan gagal. Coba di tempat yang lebih terang."
	case errors.Is(err, ErrTooFewFrames):
		return "Verifikasi gerakan butuh beberapa frame."
	case errors.Is(err, ErrNoCheckInYet):
		return "Belum ada absen masuk hari ini."
	case errors.Is(err, ErrNoEmployeeRecord):
		return "Akun ini tidak terhubung ke data karyawan."
	case errors.Is(err, ErrOutsideOfficeNet):
		return "Absen hanya bisa dilakukan dari jaringan kantor, kecuali diizinkan remote."
	case errors.Is(err, ErrDeviceNotApproved):
		return "Perangkat ini belum disetujui HR. Hubungi HR untuk mendaftarkannya."
	default:
		return "Terjadi kesalahan saat memverifikasi wajah."
	}
}
