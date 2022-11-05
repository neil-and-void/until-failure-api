package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/accesscontrol"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
)

type AddSetEntryResp struct {
	AddSet string
}

type GetSetEntriesResp struct {
	Sets []struct {
		ID     string
		Weight float32
		Reps   int
	}
}

func TestSetEntryResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	u := User
	e := WorkoutSession.Exercises[0]
	ws := WorkoutSession
	s := WorkoutSession.Exercises[0].Sets[0]

	t.Run("Add Set Entry Success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		exerciseRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExercisesQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExercisesQuery)).
			WithArgs(e.ID).
			WillReturnRows(exerciseRow)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"})
		for _, s := range e.Sets {
			setEntryRows.AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		}
		const getSetEntries = `SELECT * FROM "set_entries" WHERE "set_entries"."exercise_id" = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntries)).
			WithArgs(e.ID).
			WillReturnRows(setEntryRows)
	
		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()
		addSetEntriesQuery := `INSERT INTO "set_entries" ("created_at","updated_at","deleted_at","weight","reps","exercise_id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addSetEntriesQuery)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), s.Weight, s.Reps, s.ExerciseID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(s.ID))
		mock.ExpectCommit()
	
		var resp AddSetEntryResp
		c.MustPost(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 225.0, reps: 8 })
			}
			`,
			&resp,
			AddContext(u),
		)

		err := mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
	
	t.Run("Get Set Entries Success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		exerciseRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExercisesQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExercisesQuery)).
			WithArgs(e.ID).
			WillReturnRows(exerciseRow)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"})
		for _, s := range e.Sets {
			setEntryRows.AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		}
		const getSetEntries = `SELECT * FROM "set_entries" WHERE "set_entries"."exercise_id" = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntries)).
			WithArgs(e.ID).
			WillReturnRows(setEntryRows)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		var resp GetSetEntriesResp
		c.MustPost(`
			query GetSets {
				sets(exerciseId: "44") {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			AddContext(u),
		)

		err := mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}