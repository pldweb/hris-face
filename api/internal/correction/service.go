// Package correction handles manual attendance fixes (docs/PRD.md F7).
// This is the escape hatch for the FRR the face model cannot avoid: without it,
// an employee the camera refuses simply loses the day.
package correction

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hris-face/api/internal/notify"
)

var (
	ErrNotFound         = errors.New("pengajuan tidak ditemukan")
	ErrAlreadyReviewed  = errors.New("pengajuan sudah diproses")
	ErrNoEmployeeRecord = errors.New("akun ini tidak terhubung ke data karyawan")
	ErrInvalidType      = errors.New("tipe harus check_in atau check_out")
)

type Service struct {
	pool   *pgxpool.Pool
	mailer *notify.Sender
}

func NewService(pool *pgxpool.Pool, mailer *notify.Sender) *Service {
	return &Service{pool: pool, mailer: mailer}
}

type Correction struct {
	ID            string     `json:"id"`
	EmployeeID    string     `json:"employee_id"`
	FullName      string     `json:"full_name"`
	RequestedType string     `json:"requested_type"`
	RequestedTime time.Time  `json:"requested_time"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (s *Service) Create(ctx context.Context, userID, reqType string, reqTime time.Time, reason string) (*Correction, error) {
	if reqType != "check_in" && reqType != "check_out" {
		return nil, ErrInvalidType
	}

	var employeeID, fullName string
	err := s.pool.QueryRow(ctx,
		`SELECT id, full_name FROM employees WHERE user_id = $1`, userID).Scan(&employeeID, &fullName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEmployeeRecord
	}
	if err != nil {
		return nil, err
	}

	c := Correction{
		EmployeeID: employeeID, FullName: fullName, RequestedType: reqType,
		RequestedTime: reqTime, Reason: reason, Status: "pending",
	}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO attendance_corrections (employee_id, requested_type, requested_time, reason)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		employeeID, reqType, reqTime, reason).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return nil, err
	}

	go s.notifyHR(c) //nolint:errcheck
	return &c, nil
}

func (s *Service) List(ctx context.Context, status, userID string, mineOnly bool) ([]Correction, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		args = append(args, status)
		where += " AND c.status = $1"
	}
	if mineOnly {
		args = append(args, userID)
		where += " AND e.user_id = $" + strconv.Itoa(len(args))
	}

	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.employee_id, e.full_name, c.requested_type, c.requested_time,
		       c.reason, c.status, c.reviewed_at, c.created_at
		FROM attendance_corrections c
		JOIN employees e ON e.id = c.employee_id
		WHERE `+where+` ORDER BY c.created_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Correction{}
	for rows.Next() {
		var c Correction
		if err := rows.Scan(&c.ID, &c.EmployeeID, &c.FullName, &c.RequestedType, &c.RequestedTime,
			&c.Reason, &c.Status, &c.ReviewedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Review approves or rejects. Approving writes the attendance row it stands for,
// tagged source='manual_correction' so a later audit can tell it apart from a
// face-verified record; both paths land in audit_logs.
func (s *Service) Review(ctx context.Context, correctionID, reviewerUserID, decision, note string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var employeeID, reqType, status string
	var reqTime time.Time
	err = tx.QueryRow(ctx, `
		SELECT employee_id, requested_type, requested_time, status
		FROM attendance_corrections WHERE id = $1 FOR UPDATE`, correctionID).
		Scan(&employeeID, &reqType, &reqTime, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "pending" {
		return ErrAlreadyReviewed
	}

	newStatus := "rejected"
	if decision == "approve" {
		newStatus = "approved"
	}

	if _, err := tx.Exec(ctx, `
		UPDATE attendance_corrections
		SET status = $1, reviewed_by = $2, reviewed_at = now()
		WHERE id = $3`, newStatus, reviewerUserID, correctionID); err != nil {
		return err
	}

	if newStatus == "approved" {
		if _, err := tx.Exec(ctx, `
			INSERT INTO attendances (employee_id, type, occurred_at, status, source)
			VALUES ($1, $2, $3, 'on_time', 'manual_correction')`,
			employeeID, reqType, reqTime); err != nil {
			return err
		}
	}

	after, _ := json.Marshal(map[string]string{"status": newStatus, "note": note})
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, entity, entity_id, after_data)
		VALUES ($1, $2, 'attendance_correction', $3, $4)`,
		reviewerUserID, "correction_"+newStatus, correctionID, after); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	go s.notifyEmployee(employeeID, newStatus, note) //nolint:errcheck
	return nil
}

func (s *Service) notifyHR(c Correction) {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT email FROM users WHERE role IN ('hr', 'superadmin')`)
	if err != nil {
		return
	}
	defer rows.Close()

	var to []string
	for rows.Next() {
		var email string
		if rows.Scan(&email) == nil {
			to = append(to, email)
		}
	}
	s.mailer.Send(to, "Koreksi absen menunggu persetujuan", //nolint:errcheck
		c.FullName+" mengajukan koreksi "+c.RequestedType+" untuk "+
			c.RequestedTime.Format("2006-01-02 15:04")+".\n\nAlasan: "+c.Reason)
}

func (s *Service) notifyEmployee(employeeID, status, note string) {
	ctx := context.Background()
	var email string
	if err := s.pool.QueryRow(ctx, `
		SELECT u.email FROM users u JOIN employees e ON e.user_id = u.id WHERE e.id = $1`,
		employeeID).Scan(&email); err != nil {
		return
	}
	body := "Pengajuan koreksi absen Anda telah " + status + "."
	if note != "" {
		body += "\n\nCatatan HR: " + note
	}
	s.mailer.Send([]string{email}, "Status koreksi absen Anda", body) //nolint:errcheck
}
