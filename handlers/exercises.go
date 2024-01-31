package handlers

import "github.com/gofiber/fiber/v2"

func (h Handler) CreateExericse(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) GetExercises(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) UpdateExercise(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}
