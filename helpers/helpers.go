package helpers

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neilZon/workout-logger-api/accesscontroller"
	"github.com/neilZon/workout-logger-api/common"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/token"
	"github.com/vektah/gqlparser/v2/gqlerror"
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

func NewGqlServer(gormDB *gorm.DB, acs accesscontroller.AccessControllerService) *handler.Server {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		DB:  gormDB,
		ACS: acs,
	}}))

	srv.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)
		// add status code for unauthorized errors so client knows to refresh token
		var unauthorizedError *common.UnauthorizedError
		if errors.As(e, &unauthorizedError) {
			err.Extensions = map[string]interface{}{
				"code": "UNAUTHORIZED",
			}
		}
		return err
	})
	return srv
}

func NewGqlClient(gormDB *gorm.DB, acs accesscontroller.AccessControllerService) *client.Client {
	srv := NewGqlServer(gormDB, acs)
	return client.New(srv)
}

func AddContext(u *token.Claims) client.Option {
	return func(bd *client.Request) {
		ctx := context.WithValue(bd.HTTP.Context(), middleware.UserCtxKey, u)
		bd.HTTP = bd.HTTP.WithContext(ctx)
	}
}

func StringToUInt(s string) uint {
	num, err := strconv.ParseUint(s, 10, strconv.IntSize)
	if err != nil {
		panic(err)
	}
	return uint(num)
}

func UIntToString(num uint) string {
	return fmt.Sprintf("%d", num)
}
