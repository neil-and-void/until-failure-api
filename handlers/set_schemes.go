package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h Handler) CreateSetScheme(c *fiber.Ctx) error {
	newSetScheme := NewSetScheme{}
	if err := c.BodyParser(&newSetScheme); err != nil {
		return err
	}

	if err := h.Validate.Struct(newSetScheme); err != nil {
		// Return error to client
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return nil
}
