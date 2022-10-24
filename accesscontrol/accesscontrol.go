package accesscontrol

import (
	"errors"

	"github.com/neilZon/workout-logger-api/database"
	"gorm.io/gorm"
)

type AccessController struct {
	DB *gorm.DB
}

func (ac *AccessController) CanAccessWorkoutRoutine(userId string, workoutRoutineId string) error {
	_, err := database.GetWorkoutRoutine(ac.DB, userId, workoutRoutineId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("Access Denied")
	}
	if err != nil {
		return err
	}

	return nil
}

func NewAccessController(db *gorm.DB) *AccessController {
	return &AccessController{
		DB: db,
	}
}
