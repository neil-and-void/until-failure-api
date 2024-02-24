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

	routineCount, err := h.DB.GetRoutineCount(details.Subject)
	if err != nil {
		return err
	}

	if routineCount >= 100 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "You cannot have more than 100 routines",
		})
	}

	routine := database.Routine{
		Name:   newRoutine.Name,
		UserID: newRoutine.UserID,
		Active: true,
	}

	err = h.DB.CreateRoutine(&routine)
	if err != nil {
		return err
	}

	createdAt := routine.CreatedAt.Format(utils.ISO8601Format)

	return c.JSON(Routine{
		ID:        routine.ID.String(),
		Name:      routine.Name,
		Active:    routine.Active,
		UserID:    routine.UserID,
		CreatedAt: createdAt,
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

	routines, err := h.DB.GetRoutines(userId)
	if err != nil {
		return err
	}

	routinesResponse := []Routine{}
	for _, routine := range routines {
		exerciseRoutine := []ExerciseRoutine{}

		routinesResponse = append(routinesResponse, Routine{
			ID:               routine.ID.String(),
			Name:             routine.Name,
			Active:           routine.Active,
			UserID:           routine.UserID,
			Private:          routine.Private,
			ExerciseRoutines: exerciseRoutine,
			CreatedAt:        routine.CreatedAt.Format(utils.ISO8601Format),
		})
	}

	return c.JSON(routinesResponse)
}
func (h Handler) GetRoutine(c *fiber.Ctx) error {
	routineId := c.Params("routineId")

	sessionClaims := c.Locals(middleware.SESSION_CLAIMS)
	details, ok := sessionClaims.(*clerk.SessionClaims)
	if !ok {
		return errors.New("token error")
	}

	routine, err := h.DB.GetRoutine(routineId)
	if err != nil {
		return err
	}

	if details.Subject != routine.UserID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	exerciseRoutines := []ExerciseRoutine{}
	for _, exerciseRoutine := range routine.ExerciseRoutines {

		setSchemes := []SetScheme{}
		for _, setScheme := range exerciseRoutine.SetSchemes {
			setSchemes = append(setSchemes, SetScheme{
				ID:                setScheme.ID.String(),
				TargetReps:        setScheme.TargetReps,
				SetType:           SetType(setScheme.SetType),
				Measurement:       MeasurementType(setScheme.Measurement),
				ExerciseRoutineId: setScheme.ExerciseRoutineId.String(),
				CreatedAt:         setScheme.CreatedAt.Format(utils.ISO8601Format),
			})
		}

		exerciseRoutines = append(exerciseRoutines, ExerciseRoutine{
			ID:         exerciseRoutine.ID.String(),
			Name:       exerciseRoutine.Name,
			Active:     exerciseRoutine.Active,
			RoutineId:  exerciseRoutine.RoutineID.String(),
			CreatedAt:  exerciseRoutine.CreatedAt.Format(utils.ISO8601Format),
			SetSchemes: setSchemes,
		})
	}

	return c.JSON(Routine{
		ID:               routine.ID.String(),
		Name:             routine.Name,
		Active:           routine.Active,
		ExerciseRoutines: exerciseRoutines,
		Private:          routine.Private,
		UserID:           routine.UserID,
		CreatedAt:        routine.CreatedAt.Format(utils.ISO8601Format),
	})
}
