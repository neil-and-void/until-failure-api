package handlers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/handlers/validators"

	"gorm.io/gorm"
)

func (h Handler) CreateUser(c *fiber.Ctx) error {
	userCreatedEvent := validators.UserCreatedEvent{}

	if err := c.BodyParser(&userCreatedEvent); err != nil {
		return err
	}

	if len(userCreatedEvent.Data.EmailAddresses) == 0 {
		return errors.New("No email addresses supplied")
	}

	email := userCreatedEvent.Data.EmailAddresses[0].EmailAddress
	id := userCreatedEvent.Data.ID

	_, err := h.DB.GetUser(id)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	user := database.User{
		ID:    id,
		Email: email,
	}

	err = h.DB.CreateUser(user)
	if err != nil {
		return err
	}

	fmt.Println(user)

	return c.SendString("ok")
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
