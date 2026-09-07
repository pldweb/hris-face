// Package leave handles employee leave/permission/sick requests (cuti/izin/sakit).
package leave

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
	ErrNoEmployeeRecord = errors.New("akun ini tidak terhubung ke data karyawan")
	ErrInvalidType      = errors.New("jenis cuti tidak valid")
	ErrInvalidRange     = errors.New("rentang tanggal tidak valid")
	ErrNoWorkdays       = errors.New("rentang tanggal tidak memiliki hari kerja")
	ErrOverlap          = errors.New("rentang tanggal bertabrakan dengan pengajuan cuti lain")
	ErrQuotaExceeded    = errors.New("sisa kuota cuti tahunan tidak cukup")
	ErrNotFound         = errors.New("pengajuan cuti tidak ditemukan")
	ErrAlreadyReviewed  = errors.New("pengajuan cuti sudah diproses")
)

var defaultWorkDays = []int16{1, 2, 3, 4, 5}

const dateLayout = "2006-01-02"

type Service struct {
	pool   *pgxpool.Pool
	mailer *notify.Sender
}

func NewService(pool *pgxpool.Pool, mailer *notify.Sender) *Service {
	return &Service{pool: pool, mailer: mailer}
}

type LeaveRequest struct {
	ID         string     `json:"id"`
	EmployeeID string     `json:"employee_id"`
	FullName   string     `json:"full_name,omitempty"`
	Type       string     `json:"type"`
	StartDate  string     `json:"start_date"`
	EndDate    string     `json:"end_date"`
	DaysCount  int        `json:"days_count"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Balance struct {
	Year      int `json:"year"`
	Quota     int `json:"quota"`
	Used      int `json:"used"`
	Remaining int `json:"remaining"`
}

func (s *Service) Create(ctx context.Context, userID, leaveType, startDate, endDate, reason string) (*LeaveRequest, error) {
	if leaveType != "annual" && leaveType != "sick" && leaveType != "permit" {
		return nil, ErrInvalidType
	}

	start, err := time.Parse(dateLayout, startDate)
	if err != nil {
		return nil, ErrInvalidRange
	}
	end, err := time.Parse(dateLayout, endDate)
	if err != nil {
		return nil, ErrInvalidRange
	}
	if end.Before(start) {
		return nil, ErrInvalidRange
	}

	var employeeID, fullName string
	var scheduleID *string
	err = s.pool.QueryRow(ctx,
		`SELECT id, full_name, schedule_id::text FROM employees WHERE user_id = $1`, userID).
		Scan(&employeeID, &fullName, &scheduleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEmployeeRecord
	}
	if err != nil {
		return nil, err
	}

	workDays := defaultWorkDays
	if scheduleID != nil {
		if err := s.pool.QueryRow(ctx,
			`SELECT work_days FROM work_schedules WHERE id = $1`, *scheduleID).Scan(&workDays); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	daysCount := countWorkdays(start, end, workDays)
	if daysCount == 0 {
		return nil, ErrNoWorkdays
	}

	var overlap bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM leave_requests
			WHERE employee_id = $1 AND status IN ('pending','approved')
			  AND daterange(start_date, end_date, '[]') && daterange($2::date, $3::date, '[]')
		)`, employeeID, startDate, endDate).Scan(&overlap); err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
	}

	if leaveType == "annual" {
		bal, err := s.Balance(ctx, employeeID, start.Year())
		if err != nil {
			return nil, err
		}
		if bal.Remaining < daysCount {
			return nil, ErrQuotaExceeded
		}
	}

	lr := LeaveRequest{
		EmployeeID: employeeID, FullName: fullName, Type: leaveType,
		StartDate: startDate, EndDate: endDate, DaysCount: daysCount,
		Reason: reason, Status: "pending",
	}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO leave_requests (employee_id, type, start_date, end_date, days_count, reason)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`,
		employeeID, leaveType, startDate, endDate, daysCount, reason).Scan(&lr.ID, &lr.CreatedAt)
	if err != nil {
		return nil, err
	}

	go s.notifyHR(lr) //nolint:errcheck
	return &lr, nil
}

// countWorkdays counts calendar dates in [start, end] whose ISO weekday
// (Monday=1..Sunday=7) appears in workDays.
func countWorkdays(start, end time.Time, workDays []int16) int {
	allowed := make(map[int]struct{}, len(workDays))
	for _, d := range workDays {
		allowed[int(d)] = struct{}{}
	}
	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if _, ok := allowed[int(d.Weekday()+6)%7 + 1]; ok {
			count++
		}
	}
	return count
}

func (s *Service) employeeIDForUser(ctx context.Context, userID string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id FROM employees WHERE user_id = $1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoEmployeeRecord
	}
	return id, err
}

const leaveSelect = `
	SELECT lr.id, lr.employee_id, e.full_name, lr.type, lr.start_date::text, lr.end_date::text,
	       lr.days_count, lr.reason, lr.status, lr.reviewed_at, lr.created_at
	FROM leave_requests lr
	JOIN employees e ON e.id = lr.employee_id
	WHERE `

func scanLeaveRows(rows pgx.Rows) ([]LeaveRequest, error) {
	out := []LeaveRequest{}
	for rows.Next() {
		var lr LeaveRequest
		if err := rows.Scan(&lr.ID, &lr.EmployeeID, &lr.FullName, &lr.Type, &lr.StartDate, &lr.EndDate,
			&lr.DaysCount, &lr.Reason, &lr.Status, &lr.ReviewedAt, &lr.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, lr)
	}
	return out, rows.Err()
}

func (s *Service) ListMine(ctx context.Context, userID, status string) ([]LeaveRequest, error) {
	employeeID, err := s.employeeIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, leaveSelect+`lr.employee_id = $1 AND ($2 = '' OR lr.status = $2) ORDER BY lr.start_date DESC`,
		employeeID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaveRows(rows)
}

func (s *Service) ListAdmin(ctx context.Context, status, leaveType, employeeID, from, to string) ([]LeaveRequest, error) {
	where := "1=1"
	args := []any{}
	add := func(clause string, value any) {
		args = append(args, value)
		where += " AND " + clause + " $" + strconv.Itoa(len(args))
	}
	if status != "" {
		add("lr.status =", status)
	}
	if leaveType != "" {
		add("lr.type =", leaveType)
	}
	if employeeID != "" {
		add("lr.employee_id =", employeeID)
	}
	if from != "" {
		add("lr.start_date >=", from)
	}
	if to != "" {
		add("lr.end_date <=", to)
	}

	rows, err := s.pool.Query(ctx, leaveSelect+where+` ORDER BY lr.created_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaveRows(rows)
}

