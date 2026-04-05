package handlers

import (
	"context"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/phathdt/login-oauth/api/internal/auth"
	"github.com/phathdt/login-oauth/api/internal/config"
	dbpkg "github.com/phathdt/login-oauth/api/internal/db"
	"github.com/phathdt/login-oauth/api/internal/models"
)

type AuthHandler struct {
	cfg     *config.Config
	queries *dbpkg.Queries
}

func NewAuthHandler(cfg *config.Config, queries *dbpkg.Queries) *AuthHandler {
	return &AuthHandler{cfg: cfg, queries: queries}
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
