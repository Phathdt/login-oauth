package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	dbpkg "github.com/phathdt/login-oauth/api/internal/db"
)

func hashToken(plainToken string) string {
	sum := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(sum[:])
}

// CreateRefreshToken generates a random token, stores its hash via sqlc, and returns the plain token.
func CreateRefreshToken(q *dbpkg.Queries, userID string, expiresAt time.Time) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token bytes: %w", err)
	}

	plainToken := base64.URLEncoding.EncodeToString(b)
	tokenHash := hashToken(plainToken)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}

	_, err = q.CreateRefreshToken(context.Background(), dbpkg.CreateRefreshTokenParams{
		UserID:    uid,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return plainToken, nil
}

// ValidateRefreshToken hashes the plain token and looks it up via sqlc (filters expired/revoked in SQL).
func ValidateRefreshToken(q *dbpkg.Queries, plainToken string) (string, error) {
	tokenHash := hashToken(plainToken)

	rt, err := q.FindRefreshToken(context.Background(), tokenHash)
	if err != nil {
		return "", fmt.Errorf("refresh token not found or expired")
	}

	return rt.UserID.String(), nil
}

// RevokeRefreshToken hashes the plain token and marks it revoked via sqlc.
func RevokeRefreshToken(q *dbpkg.Queries, plainToken string) error {
	tokenHash := hashToken(plainToken)

	if err := q.RevokeRefreshToken(context.Background(), tokenHash); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
