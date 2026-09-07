package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hris-face/api/internal/attendance"
	"github.com/hris-face/api/internal/auth"
	"github.com/hris-face/api/internal/config"
	"github.com/hris-face/api/internal/correction"
	"github.com/hris-face/api/internal/db"
	"github.com/hris-face/api/internal/employee"
	"github.com/hris-face/api/internal/enrollment"
	"github.com/hris-face/api/internal/faceclient"
	"github.com/hris-face/api/internal/masterdata"
	"github.com/hris-face/api/internal/middleware"
	"github.com/hris-face/api/internal/notify"
	"github.com/hris-face/api/internal/report"
)

func main() {
	cfg := config.Load()

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pool, err := db.Connect(cfg.DatabaseURL, cfg.Timezone.String())
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	face := faceclient.New(cfg.FaceServiceURL)
	authSvc := auth.NewService(pool, cfg.JWTSecret)
	empSvc := employee.NewService(pool)
	photoStore := attendance.NewPhotoStore(cfg.PhotoDir, cfg.PhotoRetentionDays)
	photoStore.StartCleanup()
	attSvc := attendance.NewService(pool, face, cfg.OfficeIPAllowlist, cfg.Timezone, photoStore)
	enrollSvc := enrollment.NewService(pool, face)
	mailer := notify.New(cfg.SMTP)
	reportSvc := report.NewService(pool, cfg.Timezone)
	correctionSvc := correction.NewService(pool, mailer)
	masterSvc := masterdata.NewService(pool)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logger())

	r.GET("/healthz", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// The IP limiter only stops crude floods. It must stay generous because the
	// whole office shares one NAT address -- per-account brute-force protection
	// lives in the auth handler, keyed by email.
	authGroup := v1.Group("")
	authGroup.Use(middleware.RateLimit(120, time.Minute))
	auth.RegisterRoutes(authGroup, authSvc)

	protected := v1.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret))
	employee.RegisterRoutes(protected, empSvc)
	report.RegisterRoutes(protected, reportSvc)
	correction.RegisterRoutes(protected, correctionSvc)
	masterdata.RegisterRoutes(protected, masterSvc)
	enrollment.RegisterRoutes(protected, enrollSvc)

	attendanceGroup := v1.Group("")
	attendanceGroup.Use(middleware.RequireAuth(cfg.JWTSecret), middleware.RateLimit(120, time.Minute))
	attendance.RegisterRoutes(attendanceGroup, attSvc)
	attendance.RegisterScanRoute(protected, attSvc)

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
