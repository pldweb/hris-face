// Package attendance implements check-in/check-out with face verification.
// Every decision (match, liveness, timestamp, status) is made here on the server;
// the browser only frames the shot. See docs/PRD.md sections 6.2, 7.4, 9.
package attendance

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/hris-face/api/internal/faceclient"
)

const (
	matchThreshold         = 0.45
	lowConfidenceMarginGap = 0.05
	minLivenessScore       = 0.5
	maxDevicesPerEmployee  = 2 // docs/PRD.md 7.4
)

var (
	ErrNoFaceMatch       = errors.New("wajah tidak dikenali")
	ErrLivenessFailed    = errors.New("verifikasi keaslian wajah gagal")
	ErrOutsideOfficeNet  = errors.New("di luar jaringan kantor")
	ErrNoCheckInYet      = errors.New("belum ada absen masuk hari ini")
	ErrNoEmployeeRecord  = errors.New("akun ini tidak terhubung ke data karyawan")
	ErrDeviceNotApproved = errors.New("perangkat ini belum disetujui HR")
)

// ErrWrongPerson is returned when the frame matches an enrolled face, just not
// the one logged in. Unlike ErrNoFaceMatch it names the employee whose face was
// recognised, at the requester's request -- this intentionally surfaces another
// employee's identity to whoever is holding the camera, with no enrollment
// consent check, so any future report of it being misused should point here.
type ErrWrongPerson struct {
	Name string
}

func (e *ErrWrongPerson) Error() string {
	return "wajah cocok dengan karyawan lain: " + e.Name
}

type Service struct {
	pool        *pgxpool.Pool
	face        *faceclient.Client
	officeCIDRs []*net.IPNet
	timezone    *time.Location
	photos      *PhotoStore
}

func NewService(pool *pgxpool.Pool, face *faceclient.Client, officeAllowlist []string, tz *time.Location, photos *PhotoStore) *Service {
	var cidrs []*net.IPNet
	for _, entry := range officeAllowlist {
		if _, ipNet, err := net.ParseCIDR(entry); err == nil {
			cidrs = append(cidrs, ipNet)
		}
	}
	if tz == nil {
		tz = time.UTC
	}
	return &Service{pool: pool, face: face, officeCIDRs: cidrs, timezone: tz, photos: photos}
}

