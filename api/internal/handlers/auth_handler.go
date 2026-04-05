package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/phathdt/login-oauth/api/internal/auth"
	"github.com/phathdt/login-oauth/api/internal/config"
	dbpkg "github.com/phathdt/login-oauth/api/internal/db"
	"github.com/phathdt/login-oauth/api/internal/models"
)

type AuthHandler struct {
	cfg            *config.Config
	queries        *dbpkg.Queries
	firebaseClient *auth.FirebaseClient
}

func NewAuthHandler(cfg *config.Config, queries *dbpkg.Queries, firebaseClient *auth.FirebaseClient) *AuthHandler {
	return &AuthHandler{cfg: cfg, queries: queries, firebaseClient: firebaseClient}
}

type firebaseLoginRequest struct {
	IDToken string `json:"id_token"`
}

// FirebaseLogin verifies a Firebase ID token, upserts the user, and returns a backend JWT.
func (h *AuthHandler) FirebaseLogin(c *fiber.Ctx) error {
	var req firebaseLoginRequest
	if err := c.BodyParser(&req); err != nil || req.IDToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id_token is required"})
	}

	tokenInfo, err := h.firebaseClient.VerifyIDToken(c.Context(), req.IDToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid firebase token"})
	}

	user, err := h.queries.FindOrCreateUser(context.Background(), dbpkg.FindOrCreateUserParams{
		FirebaseUID: tokenInfo.UID,
		Provider:    tokenInfo.Provider,
		Email:       tokenInfo.Email,
		Name:        sql.NullString{String: tokenInfo.Name, Valid: tokenInfo.Name != ""},
		Picture:     sql.NullString{String: tokenInfo.Picture, Valid: tokenInfo.Picture != ""},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to upsert user"})
	}

	accessToken, err := auth.GenerateAccessToken(h.cfg, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate access token"})
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshToken, err := models.CreateRefreshToken(h.queries, user.ID.String(), expiresAt)
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

	return c.JSON(fiber.Map{
		"access_token": accessToken,
		"user": fiber.Map{
			"id":      user.ID.String(),
			"email":   user.Email,
			"name":    user.Name.String,
			"picture": user.Picture.String,
		},
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshTokenVal := c.Cookies("refresh_token")
	if refreshTokenVal == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing refresh token"})
	}

	userID, err := models.ValidateRefreshToken(h.queries, refreshTokenVal)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user id"})
	}

	user, err := h.queries.FindUserByID(context.Background(), uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}

	accessToken, err := auth.GenerateAccessToken(h.cfg, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate access token"})
	}

	return c.JSON(fiber.Map{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshTokenVal := c.Cookies("refresh_token")
	if refreshTokenVal != "" {
		models.RevokeRefreshToken(h.queries, refreshTokenVal)
	}

	c.Cookie(&fiber.Cookie{
		Name:   "refresh_token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*auth.Claims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Construct a db.User from JWT claims for token refresh — no DB round-trip needed.
	user := dbpkg.User{
		Email:   claims.Email,
		Name:    sql.NullString{String: claims.Name, Valid: claims.Name != ""},
		Picture: sql.NullString{String: claims.Picture, Valid: claims.Picture != ""},
	}
	if uid, err := uuid.Parse(claims.UserID); err == nil {
		user.ID = uid
	}

	freshToken, err := auth.GenerateAccessToken(h.cfg, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":      claims.UserID,
			"email":   claims.Email,
			"name":    claims.Name,
			"picture": claims.Picture,
		},
		"access_token": freshToken,
	})
}
