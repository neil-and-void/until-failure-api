package handlers

import (
	"errors"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gofiber/fiber/v2"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/utils"
)

func (h Handler) CreateRoutine(c *fiber.Ctx) error {
	newRoutine := NewRoutine{}
	if err := c.BodyParser(&newRoutine); err != nil {
		return err
	}

	sessionClaims := c.Locals(middleware.SESSION_CLAIMS)
	details, ok := sessionClaims.(*clerk.SessionClaims)
	if !ok {
		return errors.New("token error")
	}
	if details.Subject != newRoutine.UserID {
		return fiber.NewError(fiber.StatusForbidden)
	}

	routine := database.Routine{
		Name:   newRoutine.Name,
		UserID: newRoutine.UserID,
		Active: true,
	}

	err := h.DB.CreateRoutine(&routine)
	if err != nil {
		return err
	}

	createdAt := routine.CreatedAt.Format(utils.ISO8601Format)

	return c.JSON(Routine{
		ID:               routine.ID.String(),
		Name:             routine.Name,
		ExerciseRoutines: []ExerciseRoutine{},
		Active:           routine.Active,
		UserID:           routine.UserID,
		CreatedAt:        createdAt,
	})
}

func (h Handler) GetRoutines(c *fiber.Ctx) error {
	userId := c.Params("userId")

	sessionClaims := c.Locals(middleware.SESSION_CLAIMS)
	details, ok := sessionClaims.(*clerk.SessionClaims)
	if !ok {
		return errors.New("token error")
	}
	if details.Subject != userId {
		return fiber.NewError(fiber.StatusForbidden)
	}

	return c.SendString("I'm a GET request!")
}

func (h Handler) UpdateRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a PATCH request!")
}

func (h Handler) DeleteRoutine(c *fiber.Ctx) error {
	return c.SendString("I'm a GET request!")
}
