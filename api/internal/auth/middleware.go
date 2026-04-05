package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/phathdt/login-oauth/api/internal/config"
)

func JWTAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := ValidateToken(cfg, tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}
