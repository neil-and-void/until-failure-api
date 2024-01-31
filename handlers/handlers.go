package handlers

import (
	"github.com/neilZon/workout-logger-api/database"
)

type Handler struct {
	DB database.UntilFailureDB
}
