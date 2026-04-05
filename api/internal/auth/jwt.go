package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	dbpkg "github.com/phathdt/login-oauth/api/internal/db"
	"github.com/phathdt/login-oauth/api/internal/config"
)

type Claims struct {
	UserID  string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

type contextKey string

const ClaimsKey contextKey = "claims"

func GenerateAccessToken(cfg *config.Config, user dbpkg.User) (string, error) {
	claims := Claims{
		UserID:  user.ID.String(),
		Email:   user.Email,
		Name:    user.Name.String,
		Picture: user.Picture.String,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func ValidateToken(cfg *config.Config, tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
