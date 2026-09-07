// Package report powers the HR monitoring screen and CSV export.
// docs/PRD.md F6: the whole point is that closing the month stops being manual.
package report

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAttendanceNotFound = errors.New("data absensi tidak ditemukan")

type Service struct {
	pool     *pgxpool.Pool
	timezone *time.Location
}

func NewService(pool *pgxpool.Pool, tz *time.Location) *Service {
	if tz == nil {
		tz = time.UTC
	}
	return &Service{pool: pool, timezone: tz}
}

type Filter struct {
	From         *time.Time
	To           *time.Time
	DepartmentID string
	EmployeeID   string
	Status       string
	// ManagerID limits results to that manager's direct reports; empty means no limit.
	ManagerID string
	Limit     int
	Offset    int
}

type Row struct {
	ID            string    `json:"id"`
	EmployeeID    string    `json:"employee_id"`
	NIK           string    `json:"nik"`
	FullName      string    `json:"full_name"`
	Department    string    `json:"department"`
	Type          string    `json:"type"`
	OccurredAt    time.Time `json:"occurred_at"`
	Status        string    `json:"status"`
	Source        string    `json:"source"`
	LowConfidence bool      `json:"low_confidence"`
}

// buildWhere keeps the filter logic in one place so the list, the count and the
// export can never drift apart -- an export that ignores the active filter is
// the classic way these screens lie to people.
func (f Filter) buildWhere() (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}

	add := func(clause string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}

	if f.From != nil {
		add("a.occurred_at >= $%d", *f.From)
	}
	if f.To != nil {
		add("a.occurred_at < $%d", *f.To)
	}
	if f.DepartmentID != "" {
		add("e.department_id = $%d", f.DepartmentID)
	}
	if f.EmployeeID != "" {
		add("e.id = $%d", f.EmployeeID)
	}
	if f.Status != "" {
		add("a.status = $%d", f.Status)
	}
	if f.ManagerID != "" {
		add("e.manager_id = $%d", f.ManagerID)
	}
	return strings.Join(clauses, " AND "), args
}

const rowSelect = `
	SELECT a.id, e.id, e.nik, e.full_name, COALESCE(d.name, ''),
	       a.type, a.occurred_at, a.status, a.source, a.low_confidence
	FROM attendances a
	JOIN employees e ON e.id = a.employee_id
	LEFT JOIN departments d ON d.id = e.department_id
	WHERE `