type Result struct {
	EmployeeID    string    `json:"employee_id"`
	FullName      string    `json:"full_name"`
	OccurredAt    time.Time `json:"occurred_at"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	LowConfidence bool      `json:"low_confidence"`
}

type FaceScanResult struct {
	Found      bool    `json:"found"`
	FullName   string  `json:"full_name,omitempty"`
	Similarity float32 `json:"similarity,omitempty"`
}

// Scan checks whether a face belongs to an active employee without recording
// attendance. It intentionally skips liveness: this is a directory lookup,
// not proof of presence.
func (s *Service) Scan(ctx context.Context, imageJPEG []byte) (*FaceScanResult, error) {
	analysis, err := s.face.Analyze(imageJPEG)
	if err != nil {
		return nil, err
	}
	if analysis.FaceCount != 1 {
		return &FaceScanResult{}, nil
	}

	match, err := s.findBestMatch(ctx, analysis.Embedding)
	if err != nil {
		return nil, err
	}
	if match == nil || match.best < matchThreshold {
		return &FaceScanResult{}, nil
	}
	return &FaceScanResult{Found: true, FullName: match.fullName, Similarity: match.best}, nil
}

// CheckIn verifies the frame, enforces network/device rules, derives on_time vs
// late from the employee's schedule, and records the row.
// CheckIn's third return value is the device cookie to set, independent of
// err: a device can be resolved (and consume a slot) even when the attendance
// write itself then fails a business rule. The caller must set the cookie
// whenever this is non-empty, success or not -- see the comment on record().
func (s *Service) CheckIn(ctx context.Context, userID string, imageJPEG []byte, deviceKey, userAgent, clientIP string) (*Result, string, error) {
	return s.record(ctx, "check_in", userID, imageJPEG, deviceKey, userAgent, clientIP)
}

// CheckOut requires a check-in earlier the same day and marks early_leave when
// it happens before the schedule's end time.
func (s *Service) CheckOut(ctx context.Context, userID string, imageJPEG []byte, deviceKey, userAgent, clientIP string) (*Result, string, error) {
	return s.record(ctx, "check_out", userID, imageJPEG, deviceKey, userAgent, clientIP)
}

func (s *Service) record(ctx context.Context, kind, userID string, imageJPEG []byte, deviceKey, userAgent, clientIP string) (*Result, string, error) {
	// The session decides WHO is being recorded; the face only proves that person
	// is really the one in front of the camera. Identifying by face alone would
	// let anyone logged in mark a colleague present just by pointing the webcam
	// at them -- buddy punching with extra steps.
	sessionEmployeeID, err := s.employeeForUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	analysis, err := s.face.Analyze(imageJPEG)
	if err != nil {
		return nil, "", err
	}
	if analysis.LivenessScore < challengeFloor {
		return nil, "", ErrLivenessFailed
	}
	if analysis.LivenessScore < minLivenessScore {
		// The passive model is unsure. Rather than reject an employee who may
		// simply be badly lit, ask for the movement challenge (docs/PRD.md F5).
		return nil, "", ErrChallengeRequired
	}

	match, err := s.findBestMatch(ctx, analysis.Embedding)
	if err != nil {
		return nil, "", err
	}
	if match == nil || match.best < matchThreshold {
		return nil, "", ErrNoFaceMatch
	}
	if match.employeeID != sessionEmployeeID {
		return nil, "", &ErrWrongPerson{Name: match.fullName}
	}

	onSite := s.isOfficeIP(clientIP)
	if !onSite && !match.allowRemote {
		return nil, "", ErrOutsideOfficeNet
	}

	deviceID, issuedKey, err := s.resolveDevice(ctx, match.employeeID, deviceKey, userAgent)
	if err != nil {
		return nil, "", err
	}

	result, err := s.commitWithDevice(ctx, kind, match, analysis.LivenessScore, imageJPEG, deviceID, clientIP, time.Now())
	// issuedKey surfaces here even when err != nil: resolveDevice already
	// committed the device row to the database above, so the client must learn
	// its key regardless of what commitWithDevice does with the attendance rules.
	return result, issuedKey, err
}

// commit resolves the device itself; used by the challenge path, which has not
// touched device binding yet.
func (s *Service) commit(ctx context.Context, kind string, match *matchResult, liveness float32, imageJPEG []byte, deviceKey, userAgent, clientIP string, now time.Time) (*Result, string, error) {
	deviceID, issuedKey, err := s.resolveDevice(ctx, match.employeeID, deviceKey, userAgent)
	if err != nil {
		return nil, "", err
	}
	result, err := s.commitWithDevice(ctx, kind, match, liveness, imageJPEG, deviceID, clientIP, now)
	return result, issuedKey, err
}

// commitWithDevice applies the once-a-day rules, derives the status and writes
// the row. Both the direct and the challenge path end here so they cannot drift
// apart on what counts as a duplicate or how status is decided.
func (s *Service) commitWithDevice(ctx context.Context, kind string, match *matchResult, liveness float32, imageJPEG []byte, deviceID *string, clientIP string, now time.Time) (*Result, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	// Serialises submits for one employee: auto-capture and a button click can
	// land together, and without the lock both read "not yet marked" and insert.
	// todayMarks runs after the lock is held, so it sees the other request's commit.
	if _, err := tx.Exec(ctx, `SELECT 1 FROM employees WHERE id = $1 FOR UPDATE`, match.employeeID); err != nil {
		return nil, err
	}

	checkedIn, checkedOut, err := s.todayMarks(ctx, match.employeeID)
	if err != nil {
		return nil, err
	}
	if kind == "check_out" && !checkedIn {
		return nil, ErrNoCheckInYet
	}
	// Presensi ulang: scanning again for a session already recorded today
	// overwrites that row instead of being rejected, at the employee's own
	// request and with no HR review -- source='face_rescan' is the only trace
	// that the original timestamp was replaced.
	rescan := (kind == "check_in" && checkedIn) || (kind == "check_out" && checkedOut)

	status, err := s.deriveStatus(ctx, match.employeeID, kind, now)
	if err != nil {
		return nil, err
	}

	lowConfidence := (match.best - match.second) < lowConfidenceMarginGap
	photoPath := s.photos.Save(match.employeeID, now, imageJPEG)

	if rescan {
		_, err = tx.Exec(ctx, `
			UPDATE attendances
			SET occurred_at = $1, similarity = $2, liveness_score = $3, ip_address = $4,
			    device_id = $5, status = $6, low_confidence = $7,
			    photo_path = COALESCE(NULLIF($8, ''), photo_path), source = 'face_rescan'
			WHERE employee_id = $9 AND type = $10 AND occurred_at::date = CURRENT_DATE`,
			now, match.best, liveness, clientIP, deviceID, status, lowConfidence, photoPath, match.employeeID, kind)
	} else {
		_, err = tx.Exec(ctx, `
			INSERT INTO attendances
				(employee_id, type, occurred_at, similarity, liveness_score, ip_address, device_id, status, low_confidence, photo_path)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''))`,
			match.employeeID, kind, now, match.best, liveness, clientIP, deviceID, status, lowConfidence, photoPath)
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &Result{
		EmployeeID:    match.employeeID,
		FullName:      match.fullName,
		OccurredAt:    now,
		Type:          kind,
		Status:        status,
		LowConfidence: lowConfidence,
	}, nil
}

func (s *Service) employeeForUser(ctx context.Context, userID string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id FROM employees WHERE user_id = $1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoEmployeeRecord
	}
	return id, err
}

// deriveStatus compares the local wall-clock time against the employee's
// schedule. An employee with no schedule cannot be judged late, so they stay
// on_time rather than being penalised for missing master data.
func (s *Service) deriveStatus(ctx context.Context, employeeID, kind string, at time.Time) (string, error) {
	var start, end time.Time
	var tolerance int
	err := s.pool.QueryRow(ctx, `
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

