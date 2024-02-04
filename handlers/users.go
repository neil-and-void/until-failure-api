package handlers

import (
	"errors"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gofiber/fiber/v2"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/middleware"

	"gorm.io/gorm"
)

func (h Handler) CreateUser(c *fiber.Ctx) error {
	userCreatedEvent := UserCreatedEvent{}
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

	err = h.DB.CreateUser(&user)
	if err != nil {
		return err
	}

	return c.SendString("ok")
}

func (h Handler) GetUser(c *fiber.Ctx) error {
	userId := c.Params("userId")

	sessionClaims := c.Locals(middleware.SESSION_CLAIMS)
	details, ok := sessionClaims.(*clerk.SessionClaims)
	if !ok {
		return errors.New("token error")
	}
	if details.Subject != userId {
		return fiber.NewError(fiber.StatusForbidden)
	}

	user, err := h.DB.GetUser(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user does not exist")
		}
		return err
	}

	userResponse := User{ID: user.ID, Email: user.Email}

	return c.JSON(userResponse)
}