func (s *Service) List(ctx context.Context, f Filter) ([]Row, int, error) {
	where, args := f.buildWhere()

	var total int
	countSQL := `SELECT count(*) FROM attendances a
		JOIN employees e ON e.id = a.employee_id
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE ` + where
	if err := s.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	sql := rowSelect + where + fmt.Sprintf(" ORDER BY a.occurred_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, f.Offset)

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []Row{}
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.NIK, &r.FullName, &r.Department,
			&r.Type, &r.OccurredAt, &r.Status, &r.Source, &r.LowConfidence); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// All returns every matching row without paging, for the CSV export.
func (s *Service) All(ctx context.Context, f Filter) ([]Row, error) {
	where, args := f.buildWhere()
	rows, err := s.pool.Query(ctx, rowSelect+where+" ORDER BY a.occurred_at DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Row{}
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.NIK, &r.FullName, &r.Department,
			&r.Type, &r.OccurredAt, &r.Status, &r.Source, &r.LowConfidence); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Delete removes an incorrect attendance record at HR's direction. The audit
// row keeps the original event and its actor available for later investigation.
func (s *Service) Delete(ctx context.Context, attendanceID, actorID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var employeeID, kind, status string
	var occurredAt time.Time
	err = tx.QueryRow(ctx, `
		DELETE FROM attendances
		WHERE id = $1
		RETURNING employee_id, type, status, occurred_at`, attendanceID).
		Scan(&employeeID, &kind, &status, &occurredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAttendanceNotFound
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, entity, entity_id, before_data)
		VALUES ($1, 'attendance_deleted', 'attendance', $2,
			jsonb_build_object('employee_id', $3::text, 'type', $4::text, 'status', $5::text, 'occurred_at', $6::timestamptz))`,
		actorID, attendanceID, employeeID, kind, status, occurredAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Update lets HR correct the date/time of an existing record directly from
// the monitoring table (docs/PRD.md 10.4), as an alternative to the
// employee-initiated correction request when HR already knows the fix.
// Status is re-derived from the employee's schedule against the NEW time --
// letting HR pick status directly would drift from what the schedule math
// says and desync the two. The row is tagged manual_correction: once HR has
// touched the time, it is no longer purely what the face model produced.
func (s *Service) Update(ctx context.Context, attendanceID, actorID string, newOccurredAt time.Time) (*Row, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var employeeID, kind, oldStatus string
	var oldOccurredAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT employee_id, type, status, occurred_at FROM attendances WHERE id = $1 FOR UPDATE`,
		attendanceID).Scan(&employeeID, &kind, &oldStatus, &oldOccurredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttendanceNotFound
	}
	if err != nil {
		return nil, err
	}

	newStatus, err := s.deriveStatus(ctx, tx, employeeID, kind, newOccurredAt)
	if err != nil {
		return nil, err
	}

	var row Row
	err = tx.QueryRow(ctx, `
		UPDATE attendances
		SET occurred_at = $1, status = $2, source = 'manual_correction'
		WHERE id = $3
		RETURNING id, employee_id,
		          (SELECT nik FROM employees WHERE id = employee_id),
		          (SELECT full_name FROM employees WHERE id = employee_id),
		          COALESCE((SELECT d.name FROM departments d JOIN employees e ON e.department_id = d.id WHERE e.id = employee_id), ''),
		          type, occurred_at, status, source, low_confidence`,
		newOccurredAt, newStatus, attendanceID).
		Scan(&row.ID, &row.EmployeeID, &row.NIK, &row.FullName, &row.Department,
			&row.Type, &row.OccurredAt, &row.Status, &row.Source, &row.LowConfidence)
	if err != nil {
		return nil, err
	}

	before, _ := json.Marshal(map[string]string{"occurred_at": oldOccurredAt.Format(time.RFC3339), "status": oldStatus})
	after, _ := json.Marshal(map[string]string{"occurred_at": newOccurredAt.Format(time.RFC3339), "status": newStatus})
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, entity, entity_id, before_data, after_data)
		VALUES ($1, 'attendance_updated', 'attendance', $2, $3, $4)`,
		actorID, attendanceID, before, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &row, nil
}

// deriveStatus mirrors internal/attendance's schedule-derived status math.
// Duplicated rather than imported: report (HR editing history after the fact)
// and attendance (live face-verification) are genuinely separate bounded
// contexts, and this is a ~15-line query, not a shared library worth coupling
// the two packages over.
func (s *Service) deriveStatus(ctx context.Context, tx pgx.Tx, employeeID, kind string, at time.Time) (string, error) {
	var start, end time.Time
	var tolerance int
	err := tx.QueryRow(ctx, `
		SELECT ws.start_time, ws.end_time, ws.late_tolerance_minutes
		FROM employees e
		JOIN work_schedules ws ON ws.id = e.schedule_id
		WHERE e.id = $1`, employeeID).Scan(&start, &end, &tolerance)
	if errors.Is(err, pgx.ErrNoRows) {
		return "on_time", nil
	}
	if err != nil {
		return "", err
	}

	local := at.In(s.timezone)
	minutesNow := local.Hour()*60 + local.Minute()

	switch kind {
	case "check_in":
		if minutesNow > start.Hour()*60+start.Minute()+tolerance {
			return "late", nil
		}
	case "check_out":
		if minutesNow < end.Hour()*60+end.Minute() {
			return "early_leave", nil
		}
	}
	return "on_time", nil
}

// DailySummary is the number HR actually acts on at 09:00: who is still missing.
type DailySummary struct {
	Date        string `json:"date"`
	TotalActive int    `json:"total_active"`
	OnTime      int    `json:"on_time"`
	Late        int    `json:"late"`
	NotYet      int    `json:"not_yet"`
	CheckedOut  int    `json:"checked_out"`
}

func (s *Service) Today(ctx context.Context, managerID string) (*DailySummary, error) {
	scope := ""
	args := []any{}
	if managerID != "" {
		scope = " AND e.manager_id = $1"
		args = append(args, managerID)
	}

	var sum DailySummary
	err := s.pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE e.status = 'active'),
			count(*) FILTER (WHERE ci.status = 'on_time'),
			count(*) FILTER (WHERE ci.status = 'late'),
			count(*) FILTER (WHERE e.status = 'active' AND ci.id IS NULL),
			count(*) FILTER (WHERE co.id IS NOT NULL)
		FROM employees e
		LEFT JOIN LATERAL (
			SELECT a.id, a.status FROM attendances a
			WHERE a.employee_id = e.id AND a.type = 'check_in'
			  AND a.occurred_at::date = CURRENT_DATE
			LIMIT 1
		) ci ON true
		LEFT JOIN LATERAL (
			SELECT a.id FROM attendances a
			WHERE a.employee_id = e.id AND a.type = 'check_out'
			  AND a.occurred_at::date = CURRENT_DATE
			LIMIT 1
		) co ON true
		WHERE e.status != 'inactive'`+scope, args...).
		Scan(&sum.TotalActive, &sum.OnTime, &sum.Late, &sum.NotYet, &sum.CheckedOut)
	if err != nil {
		return nil, err
	}
	sum.Date = time.Now().In(s.timezone).Format("2006-01-02")
	return &sum, nil
}

// MyHistory is an employee's own month, paired per day so the row reads the way
// a payslip does rather than as a raw event log.
type DayRecord struct {
	Date       string     `json:"date"`
	CheckInAt  *time.Time `json:"check_in_at"`
	CheckOutAt *time.Time `json:"check_out_at"`
	Status     string     `json:"status"`
	Minutes    int        `json:"minutes"`
}

func (s *Service) MyHistory(ctx context.Context, userID string, from, to time.Time) ([]DayRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d::date::text,
		       ci.occurred_at, co.occurred_at, COALESCE(ci.status, 'absent')
		FROM generate_series($2::date, $3::date - interval '1 day', interval '1 day') d
		LEFT JOIN employees e ON e.user_id = $1
		LEFT JOIN LATERAL (
			SELECT a.occurred_at, a.status FROM attendances a
			WHERE a.employee_id = e.id AND a.type = 'check_in' AND a.occurred_at::date = d::date
			ORDER BY a.occurred_at LIMIT 1
		) ci ON true
		LEFT JOIN LATERAL (
			SELECT a.occurred_at FROM attendances a
			WHERE a.employee_id = e.id AND a.type = 'check_out' AND a.occurred_at::date = d::date
			ORDER BY a.occurred_at DESC LIMIT 1
		) co ON true
		ORDER BY d DESC`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []DayRecord{}
	for rows.Next() {
		var r DayRecord
		if err := rows.Scan(&r.Date, &r.CheckInAt, &r.CheckOutAt, &r.Status); err != nil {
			return nil, err
		}
		if r.CheckInAt != nil && r.CheckOutAt != nil {
			r.Minutes = int(r.CheckOutAt.Sub(*r.CheckInAt).Minutes())
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) employeeIDForUser(ctx context.Context, userID string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id FROM employees WHERE user_id = $1`, userID).Scan(&id)
	return id, err
}