// resolveDevice enforces docs/PRD.md 7.4: at most two approved devices per
// employee. A known device passes; a new one self-registers while under the
// limit; beyond that it is parked as a single pending row for HR to approve.
//
// Only ONE pending row is ever kept per employee: minting a new row per attempt
// would let a browser without cookies fill the table indefinitely.
func (s *Service) resolveDevice(ctx context.Context, employeeID, deviceKey, userAgent string) (deviceID *string, issuedKey string, err error) {
	if deviceKey != "" {
		var id, owner string
		var approved bool
		err := s.pool.QueryRow(ctx,
			`SELECT id, employee_id, approved FROM devices WHERE device_key = $1`, deviceKey).
			Scan(&id, &owner, &approved)
		switch {
		case err == nil && owner == employeeID:
			if !approved {
				return nil, "", ErrDeviceNotApproved
			}
			return &id, "", nil
		case err == nil:
			// Registered to someone else: this browser was used by another
			// employee. Fall through and treat it as a new device for this one.
		case errors.Is(err, pgx.ErrNoRows):
			// Unknown key (stale cookie); fall through to registration.
		default:
			return nil, "", err
		}
	}

	var approvedCount int
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM devices WHERE employee_id = $1 AND approved`, employeeID).Scan(&approvedCount); err != nil {
		return nil, "", err
	}

	if approvedCount >= maxDevicesPerEmployee {
		var pending string
		err := s.pool.QueryRow(ctx,
			`SELECT id FROM devices WHERE employee_id = $1 AND NOT approved LIMIT 1`, employeeID).Scan(&pending)
		if errors.Is(err, pgx.ErrNoRows) {
			key, kerr := randomKey()
			if kerr != nil {
				return nil, "", kerr
			}
			if _, ierr := s.pool.Exec(ctx, `
				INSERT INTO devices (employee_id, device_key, user_agent, approved)
				VALUES ($1, $2, $3, false)`, employeeID, key, userAgent); ierr != nil {
				return nil, "", ierr
			}
		} else if err != nil {
			return nil, "", err
		}
		return nil, "", ErrDeviceNotApproved
	}

	newKey, err := randomKey()
	if err != nil {
		return nil, "", err
	}
	var id string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO devices (employee_id, device_key, user_agent, approved)
		VALUES ($1, $2, $3, true) RETURNING id`,
		employeeID, newKey, userAgent).Scan(&id); err != nil {
		return nil, "", err
	}
	return &id, newKey, nil
}

func randomKey() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

type matchResult struct {
	employeeID  string
	fullName    string
	allowRemote bool
	best        float32
	second      float32
}

// findBestMatch scans all active embeddings and returns the best and second-best
// per-employee score. At <200 employees (~1,000 embeddings) this sequential scan
// is faster than maintaining an approximate index (docs/PRD.md section 4).
func (s *Service) findBestMatch(ctx context.Context, queryEmbedding []float32) (*matchResult, error) {
	vec := pgvector.NewVector(queryEmbedding)
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.full_name, e.allow_remote, 1 - (fe.embedding <=> $1) AS score
		FROM face_embeddings fe
		JOIN employees e ON e.id = fe.employee_id
		WHERE fe.is_active AND e.status = 'active'
		ORDER BY score DESC`, vec)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Rows arrive sorted by score DESC across all embeddings; since each employee
	// has multiple embeddings, the first row we see for an employee is already
	// their best score. We only need the best score of the top two distinct employees.
	seen := make(map[string]bool)
	var out *matchResult
	second := float32(-1)

	for rows.Next() {
		var id, name string
		var remote bool
		var score float32
		if err := rows.Scan(&id, &name, &remote, &score); err != nil {
			return nil, err
		}
		if seen[id] {
			continue
		}
		seen[id] = true

		if out == nil {
			out = &matchResult{employeeID: id, fullName: name, allowRemote: remote, best: score}
		} else {
			second = score
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out != nil {
		out.second = second
	}
	return out, nil
}

func (s *Service) todayMarks(ctx context.Context, employeeID string) (checkedIn, checkedOut bool, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT
			EXISTS(SELECT 1 FROM attendances WHERE employee_id = $1 AND type = 'check_in'  AND occurred_at::date = CURRENT_DATE),
			EXISTS(SELECT 1 FROM attendances WHERE employee_id = $1 AND type = 'check_out' AND occurred_at::date = CURRENT_DATE)
	`, employeeID).Scan(&checkedIn, &checkedOut)
	return checkedIn, checkedOut, err
}

func (s *Service) isOfficeIP(clientIP string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}
	for _, cidr := range s.officeCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
