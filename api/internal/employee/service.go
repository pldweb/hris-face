// Package employee covers master data: employees, departments, locations, schedules.
package employee

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailTaken = errors.New("email sudah terdaftar")
var ErrNIKTaken = errors.New("NIK sudah terdaftar")
var ErrNoEmployeeRecord = errors.New("akun ini tidak terhubung ke data karyawan")
var ErrDeviceNotFound = errors.New("perangkat tidak ditemukan")
var ErrEmployeeNotFound = errors.New("karyawan tidak ditemukan")

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Employee struct {
	ID           string `json:"id"`
	NIK          string `json:"nik"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	DepartmentID string `json:"department_id,omitempty"`
	LocationID   string `json:"location_id,omitempty"`
	Status       string `json:"status"`
}

type CreateEmployeeInput struct {
	NIK          string
	FullName     string
	Email        string
	DepartmentID *string
	LocationID   *string
	ScheduleID   *string
}

// UpdateEmployeeInput edits the employee record and, since the login lives in
// a joined users row, optionally its email too. NIK is intentionally not
// editable here -- it is the durable HR identifier a re-enrollment or a
// device-binding row may already reference by employee_id, not by NIK, but
// treating it as freely renamable invites HR to "fix" it into a collision
// with a real second employee.
type UpdateEmployeeInput struct {
	FullName     string
	Email        string
	DepartmentID *string
	LocationID   *string
}

type CreateEmployeeResult struct {
	Employee     Employee
	TempPassword string
}

func (s *Service) List(ctx context.Context) ([]Employee, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.nik, e.full_name, u.email,
		       COALESCE(e.department_id::text, ''), COALESCE(e.location_id::text, ''), e.status
		FROM employees e
		JOIN users u ON u.id = e.user_id
		ORDER BY e.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Employee
	for rows.Next() {
		var e Employee
		if err := rows.Scan(&e.ID, &e.NIK, &e.FullName, &e.Email, &e.DepartmentID, &e.LocationID, &e.Status); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Create provisions a login (users) and an employee record in one transaction,
// with status pending_enrollment (docs/PRD.md 6.1) and a temp password HR
// hands to the employee out-of-band (chat, printed slip) for their first login.
func (s *Service) Create(ctx context.Context, in CreateEmployeeInput) (*CreateEmployeeResult, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, in.Email).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM employees WHERE nik = $1)`, in.NIK).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrNIKTaken
	}

	tempPassword, err := randomPassword()
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var userID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'employee') RETURNING id`,
		in.Email, string(hash)).Scan(&userID); err != nil {
		return nil, err
	}

	e := Employee{NIK: in.NIK, FullName: in.FullName, Email: in.Email, Status: "pending_enrollment"}
	if in.DepartmentID != nil {
		e.DepartmentID = *in.DepartmentID
	}
	// Fall back to the default schedule when HR did not pick one: an employee
	// with no schedule can never be marked late, which silently turns the whole
	// lateness feature off for everyone created through this endpoint.
	if in.LocationID != nil {
		e.LocationID = *in.LocationID
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO employees (user_id, nik, full_name, department_id, location_id, schedule_id, status)
		VALUES ($1, $2, $3, $4, $5,
			COALESCE($6, (SELECT id FROM work_schedules ORDER BY created_at LIMIT 1)),
			'pending_enrollment')
		RETURNING id`,
		userID, in.NIK, in.FullName, in.DepartmentID, in.LocationID, in.ScheduleID).Scan(&e.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateEmployeeResult{Employee: e, TempPassword: tempPassword}, nil
}

