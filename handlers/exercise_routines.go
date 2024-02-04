package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h Handler) CreateExerciseRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) UpdateExerciseRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a PATCH request!")
}

func (h Handler) DeleteExerciseRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a DELETE request!")
}
