package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/phathdt/login-oauth/api/internal/auth"
	"github.com/phathdt/login-oauth/api/internal/config"
	"github.com/phathdt/login-oauth/api/internal/models"
)

type OAuthHandler struct {
	cfg         *config.Config
	oauthConfig *oauth2.Config
	db          *pgxpool.Pool
}

func NewOAuthHandler(cfg *config.Config, oauthConfig *oauth2.Config, db *pgxpool.Pool) *OAuthHandler {
	return &OAuthHandler{cfg: cfg, oauthConfig: oauthConfig, db: db}
}

func (h *OAuthHandler) Login(c *fiber.Ctx) error {
	state, err := auth.GenerateState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate state"})
	}

	url := h.oauthConfig.AuthCodeURL(state)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

func (h *OAuthHandler) Callback(c *fiber.Ctx) error {
	state := c.Query("state")
	if !auth.ValidateState(state) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid state"})
	}

	code := c.Query("code")
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to exchange code"})
	}

	userInfo, err := fetchGoogleUserInfo(token.AccessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch user info"})
	}

	user, err := models.FindOrCreateByGoogle(h.db, userInfo.Sub, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to upsert user"})
	}

	accessToken, err := auth.GenerateAccessToken(h.cfg, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate access token"})
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshToken, err := models.CreateRefreshToken(h.db, user.ID, expiresAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create refresh token"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   h.cfg.Env == "production",
		MaxAge:   7 * 24 * 3600,
		Path:     "/",
	})

	redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s", h.cfg.FrontendURL, accessToken)
	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func fetchGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	return &info, nil
}
