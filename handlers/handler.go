package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/neilZon/workout-logger-api/database"
)

type Handler struct {
	DB       database.UntilFailureDB
	Validate *validator.Validate
}