// Update edits name/email/department/location. Email lives on the joined
// users row, so this touches two tables in one transaction.
func (s *Service) Update(ctx context.Context, employeeID string, in UpdateEmployeeInput) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var userID string
	if err := tx.QueryRow(ctx, `SELECT user_id FROM employees WHERE id = $1`, employeeID).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEmployeeNotFound
		}
		return err
	}

	var emailTaken bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`, in.Email, userID).Scan(&emailTaken); err != nil {
		return err
	}
	if emailTaken {
		return ErrEmailTaken
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET email = $1 WHERE id = $2`, in.Email, userID); err != nil {
		return err
	}
	// COALESCE, not a bare overwrite: a caller that omits department_id/location_id
	// (e.g. a partial payload, or a future frontend that doesn't round-trip every
	// field) means "leave it alone", never "clear it". Without this, editing just
	// full_name would silently wipe an employee's department and location --
	// the same destructive-omission bug already fixed once for schedule.work_days.
	if _, err := tx.Exec(ctx, `
		UPDATE employees
		SET full_name = $1,
		    department_id = COALESCE($2, department_id),
		    location_id = COALESCE($3, location_id)
		WHERE id = $4`, in.FullName, in.DepartmentID, in.LocationID, employeeID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Deactivate soft-deletes: sets status to inactive rather than removing the
// row. A hard delete would either violate the FK from attendances/devices/
// face_embeddings (if they still reference this employee) or, worse, cascade
// and erase real attendance history along with the person who left.
func (s *Service) Deactivate(ctx context.Context, employeeID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE employees SET status = 'inactive' WHERE id = $1`, employeeID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrEmployeeNotFound
	}
	return nil
}

func randomPassword() (string, error) {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// Me is what the check-in screen needs to greet someone by name and show
// today's own record: without it the screen says "Karyawan" and "Masuk —"
// even right after a successful check-in.
type Me struct {
	EmployeeID string     `json:"employee_id"`
	FullName   string     `json:"full_name"`
	Status     string     `json:"status"`
	CheckInAt  *time.Time `json:"check_in_at"`
	CheckOutAt *time.Time `json:"check_out_at"`
}

func (s *Service) Me(ctx context.Context, userID string) (*Me, error) {
	var me Me
	err := s.pool.QueryRow(ctx, `
		SELECT e.id, e.full_name, e.status,
		       (SELECT a.occurred_at FROM attendances a
		         WHERE a.employee_id = e.id AND a.type = 'check_in'
		           AND a.occurred_at::date = CURRENT_DATE
		         ORDER BY a.occurred_at LIMIT 1),
		       (SELECT a.occurred_at FROM attendances a
		         WHERE a.employee_id = e.id AND a.type = 'check_out'
		           AND a.occurred_at::date = CURRENT_DATE
		         ORDER BY a.occurred_at DESC LIMIT 1)
		FROM employees e
		WHERE e.user_id = $1`, userID).
		Scan(&me.EmployeeID, &me.FullName, &me.Status, &me.CheckInAt, &me.CheckOutAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEmployeeRecord
	}
	if err != nil {
		return nil, err
	}
	return &me, nil
}

type Department struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Service) ListDepartments(ctx context.Context) ([]Department, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name FROM departments ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Department
	for rows.Next() {
		var d Department
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) CreateDepartment(ctx context.Context, name string) (*Department, error) {
	d := Department{Name: name}
	err := s.pool.QueryRow(ctx, `INSERT INTO departments (name) VALUES ($1) RETURNING id`, name).Scan(&d.ID)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Device is a registered browser for an employee (docs/PRD.md 7.4). A device
// beyond the per-employee limit is stored unapproved so HR can see and approve
// it; without this listing the limit would be a dead end for the employee.
type Device struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employee_id"`
	FullName   string    `json:"full_name"`
	UserAgent  string    `json:"user_agent"`
	Approved   bool      `json:"approved"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Service) ListDevices(ctx context.Context) ([]Device, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.employee_id, e.full_name, COALESCE(d.user_agent, ''), d.approved, d.created_at
		FROM devices d JOIN employees e ON e.id = d.employee_id
		ORDER BY d.approved, d.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.EmployeeID, &d.FullName, &d.UserAgent, &d.Approved, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) ApproveDevice(ctx context.Context, deviceID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE devices SET approved = true WHERE id = $1`, deviceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDeviceNotFound
	}
	return nil
}
