package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken struct {
	ID        int        `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func hashToken(plainToken string) string {
	sum := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(sum[:])
}

func CreateRefreshToken(db *pgxpool.Pool, userID string, expiresAt time.Time) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token bytes: %w", err)
	}

	plainToken := base64.URLEncoding.EncodeToString(b)
	tokenHash := hashToken(plainToken)

	_, err := db.Exec(context.Background(),
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return plainToken, nil
}

func ValidateRefreshToken(db *pgxpool.Pool, plainToken string) (string, error) {
	tokenHash := hashToken(plainToken)

	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time

	err := db.QueryRow(context.Background(),
		`SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &expiresAt, &revokedAt)

	if err != nil {
		return "", fmt.Errorf("refresh token not found")
	}

	if revokedAt != nil {
		return "", fmt.Errorf("refresh token has been revoked")
	}

	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("refresh token has expired")
	}

	return userID, nil
}

func RevokeRefreshToken(db *pgxpool.Pool, plainToken string) error {
	tokenHash := hashToken(plainToken)

	_, err := db.Exec(context.Background(),
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
