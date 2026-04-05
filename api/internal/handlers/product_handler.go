package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/phathdt/login-oauth/api/internal/models"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

func (h *ProductHandler) List(c *fiber.Ctx) error {
	products := models.GetAllProducts()
	return c.JSON(products)
}
