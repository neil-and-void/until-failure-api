package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h Handler) CreateRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) GetRoutines(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) UpdateRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) DeleteRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}
