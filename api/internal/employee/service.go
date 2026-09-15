// Package employee covers master data: employees, departments, locations, schedules.
package employee

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailTaken = errors.New("email sudah terdaftar")
var ErrNIKTaken = errors.New("NIK sudah terdaftar")
var ErrNoEmployeeRecord = errors.New("akun ini tidak terhubung ke data karyawan")
var ErrDeviceNotFound = errors.New("perangkat tidak ditemukan")
var ErrEmployeeNotFound = errors.New("karyawan tidak ditemukan")
var ErrWeakPassword = errors.New("password minimal 8 karakter")
var ErrWrongPassword = errors.New("password lama tidak cocok")
var ErrInvalidStatus = errors.New("status karyawan tidak valid")
var ErrSelfManager = errors.New("karyawan tidak bisa menjadi atasan dirinya sendiri")

// MinPasswordLength is the floor for any password a person types in (HR
// resetting someone's, or an employee changing their own). The generated
// temp password from Create is longer than this by construction.
const MinPasswordLength = 8

var validStatuses = map[string]bool{"pending_enrollment": true, "active": true, "inactive": true}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Employee struct {
	ID               string `json:"id"`
	NIK              string `json:"nik"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	DepartmentID     string `json:"department_id,omitempty"`
	LocationID       string `json:"location_id,omitempty"`
	ScheduleID       string `json:"schedule_id,omitempty"`
	ManagerID        string `json:"manager_id,omitempty"`
	AnnualLeaveQuota int    `json:"annual_leave_quota"`
	Status           string `json:"status"`
	HasFacePhoto     bool   `json:"has_face_photo"`
	// AllowRemote lets attendance pass the office-network check from anywhere
	// (internal/attendance/service.go). Defaults to false: a deployment with no
	// office IP allowlist configured must turn this on per employee, or every
	// check-in fails "di luar jaringan kantor" regardless of who it is.
	AllowRemote bool `json:"allow_remote"`
}

type CreateEmployeeInput struct {
	NIK          string
	FullName     string
	Email        string
	DepartmentID *string
	LocationID   *string
	ScheduleID   *string
	AllowRemote  bool
}

// UpdateEmployeeInput edits the employee record and, since the login lives in
// a joined users row, its email and password too. Every field HR can see is
// editable here; the pointer fields mean "leave alone when omitted" rather
// than "clear", so a partial payload cannot silently wipe data.
//
// NIK is editable but uniqueness-checked: HR does need to fix a typo'd payroll
// number, and refusing that just pushes them to delete and recreate the person,
// which loses the attendance history hanging off employee_id.
type UpdateEmployeeInput struct {
	NIK              string
	FullName         string
	Email            string
	DepartmentID     *string
	LocationID       *string
	ScheduleID       *string
	ManagerID        *string
	AnnualLeaveQuota *int
	Status           *string
	// AllowRemote is a pointer for the same reason as the other optional
	// fields: nil leaves the stored value alone, so a partial edit (e.g. just
	// fixing a typo'd name) cannot silently flip this back off.
	AllowRemote *bool
	// Password, when non-empty, resets the employee's login. Empty means
	// "keep the current password", so the common edit does not touch it.
	Password string
}

type CreateEmployeeResult struct {
	Employee     Employee
	TempPassword string
}

func (s *Service) List(ctx context.Context) ([]Employee, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.nik, e.full_name, u.email,
		       COALESCE(e.department_id::text, ''), COALESCE(e.location_id::text, ''),
		       COALESCE(e.schedule_id::text, ''), COALESCE(e.manager_id::text, ''),
		       e.annual_leave_quota, e.status, e.allow_remote, e.face_photo_path IS NOT NULL
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
		if err := rows.Scan(&e.ID, &e.NIK, &e.FullName, &e.Email, &e.DepartmentID, &e.LocationID,
			&e.ScheduleID, &e.ManagerID, &e.AnnualLeaveQuota, &e.Status, &e.AllowRemote, &e.HasFacePhoto); err != nil {
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

	e := Employee{NIK: in.NIK, FullName: in.FullName, Email: in.Email, Status: "pending_enrollment", AllowRemote: in.AllowRemote}
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
		INSERT INTO employees (user_id, nik, full_name, department_id, location_id, schedule_id, status, allow_remote)
		VALUES ($1, $2, $3, $4, $5,
			COALESCE($6, (SELECT id FROM work_schedules ORDER BY created_at LIMIT 1)),
			'pending_enrollment', $7)
		RETURNING id`,
		userID, in.NIK, in.FullName, in.DepartmentID, in.LocationID, in.ScheduleID, in.AllowRemote).Scan(&e.ID); err != nil {
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

	if in.NIK != "" {
		var nikTaken bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM employees WHERE nik = $1 AND id != $2)`, in.NIK, employeeID).Scan(&nikTaken); err != nil {
			return err
		}
		if nikTaken {
			return ErrNIKTaken
		}
	}

	if in.Status != nil && !validStatuses[*in.Status] {
		return ErrInvalidStatus
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET email = $1 WHERE id = $2`, in.Email, userID); err != nil {
		return err
	}

	// Resetting the password is opt-in: an empty value leaves the current one
	// alone, so the everyday "fix a typo in the name" edit does not lock the
	// employee out of an account they are already using.
	if in.Password != "" {
		if len(in.Password) < MinPasswordLength {
			return ErrWeakPassword
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, string(hash), userID); err != nil {
			return err
		}
		// Existing sessions keep working off a refresh token the old password
		// issued, which defeats the point of a reset, so cut them here.
		if _, err := tx.Exec(ctx,
			`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
			return err
		}
	}

	if in.ManagerID != nil && *in.ManagerID == employeeID {
		return ErrSelfManager
	}

	// Three distinct meanings per optional field, which a plain COALESCE cannot
	// express: omitted (nil) leaves the value alone, "" clears it, and a uuid
	// sets it. Without the middle case HR could assign a department but never
	// remove one. Omission must stay non-destructive -- a partial payload
	// silently wiping data is the bug already fixed once for work_days.
	if _, err := tx.Exec(ctx, `
		UPDATE employees
		SET full_name = $1,
		    nik = COALESCE(NULLIF($2, ''), nik),
		    department_id = CASE WHEN $3::text IS NULL THEN department_id ELSE NULLIF($3, '')::uuid END,
		    location_id   = CASE WHEN $4::text IS NULL THEN location_id   ELSE NULLIF($4, '')::uuid END,
		    schedule_id   = CASE WHEN $5::text IS NULL THEN schedule_id   ELSE NULLIF($5, '')::uuid END,
		    manager_id    = CASE WHEN $6::text IS NULL THEN manager_id    ELSE NULLIF($6, '')::uuid END,
		    annual_leave_quota = COALESCE($7, annual_leave_quota),
		    status = COALESCE($8, status),
		    allow_remote = COALESCE($9, allow_remote)
		WHERE id = $10`,
		in.FullName, in.NIK, in.DepartmentID, in.LocationID, in.ScheduleID,
		in.ManagerID, in.AnnualLeaveQuota, in.Status, in.AllowRemote, employeeID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// FacePhotoPath returns the stored path of the employee's enrollment reference
// photo, or "" if none was ever saved (PHOTO_DIR unset at enroll time, or
// enrollment predates this feature).
func (s *Service) FacePhotoPath(ctx context.Context, employeeID string) (string, error) {
	var path *string
	err := s.pool.QueryRow(ctx, `SELECT face_photo_path FROM employees WHERE id = $1`, employeeID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrEmployeeNotFound
	}
	if err != nil {
		return "", err
	}
	if path == nil {
		return "", nil
	}
	return *path, nil
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

// Profile is the logged-in person's own account, for the header menu. It has
// to work for an HR/superadmin account too, which has no employees row at all
// -- that account previously had no name anywhere and rendered as "Admin".
type Profile struct {
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	NIK        string `json:"nik,omitempty"`
	IsEmployee bool   `json:"is_employee"`
}

type UpdateProfileInput struct {
	FullName        string
	Email           string
	CurrentPassword string
	NewPassword     string
}

func (s *Service) Profile(ctx context.Context, userID string) (*Profile, error) {
	var p Profile
	var employeeName, nik *string
	err := s.pool.QueryRow(ctx, `
		SELECT u.email, u.role::text, COALESCE(u.display_name, ''), e.full_name, e.nik
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id
		WHERE u.id = $1`, userID).Scan(&p.Email, &p.Role, &p.FullName, &employeeName, &nik)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEmployeeNotFound
	}
	if err != nil {
		return nil, err
	}
	// The employee record wins when there is one: that is the name HR manages.
	if employeeName != nil {
		p.FullName = *employeeName
		p.IsEmployee = true
	}
	if nik != nil {
		p.NIK = *nik
	}
	if p.FullName == "" {
		p.FullName = p.Email
	}
	return &p, nil
}

// UpdateProfile lets someone edit their own name, email and password. Changing
// the password requires the current one, because an unattended logged-in
// session must not be enough to lock the real owner out of their account.
func (s *Service) UpdateProfile(ctx context.Context, userID string, in UpdateProfileInput) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var currentHash string
	var hasEmployee bool
	if err := tx.QueryRow(ctx, `
		SELECT u.password_hash, EXISTS(SELECT 1 FROM employees e WHERE e.user_id = u.id)
		FROM users u WHERE u.id = $1`, userID).Scan(&currentHash, &hasEmployee); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEmployeeNotFound
		}
		return err
	}

	if in.Email != "" {
		var taken bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`, in.Email, userID).Scan(&taken); err != nil {
			return err
		}
		if taken {
			return ErrEmailTaken
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET email = $1 WHERE id = $2`, in.Email, userID); err != nil {
			return err
		}
	}

	if in.FullName != "" {
		if hasEmployee {
			if _, err := tx.Exec(ctx,
				`UPDATE employees SET full_name = $1 WHERE user_id = $2`, in.FullName, userID); err != nil {
				return err
			}
		} else if _, err := tx.Exec(ctx,
			`UPDATE users SET display_name = $1 WHERE id = $2`, in.FullName, userID); err != nil {
			return err
		}
	}

	if in.NewPassword != "" {
		if len(in.NewPassword) < MinPasswordLength {
			return ErrWeakPassword
		}
		if bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(in.CurrentPassword)) != nil {
			return ErrWrongPassword
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, string(hash), userID); err != nil {
			return err
		}
		// Other devices keep a refresh token minted under the old password;
		// a password change that leaves them alive is not really a change.
		if _, err := tx.Exec(ctx,
			`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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

var ErrDepartmentNotFound = errors.New("departemen tidak ditemukan")
var ErrDepartmentInUse = errors.New("departemen masih dipakai oleh karyawan")

func (s *Service) UpdateDepartment(ctx context.Context, id, name string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE departments SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDepartmentNotFound
	}
	return nil
}

// DeleteDepartment refuses rather than cascading: employees.department_id
// references this row, and quietly detaching people from their department to
// satisfy a delete loses information HR did not agree to lose.
func (s *Service) DeleteDepartment(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM departments WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrDepartmentInUse
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDepartmentNotFound
	}
	return nil
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
