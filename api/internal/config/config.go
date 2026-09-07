// Package config loads runtime settings from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hris-face/api/internal/notify"
)

type Config struct {
	Env               string
	ListenAddr        string
	DatabaseURL       string
	JWTSecret         string
	FaceServiceURL    string
	OfficeIPAllowlist []string
	// Local time zone for comparing check-in time against the work schedule.
	// occurred_at is stored as timestamptz (UTC); without this, "late" would be
	// judged in UTC and be 7 hours off for a WIB office.
	Timezone *time.Location
	SMTP     notify.Config
	// Empty PHOTO_DIR disables evidence photos entirely (docs/PRD.md 8).
	PhotoDir           string
	PhotoRetentionDays int
}

func Load() Config {
	return Config{
		Env:               getenv("APP_ENV", "development"),
		ListenAddr:        getenv("LISTEN_ADDR", "127.0.0.1:8080"),
		DatabaseURL:       mustEnv("DATABASE_URL"),
		JWTSecret:         mustEnv("JWT_SECRET"),
		FaceServiceURL:    getenv("FACE_SERVICE_URL", "http://127.0.0.1:8000"),
		OfficeIPAllowlist: splitCSV(os.Getenv("OFFICE_IP_ALLOWLIST")),
		Timezone:          mustLocation(getenv("TIMEZONE", "Asia/Jakarta")),
		SMTP: notify.Config{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     getenv("SMTP_PORT", "587"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
		PhotoDir:           os.Getenv("PHOTO_DIR"),
		PhotoRetentionDays: atoiDefault(os.Getenv("PHOTO_RETENTION_DAYS"), 30),
	}
}

func atoiDefault(v string, fallback int) int {
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func mustLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("TIMEZONE tidak valid: " + name)
	}
	return loc
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env var: " + key)
	}
	return v
}

func splitCSV(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
