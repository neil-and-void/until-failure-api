package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/neilZon/workout-logger-api/database"
	"gorm.io/gorm"
)

func (h Handler) CreateUser(c *fiber.Ctx) error {
	userId := c.Params("userId")
	user, err := h.DB.GetUser(userId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	user = database.User{}

	err = h.DB.CreateUser(user)
	if err != nil {
		return err
	}

	return c.JSON(user)
}

func (h Handler) GetUser(c *fiber.Ctx) error {
	userId := c.Params("userId")

	user, err := h.DB.GetUser(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user does not exist")
		}
		return err
	}

	return c.JSON(user)
}
