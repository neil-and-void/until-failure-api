package helpers

import (
	"context"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neilZon/workout-logger-api/accesscontroller"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/token"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const WorkoutRoutineAccessQuery = `SELECT * FROM "workout_routines" WHERE (user_id = $1 AND id = $2) AND "workout_routines"."deleted_at" IS NULL ORDER BY "workout_routines"."id" LIMIT 1`
const WorkoutSessionAccessQuery = `SELECT * FROM "workout_sessions" WHERE (user_id = $1 AND id = $2) AND "workout_sessions"."deleted_at" IS NULL ORDER BY "workout_sessions"."id" LIMIT 1`

func SetupMockDB() (sqlmock.Sqlmock, *gorm.DB) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDb,
	}), &gorm.Config{})

	return mock, gormDB
}

func NewGqlClient(gormDB *gorm.DB, acs accesscontroller.AccessControllerService) (*client.Client) {
	return client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		DB: gormDB,
		ACS: acs,
	}})))
}

func AddContext(u *token.Claims) client.Option {
	return func(bd *client.Request) {
		ctx := context.WithValue(bd.HTTP.Context(), middleware.UserCtxKey, u)
		bd.HTTP = bd.HTTP.WithContext(ctx)
	}
}
