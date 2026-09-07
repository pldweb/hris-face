// Package db wraps the pgx connection pool.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens the pool and pins every connection's session timezone to tz.
//
// Without this, Postgres defaults each session to the SERVER's timezone
// (typically UTC), so bare `CURRENT_DATE` and `occurred_at::date` comparisons
// answer "what day is it in UTC" while the rest of the app (attendance status,
// the daily summary's own date label) reasons in tz. The two disagree for
// exactly the hours between UTC midnight and local midnight -- 00:00-07:00 WIB
// for Asia/Jakarta -- during which "already checked in today" can flip on its
// own, letting an early check-in be duplicated a few hours later. Setting the
// session timezone once here fixes every current and future query in the app
// that leans on CURRENT_DATE/::date, rather than patching each call site.
func Connect(url string, tz string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	if tz != "" {
		poolConfig.ConnConfig.RuntimeParams["timezone"] = tz
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pgxpool.NewWithConfig(ctx, poolConfig)
}
