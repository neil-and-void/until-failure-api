package handlers

import "github.com/gofiber/fiber/v2"

func (h Handler) CreateWorkout(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) GetWorkouts(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}

func (h Handler) UpdateWorkout(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}
