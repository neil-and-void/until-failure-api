package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/utils"
)

func (h Handler) CreateSetScheme(c *fiber.Ctx) error {
	newSetScheme := NewSetScheme{}
	if err := c.BodyParser(&newSetScheme); err != nil {
		return err
	}

	if err := h.Validate.Struct(newSetScheme); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	parsedUUID, err := uuid.Parse(newSetScheme.ExerciseRoutineID)
	if err != nil {
		return err
	}

	setScheme := database.SetScheme{
		TargetReps:        newSetScheme.TargetReps,
		SetType:           database.SetType(newSetScheme.SetType),
		Measurement:       database.MeasurementType(newSetScheme.Measurement),
		ExerciseRoutineId: parsedUUID,
	}
	if err := h.DB.CreateSetScheme(&setScheme); err != nil {
		return err
	}

	setSchemeResponse := SetScheme{
		ID:                setScheme.ID.String(),
		TargetReps:        setScheme.TargetReps,
		SetType:           SetType(setScheme.SetType),
		Measurement:       MeasurementType(setScheme.Measurement),
		ExerciseRoutineId: setScheme.ExerciseRoutineId.String(),
		CreatedAt:         setScheme.CreatedAt.Format(utils.ISO8601Format),
	}

	return c.JSON(setSchemeResponse)
}

func (h Handler) UpdateSetScheme(c *fiber.Ctx) error {
	updatedSetScheme := UpdateSetScheme{}
	if err := c.BodyParser(&updatedSetScheme); err != nil {
		return err
	}

	if err := h.Validate.Struct(updatedSetScheme); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	setSchemeId := c.Params("setSchemeId")
	parsedUUID, err := uuid.Parse(setSchemeId)
	if err != nil {
		return err
	}

	setScheme := database.SetScheme{
		ID:          parsedUUID,
		TargetReps:  updatedSetScheme.TargetReps,
		SetType:     database.SetType(updatedSetScheme.SetType),
		Measurement: database.MeasurementType(updatedSetScheme.Measurement),
	}
	if err := h.DB.UpdateSetScheme(&setScheme); err != nil {
		return err
	}

	setSchemeResponse := SetScheme{
		ID:                setScheme.ID.String(),
		TargetReps:        setScheme.TargetReps,
		SetType:           SetType(setScheme.SetType),
		Measurement:       MeasurementType(setScheme.Measurement),
		ExerciseRoutineId: setScheme.ExerciseRoutineId.String(),
		CreatedAt:         setScheme.CreatedAt.Format(utils.ISO8601Format),
	}

	return c.JSON(setSchemeResponse)
}
