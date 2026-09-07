// Package enrollment covers first-time face registration (docs/PRD.md 6.1).
// Every submitted photo is re-validated on the server; the browser's own
// framing/quality checks are a UX convenience, never the gate.
package enrollment

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/hris-face/api/internal/faceclient"
)

const (
	minPhotos          = 3
	maxPhotos          = 5
	minQualityScore    = 0.5  // mirrors face/app/face_engine.py MIN_FACE_WIDTH_PX gate
	duplicateThreshold = 0.45 // same value as attendance match threshold, docs/PRD.md section 9
)

var (
	ErrAlreadyEnrolled  = errors.New("wajah sudah terdaftar; gunakan re-enroll untuk menggantinya")
	ErrTooFewPhotos     = errors.New("minimal 3 foto diperlukan")
	ErrTooManyPhotos    = errors.New("maksimal 5 foto")
	ErrEmployeeNotFound = errors.New("data karyawan tidak ditemukan")
	ErrDuplicateFace    = errors.New("wajah terdeteksi mirip dengan karyawan lain terdaftar; hubungi HR")
)

// PhotoRejection explains, per photo index, why it failed the quality gate --
// so the UI can tell the employee exactly which shot to retake.
type PhotoRejection struct {
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

type RejectedError struct {
	Rejections []PhotoRejection
}

func (e *RejectedError) Error() string {
	return fmt.Sprintf("%d foto ditolak", len(e.Rejections))
}

type Service struct {
	pool *pgxpool.Pool
	face *faceclient.Client
}

func NewService(pool *pgxpool.Pool, face *faceclient.Client) *Service {
	return &Service{pool: pool, face: face}
}

// Enroll registers a face for the first time. It refuses when the employee is
// already enrolled so a stray second submission cannot quietly stack a new face
// onto an existing account.
func (s *Service) Enroll(ctx context.Context, userID string, photos [][]byte) error {
	return s.enroll(ctx, userID, photos, false)
}

// ReEnroll replaces the stored face (docs/PRD.md F3: beard, glasses, weight).
// Old embeddings are archived (is_active = false), never deleted -- an audit
// asking "which face was on file when this attendance was recorded?" must still
// be answerable.
func (s *Service) ReEnroll(ctx context.Context, userID string, photos [][]byte) error {
	return s.enroll(ctx, userID, photos, true)
}

func (s *Service) enroll(ctx context.Context, userID string, photos [][]byte, replace bool) error {
	if len(photos) < minPhotos {
		return ErrTooFewPhotos
	}
	if len(photos) > maxPhotos {
		return ErrTooManyPhotos
	}

	var employeeID string
	err := s.pool.QueryRow(ctx, `SELECT id FROM employees WHERE user_id = $1`, userID).Scan(&employeeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEmployeeNotFound
	}
	if err != nil {
		return err
	}

	var existing int
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM face_embeddings WHERE employee_id = $1 AND is_active`,
		employeeID).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 && !replace {
		return ErrAlreadyEnrolled
	}

	// Keep each embedding with its measured quality; storing the constant
	// instead makes face_embeddings.quality_score useless for later triage
	// ("which photo should this employee re-take?").
	type scored struct {
		embedding []float32
		quality   float32
	}
	embeddings := make([]scored, 0, len(photos))
	var rejections []PhotoRejection

	for i, photo := range photos {
		analysis, err := s.face.Analyze(photo)
		if err != nil {
			return err
		}
		if analysis.FaceCount == 0 {
			rejections = append(rejections, PhotoRejection{Index: i, Reason: "wajah tidak terdeteksi"})
			continue
		}
		if analysis.FaceCount > 1 {
			rejections = append(rejections, PhotoRejection{Index: i, Reason: "lebih dari satu wajah terdeteksi"})
			continue
		}
		if analysis.QualityScore < minQualityScore {
			rejections = append(rejections, PhotoRejection{Index: i, Reason: "wajah terlalu kecil atau gambar kurang jelas"})
			continue
		}
		embeddings = append(embeddings, scored{embedding: analysis.Embedding, quality: analysis.QualityScore})
	}

	if len(rejections) > 0 {
		return &RejectedError{Rejections: rejections}
	}

	for _, e := range embeddings {
		collided, err := s.collidesWithAnotherEmployee(ctx, employeeID, e.embedding)
		if err != nil {
			return err
		}
		if collided {
			s.flagDuplicateForReview(ctx, employeeID) //nolint:errcheck
			return ErrDuplicateFace
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if replace {
		if _, err := tx.Exec(ctx,
			`UPDATE face_embeddings SET is_active = false WHERE employee_id = $1 AND is_active`,
			employeeID); err != nil {
			return err
		}
	}

	for _, e := range embeddings {
		vec := pgvector.NewVector(e.embedding)
		if _, err := tx.Exec(ctx, `
			INSERT INTO face_embeddings (employee_id, embedding, quality_score, is_active)
			VALUES ($1, $2, $3, true)`, employeeID, vec, e.quality); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE employees SET status = 'active' WHERE id = $1`, employeeID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) collidesWithAnotherEmployee(ctx context.Context, employeeID string, embedding []float32) (bool, error) {
	vec := pgvector.NewVector(embedding)
	var best float32
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(1 - (fe.embedding <=> $1)), 0)
		FROM face_embeddings fe
		WHERE fe.is_active AND fe.employee_id != $2`, vec, employeeID).Scan(&best)
	if err != nil {
		return false, err
	}
	return best >= duplicateThreshold, nil
}

func (s *Service) flagDuplicateForReview(ctx context.Context, employeeID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_logs (action, entity, entity_id, after_data)
		VALUES ('enrollment_duplicate_flagged', 'employee', $1, '{}'::jsonb)`, employeeID)
	return err
}
