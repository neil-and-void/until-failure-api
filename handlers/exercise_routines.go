package handlers

import (
	"errors"

	"github.com/google/uuid"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gofiber/fiber/v2"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/utils"
)

func (h Handler) CreateExerciseRoutine(c *fiber.Ctx) error {
	newExerciseRoutine := NewExerciseRoutine{}
	if err := c.BodyParser(&newExerciseRoutine); err != nil {
		return err
	}

	sessionClaims := c.Locals(middleware.SESSION_CLAIMS)
	details, ok := sessionClaims.(*clerk.SessionClaims)
	if !ok {
		return errors.New("token error")
	}

	routine, err := h.DB.GetRoutine(newExerciseRoutine.RoutineID)
	if err != nil {
		return err
	}

	if routine.UserID != details.Subject {
		return fiber.NewError(fiber.StatusForbidden)
	}

	parsedUUID, err := uuid.Parse(newExerciseRoutine.RoutineID)
	if err != nil {
		return err
	}

	exerciseRoutine := database.ExerciseRoutine{
		Name:      newExerciseRoutine.Name,
		RoutineID: parsedUUID,
	}
	err = h.DB.CreateExerciseRoutine(&exerciseRoutine)
	if err != nil {
		return err
	}

	return c.JSON(ExerciseRoutine{
		ID:        exerciseRoutine.ID.String(),
		Name:      exerciseRoutine.Name,
		Active:    exerciseRoutine.Active,
		RoutineId: exerciseRoutine.RoutineID.String(),
		CreatedAt: exerciseRoutine.CreatedAt.Format(utils.ISO8601Format),
	})
}

func (h Handler) UpdateExerciseRoutine(c *fiber.Ctx) error {
	updatedExerciseRoutine := UpdateExerciseRoutine{}
	if err := c.BodyParser(&updatedExerciseRoutine); err != nil {
		return err
	}

	if err := h.Validate.Struct(updatedExerciseRoutine); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	exerciseRoutineId := c.Params("exerciseRoutineId")
	parsedUUID, err := uuid.Parse(exerciseRoutineId)
	if err != nil {
		return err
	}

	exerciseRoutine := database.ExerciseRoutine{
		ID:   parsedUUID,
		Name: updatedExerciseRoutine.Name,
	}
	if err := h.DB.UpdateExericseRoutine(&exerciseRoutine); err != nil {
		return err
	}

	return c.JSON(ExerciseRoutine{
		ID:        exerciseRoutine.ID.String(),
		Name:      exerciseRoutine.Name,
		Active:    exerciseRoutine.Active,
		RoutineId: exerciseRoutine.RoutineID.String(),
		CreatedAt: exerciseRoutine.CreatedAt.Format(utils.ISO8601Format),
	})
}
