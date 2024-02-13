package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/neilZon/workout-logger-api/middleware"
)

func RegisterRoutes(app *fiber.App, h Handler, m middleware.Middleware) {
	api := app.Group("/api")

	api.Post("/users", h.CreateUser)

	if os.Getenv("ENVIRONMENT") == "PROD" {
		api.Use(m.JWTAuthMiddleware)
	}
	api.Use(m.MockAuthMiddleware)

	api.Get("/users/:userId", h.GetUser)
	api.Get("/users/:userId/routines", h.GetRoutines)
	api.Post("/routines", h.CreateRoutine)
	api.Get("/routines/:routineId", h.GetRoutine)

	api.Post("/exerciseRoutines", h.CreateExerciseRoutine)

	api.Post("/setSchemes", h.CreateSetScheme)
}
