// Package masterdata covers work schedules and locations (docs/PRD.md F2).
// Without schedule editing, the late/early_leave rules are frozen at whatever
// the initial migration seeded.
package masterdata

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("data tidak ditemukan")
var ErrLocationInUse = errors.New("lokasi masih dipakai oleh karyawan aktif")

type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

type Schedule struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	StartTime            string  `json:"start_time"`
	EndTime              string  `json:"end_time"`
	LateToleranceMinutes int     `json:"late_tolerance_minutes"`
	WorkDays             []int16 `json:"work_days"`
}

func (s *Service) ListSchedules(ctx context.Context) ([]Schedule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
		       late_tolerance_minutes, work_days
		FROM work_schedules ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Schedule{}
	for rows.Next() {
		var sc Schedule
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.StartTime, &sc.EndTime,
			&sc.LateToleranceMinutes, &sc.WorkDays); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *Service) CreateSchedule(ctx context.Context, in Schedule) (*Schedule, error) {
	if len(in.WorkDays) == 0 {
		in.WorkDays = []int16{1, 2, 3, 4, 5}
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO work_schedules (name, start_time, end_time, late_tolerance_minutes, work_days)
		VALUES ($1, $2::time, $3::time, $4, $5) RETURNING id`,
		in.Name, in.StartTime, in.EndTime, in.LateToleranceMinutes, in.WorkDays).Scan(&in.ID)
	if err != nil {
		return nil, err
	}
	return &in, nil
}

func (s *Service) UpdateSchedule(ctx context.Context, id string, in Schedule) error {
	// A typed nil, not an untyped one: pgx cannot infer the array type of a bare
	// nil, and the ::smallint[] cast in the SQL needs something to cast.
	var days []int16
	if len(in.WorkDays) > 0 {
		days = in.WorkDays
	}
	// COALESCE, not a Mon-Fri default: an update that omits work_days means
	// "leave the working days alone", never "reset them".
	tag, err := s.pool.Exec(ctx, `
		UPDATE work_schedules
		SET name = $1, start_time = $2::time, end_time = $3::time,
		    late_tolerance_minutes = $4, work_days = COALESCE($5::smallint[], work_days)
		WHERE id = $6`,
		in.Name, in.StartTime, in.EndTime, in.LateToleranceMinutes, days, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type Location struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Lat          *float64 `json:"lat,omitempty"`
	Lng          *float64 `json:"lng,omitempty"`
	RadiusMeters *int     `json:"radius_meters,omitempty"`
}

func (s *Service) ListLocations(ctx context.Context) ([]Location, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, lat, lng, radius_meters FROM work_locations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Location{}
	for rows.Next() {
		var l Location
		if err := rows.Scan(&l.ID, &l.Name, &l.Lat, &l.Lng, &l.RadiusMeters); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Service) CreateLocation(ctx context.Context, name string, lat, lng *float64, radiusMeters *int) (*Location, error) {
	l := Location{Name: name, Lat: lat, Lng: lng, RadiusMeters: radiusMeters}
	if err := s.pool.QueryRow(ctx,
		`INSERT INTO work_locations (name, lat, lng, radius_meters) VALUES ($1, $2, $3, $4) RETURNING id`,
		name, lat, lng, radiusMeters).Scan(&l.ID); err != nil {
		return nil, err
	}
	return &l, nil
}

// UpdateLocation only touches lat/lng/radius_meters when setGeofence is true.
// A plain rename (InlineMasterSelect's quick-add-from-employee-form, which has
// no geofence UI at all) must never wipe a radius HR already configured
// through the dedicated location editor -- a bare *float64 can't tell "not
// sent" from "sent as null", so the caller has to say which it means.
func (s *Service) UpdateLocation(ctx context.Context, id, name string, setGeofence bool, lat, lng *float64, radiusMeters *int) error {
	var tag pgconn.CommandTag
	var err error
	if setGeofence {
		tag, err = s.pool.Exec(ctx,
			`UPDATE work_locations SET name = $1, lat = $2, lng = $3, radius_meters = $4 WHERE id = $5`,
			name, lat, lng, radiusMeters, id)
	} else {
		tag, err = s.pool.Exec(ctx, `UPDATE work_locations SET name = $1 WHERE id = $2`, name, id)
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteLocation refuses (with a readable error, not a raw FK-violation) when
// an employee still points at it -- work_locations has no ON DELETE clause,
// so Postgres would otherwise reject this with an opaque 23503.
func (s *Service) DeleteLocation(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM work_locations WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrLocationInUse
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// AssignSchedule lets HR move one employee onto a different shift.
func (s *Service) AssignSchedule(ctx context.Context, employeeID, scheduleID string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE employees SET schedule_id = $1 WHERE id = $2`, scheduleID, employeeID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
