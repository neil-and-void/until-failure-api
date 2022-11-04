package test

import (
	"context"

	"github.com/99designs/gqlgen/client"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/token"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const WorkoutRoutineAccessQuery = `SELECT * FROM "workout_routines" WHERE (user_id = $1 AND id = $2) AND "workout_routines"."deleted_at" IS NULL ORDER BY "workout_routines"."id" LIMIT 1`
const WorkoutSessionAccessQuery = `SELECT * FROM "workout_sessions" WHERE (user_id = $1 AND id = $2) AND "workout_sessions"."deleted_at" IS NULL ORDER BY "workout_sessions"."id" LIMIT 1`

func SetupMockDB() (sqlmock.Sqlmock, *gorm.DB) {
	mockDb, mock, err := sqlmock.New() // mock sql.DB
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDb,
	}), &gorm.Config{})

	return mock, gormDB
}

func AddContext(u *token.Claims) client.Option {
	return func(bd *client.Request) {
		ctx := context.WithValue(bd.HTTP.Context(), middleware.UserCtxKey, u)
		bd.HTTP = bd.HTTP.WithContext(ctx)
	}
}
