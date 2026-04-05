package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/phathdt/login-oauth/api/internal/config"
)

type stateEntry struct {
	expiresAt time.Time
}

var (
	stateStore sync.Map
)

func NewOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  "http://localhost:3000/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	stateStore.Store(state, stateEntry{expiresAt: time.Now().Add(10 * time.Minute)})
	return state, nil
}

func ValidateState(state string) bool {
	val, ok := stateStore.LoadAndDelete(state)
	if !ok {
		return false
	}

	entry := val.(stateEntry)
	return time.Now().Before(entry.expiresAt)
}