func (s *Service) Balance(ctx context.Context, employeeID string, year int) (*Balance, error) {
	var quota int
	if err := s.pool.QueryRow(ctx, `SELECT annual_leave_quota FROM employees WHERE id = $1`, employeeID).Scan(&quota); err != nil {
		return nil, err
	}
	var used int
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(days_count), 0) FROM leave_requests
		WHERE employee_id = $1 AND type = 'annual' AND status = 'approved' AND EXTRACT(YEAR FROM start_date) = $2`,
		employeeID, year).Scan(&used); err != nil {
		return nil, err
	}
	return &Balance{Year: year, Quota: quota, Used: used, Remaining: quota - used}, nil
}

func (s *Service) Review(ctx context.Context, leaveID, reviewerUserID, decision, note string) (*LeaveRequest, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var employeeID, leaveType, status string
	var startDate time.Time
	var daysCount int
	err = tx.QueryRow(ctx, `
		SELECT employee_id, type, start_date, days_count, status
		FROM leave_requests WHERE id = $1 FOR UPDATE`, leaveID).
		Scan(&employeeID, &leaveType, &startDate, &daysCount, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status != "pending" {
		return nil, ErrAlreadyReviewed
	}

	newStatus := "rejected"
	if decision == "approve" {
		newStatus = "approved"
	}

	if newStatus == "approved" && leaveType == "annual" {
		var quota, used int
		if err := tx.QueryRow(ctx, `SELECT annual_leave_quota FROM employees WHERE id = $1`, employeeID).Scan(&quota); err != nil {
			return nil, err
		}
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(days_count), 0) FROM leave_requests
			WHERE employee_id = $1 AND type = 'annual' AND status = 'approved' AND EXTRACT(YEAR FROM start_date) = $2`,
			employeeID, startDate.Year()).Scan(&used); err != nil {
			return nil, err
		}
		if quota-used < daysCount {
			return nil, ErrQuotaExceeded
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE leave_requests SET status = $1, reviewed_by = $2, reviewed_at = now() WHERE id = $3`,
		newStatus, reviewerUserID, leaveID); err != nil {
		return nil, err
	}

	after, _ := json.Marshal(map[string]string{"status": newStatus, "note": note})
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, entity, entity_id, after_data)
		VALUES ($1, $2, 'leave_request', $3, $4)`,
		reviewerUserID, "leave_"+newStatus, leaveID, after); err != nil {
		return nil, err
	}

	var lr LeaveRequest
	err = tx.QueryRow(ctx, `
		SELECT lr.id, lr.employee_id, e.full_name, lr.type, lr.start_date::text, lr.end_date::text,
		       lr.days_count, lr.reason, lr.status, lr.reviewed_at, lr.created_at
		FROM leave_requests lr JOIN employees e ON e.id = lr.employee_id WHERE lr.id = $1`, leaveID).
		Scan(&lr.ID, &lr.EmployeeID, &lr.FullName, &lr.Type, &lr.StartDate, &lr.EndDate,
			&lr.DaysCount, &lr.Reason, &lr.Status, &lr.ReviewedAt, &lr.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	go s.notifyEmployee(employeeID, newStatus, note) //nolint:errcheck
	return &lr, nil
}

func (s *Service) notifyHR(lr LeaveRequest) {
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
	s.mailer.Send(to, "Pengajuan cuti menunggu persetujuan", //nolint:errcheck
		lr.FullName+" mengajukan "+lr.Type+" dari "+lr.StartDate+" sampai "+lr.EndDate+
			" ("+strconv.Itoa(lr.DaysCount)+" hari kerja).\n\nAlasan: "+lr.Reason)
}

func (s *Service) notifyEmployee(employeeID, status, note string) {
	ctx := context.Background()
	var email string
	if err := s.pool.QueryRow(ctx, `
		SELECT u.email FROM users u JOIN employees e ON e.user_id = u.id WHERE e.id = $1`,
		employeeID).Scan(&email); err != nil {
		return
	}
	body := "Pengajuan cuti Anda telah " + status + "."
	if note != "" {
		body += "\n\nCatatan HR: " + note
	}
	s.mailer.Send([]string{email}, "Status pengajuan cuti Anda", body) //nolint:errcheck
}
