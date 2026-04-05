package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/phathdt/login-oauth/api/internal/auth"
	"github.com/phathdt/login-oauth/api/internal/config"
	"github.com/phathdt/login-oauth/api/internal/models"
)

type AuthHandler struct {
	cfg *config.Config
	db  *pgxpool.Pool
}

func NewAuthHandler(cfg *config.Config, db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: db}
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshTokenVal := c.Cookies("refresh_token")
	if refreshTokenVal == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing refresh token"})
	}

	userID, err := models.ValidateRefreshToken(h.db, refreshTokenVal)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	user, err := models.FindUserByID(h.db, userID)
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
		models.RevokeRefreshToken(h.db, refreshTokenVal)
	}

	c.Cookie(&fiber.Cookie{
		Name:    "refresh_token",
		Value:   "",
		MaxAge:  -1,
		Path:    "/",
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*auth.Claims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	user := &models.User{
		ID:      claims.UserID,
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
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
