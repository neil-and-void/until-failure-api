package helpers

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/graph-gophers/dataloader"
	"github.com/neilZon/workout-logger-api/accesscontroller"
	"github.com/neilZon/workout-logger-api/common"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/loader"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/reader"
	"github.com/neilZon/workout-logger-api/token"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const WorkoutRoutineAccessQuery = `SELECT * FROM "workout_routines" WHERE id = $1 AND "workout_routines"."deleted_at" IS NULL ORDER BY "workout_routines"."id" LIMIT 1`
const WorkoutSessionAccessQuery = `SELECT * FROM "workout_sessions" WHERE id = $1 AND "workout_sessions"."deleted_at" IS NULL ORDER BY "workout_sessions"."id" LIMIT 1`

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

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(gormDB *gorm.DB) *loader.Loaders {
	exerciseRoutineReader := &reader.ExerciseRoutineReader{DB: gormDB}
	setEntrySliceReader := &reader.SetEntrySliceReader{DB: gormDB}
	workoutRoutineReader := &reader.WorkoutRoutineReader{DB: gormDB}
	exerciseRoutineSliceLoader := &reader.ExerciseRoutineSliceReader{DB: gormDB}
	exerciseSliceLoader := &reader.ExerciseSliceReader{DB: gormDB}
	prevExerciseSliceLoader := &reader.PrevExerciseSliceReader{DB: gormDB}

	loaders := &loader.Loaders{
		ExerciseRoutineLoader: dataloader.NewBatchedLoader(exerciseRoutineReader.GetExerciseRoutines),
		SetEntrySliceLoader:        dataloader.NewBatchedLoader(setEntrySliceReader.GetSetEntrySlices),
		WorkoutRoutineLoader:  dataloader.NewBatchedLoader(workoutRoutineReader.GetWorkoutRoutines),
		ExerciseRoutineSliceLoader: dataloader.NewBatchedLoader(exerciseRoutineSliceLoader.GetExerciseRoutineSlices),
		ExerciseSliceLoader: dataloader.NewBatchedLoader(exerciseSliceLoader.GetExerciseSlices),
		PrevExerciseSliceLoader: dataloader.NewBatchedLoader(prevExerciseSliceLoader.GetPrevExerciseSlices),
	}
	return loaders
}

func AddContext(u *token.Claims, l *loader.Loaders) client.Option {
	return func(bd *client.Request) {
		ctx := context.WithValue(bd.HTTP.Context(), middleware.UserCtxKey, u)
		ctx = context.WithValue(ctx, middleware.LoadersKey, l)
		bd.HTTP = bd.HTTP.WithContext(ctx)
	}
}
