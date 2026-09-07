// Package auth handles login, JWT issuance, and refresh-token rotation.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")

const refreshTokenTTL = 30 * 24 * time.Hour // docs/PRD.md F1: access 15min + refresh cookie

type Service struct {
	pool   *pgxpool.Pool
	secret []byte
}

func NewService(pool *pgxpool.Pool, secret string) *Service {
	return &Service{pool: pool, secret: []byte(secret)}
}

type Claims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	RefreshTTL   time.Duration
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	var id, passwordHash, role string
	row := s.pool.QueryRow(ctx, `SELECT id, password_hash, role FROM users WHERE email = $1`, email)
	if err := row.Scan(&id, &passwordHash, &role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issuePair(ctx, id, role)
}

// Refresh validates the presented refresh token, revokes it, and issues a
// fresh pair (rotation) -- a stolen, already-used token becomes a signal,
// not a standing credential.
func (s *Service) Refresh(ctx context.Context, presentedToken string) (*TokenPair, error) {
	hash := hashToken(presentedToken)

	var id, userID, role string
	var expiresAt time.Time
	var revokedAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT rt.id, rt.user_id, u.role, rt.expires_at, rt.revoked_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1`, hash).Scan(&id, &userID, &role, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, err
	}
	if revokedAt != nil || time.Now().After(expiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	if _, err := s.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1`, id); err != nil {
		return nil, err
	}

	return s.issuePair(ctx, userID, role)
}

func (s *Service) Logout(ctx context.Context, presentedToken string) error {
	hash := hashToken(presentedToken)
	_, err := s.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`, hash)
	return err
}

func (s *Service) issuePair(ctx context.Context, userID, role string) (*TokenPair, error) {
	access, err := s.issueAccessToken(userID, role)
	if err != nil {
		return nil, err
	}

	refresh, err := randomToken()
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, hashToken(refresh), time.Now().Add(refreshTokenTTL))
	if err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: access, RefreshToken: refresh, RefreshTTL: refreshTokenTTL}, nil
}

func (s *Service) issueAccessToken(userID, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
