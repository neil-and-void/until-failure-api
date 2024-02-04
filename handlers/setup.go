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
	// api.Put("/routines/:routineId", h.UpdateRoutine)
	// api.Delete("/routines/:routineId", h.UpdateRoutine)

	// api.Get("/users/:userId/workouts", h.GetWorkouts)
	// api.Post("/users/:userId/workouts")
	// api.Put("/users/:userId/workouts")

	// api.Get("/workouts/:workoutId")
	//
	// api.Get("/routines/:routineId")
	//
	// api.Post("/routines/:routineId/exercises")
}
