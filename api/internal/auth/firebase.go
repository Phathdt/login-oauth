package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/phathdt/login-oauth/api/internal/config"
	"google.golang.org/api/option"
)

// FirebaseClient wraps the Firebase Admin Auth client for ID token verification.
type FirebaseClient struct {
	client *firebaseauth.Client
}

// FirebaseTokenInfo holds the verified claims extracted from a Firebase ID token.
type FirebaseTokenInfo struct {
	UID      string
	Provider string // 'google', 'github', 'microsoft', etc.
	Email    string
	Name     string
	Picture  string
}

// NewFirebaseClient initialises the Firebase Admin SDK.
// Credentials are loaded from FIREBASE_CREDENTIALS_JSON (raw service account JSON string).
// Falls back to Application Default Credentials (ADC) if the env var is empty.
func NewFirebaseClient(cfg *config.Config) (*FirebaseClient, error) {
	ctx := context.Background()

	var opts []option.ClientOption
	if cfg.FirebaseCredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.FirebaseCredentialsJSON)))
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: cfg.FirebaseProjectID,
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("get firebase auth client: %w", err)
	}

	return &FirebaseClient{client: client}, nil
}

// VerifyIDToken verifies a Firebase ID token and returns the decoded claims.
func (f *FirebaseClient) VerifyIDToken(ctx context.Context, idToken string) (*FirebaseTokenInfo, error) {
	token, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("verify firebase id token: %w", err)
	}

	info := &FirebaseTokenInfo{
		UID:      token.UID,
		Provider: extractProvider(token),
	}

	if v, ok := token.Claims["email"].(string); ok {
		info.Email = v
	}
	if v, ok := token.Claims["name"].(string); ok {
		info.Name = v
	}
	if v, ok := token.Claims["picture"].(string); ok {
		info.Picture = v
	}

	return info, nil
}

// extractProvider maps Firebase sign_in_provider to a short provider name.
func extractProvider(token *firebaseauth.Token) string {
	// firebase.sign_in_provider is nested under "firebase" claim
	if firebase, ok := token.Claims["firebase"].(map[string]interface{}); ok {
		if p, ok := firebase["sign_in_provider"].(string); ok {
			switch p {
			case "google.com":
				return "google"
			case "github.com":
				return "github"
			case "microsoft.com":
				return "microsoft"
			case "password":
				return "email"
			default:
				return p
			}
		}
	}
	return "unknown"
}
