// Command seed creates the first administrator account.
// Without it the system cannot be bootstrapped: every employee-creation
// endpoint requires an authenticated hr/superadmin, and nothing creates
// that first user. Run once after the first deploy:
//
//	SEED_EMAIL=admin@perusahaan.com SEED_PASSWORD=... ./seed
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/hris-face/api/internal/db"
)

func main() {
	databaseURL := mustEnv("DATABASE_URL")
	email := mustEnv("SEED_EMAIL")
	password := mustEnv("SEED_PASSWORD")

	if len(password) < 12 {
		log.Fatal("SEED_PASSWORD minimal 12 karakter")
	}

	// Seed itself has no date-boundary logic, but db.Connect's session timezone
	// applies pool-wide, so pass the same TIMEZONE the API uses for consistency.
	timezone := os.Getenv("TIMEZONE")
	if timezone == "" {
		timezone = "Asia/Jakarta"
	}
	pool, err := db.Connect(databaseURL, timezone)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing string
	err = pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&existing)
	if err == nil {
		log.Fatalf("user %s sudah ada", email)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Fatalf("cek user: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	var id string
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'superadmin') RETURNING id`, email, string(hash)).Scan(&id)
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}

	log.Printf("superadmin dibuat: %s (%s)", email, id)

	// A system with no work schedule silently loses late/early_leave: every
	// check-in reads as on_time. Bootstrapping has to guarantee one exists.
	var scheduleCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM work_schedules`).Scan(&scheduleCount); err != nil {
		log.Fatalf("cek jadwal: %v", err)
	}
	if scheduleCount == 0 {
		if _, err := pool.Exec(ctx, `
			INSERT INTO work_schedules (name, start_time, end_time, late_tolerance_minutes, work_days)
			VALUES ('Reguler 08:00-17:00', '08:00', '17:00', 15, '{1,2,3,4,5}')`); err != nil {
			log.Fatalf("buat jadwal default: %v", err)
		}
		log.Print("jadwal default dibuat: Reguler 08:00-17:00")
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("env %s wajib diisi", key)
	}
	return v
}
