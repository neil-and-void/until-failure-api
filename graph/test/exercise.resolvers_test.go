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
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type AddExerciseResp struct {
	AddExercise string
}

type GetExercisesResp struct {
	Exercises []struct {
		ID   string
		Sets []struct {
			ID     string
			Weight float32
			Reps   int
		}
		Notes  string
	}
}

type GetExerciseResp struct {
	Exercise struct {
		ID   string
		Sets []struct {
			ID     string
			Weight float32
			Reps   int
		}
		Notes  string
	}
}

func TestExerciseResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	u := User
	ws := WorkoutSession
	e := WorkoutSession.Exercises[0]

	t.Run("Add Exercise Success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()

		const createExerciseStmnt = `INSERT INTO "exercises" ("created_at","updated_at","deleted_at","workout_session_id","exercise_routine_id","notes") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createExerciseStmnt)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), e.WorkoutSessionID, e.ExerciseRoutineID, e.Notes).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(e.ID))

		const creatSetStmnt = `INSERT INTO "set_entries" ("created_at","updated_at","deleted_at","weight","reps","exercise_id") VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT ("id") DO UPDATE SET "exercise_id"="excluded"."exercise_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(creatSetStmnt)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			e.Sets[0].Weight,
			e.Sets[0].Reps,
			e.Sets[0].ExerciseID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(e.Sets[0].ID))

		mock.ExpectCommit()

		var resp AddExerciseResp
		c.MustPost(`
			mutation AddExercise {
				addExercise(
					exercise: {
						exerciseRoutineId: "3"
						setEntries: [{ weight: 225, reps: 8 }]
						notes: "This is a note"
					}
					workoutSessionId: "3",
				)
			}`,
			&resp,
			AddContext(u),
		)

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	// todo
	t.Run("Add Exercise Foreign Key Error", func(t *testing.T) {})

	t.Run("Add Exercise Invalid Token", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		var resp AddExerciseResp
		err = c.Post(`
			mutation AddExercise {
				addExercise(
					exercise: {
						exerciseRoutineId: "3"
						setEntries: [{ weight: 225, reps: 8 }],
						notes: "This is a note"
					}
					workoutSessionId: "3",
				)
			}`,
			&resp,
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Exercise: Invalid Token\",\"path\":[\"addExercise\"]}]")
	})

	t.Run("Add Exercise Access Denied", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		workoutSessionId := 1233
		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", 1233)).WillReturnError(gorm.ErrRecordNotFound)

		var resp AddExerciseResp
		gqlMutation := fmt.Sprintf(`
			mutation AddExercise {
				addExercise(
					exercise: {
						exerciseRoutineId: "3"
						setEntries: [{ weight: 225, reps: 8 }]
						notes: "This is a note"
					}
					workoutSessionId: "%d",
				)
			}`,
			workoutSessionId,
		)
		err = c.Post(gqlMutation, &resp, AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Adding Exercise: Access Denied\",\"path\":[\"addExercise\"]}]")
	})

	t.Run("Get Exercises Success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"})
		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"})
		for _, e := range ws.Exercises {
			exerciseRows.AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)

			for _, s := range e.Sets {
				setEntryRows.AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
			}
		}

		const getExercisesQuery = `SELECT * FROM "exercises" WHERE workout_session_id = $1 AND "exercises"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getExercisesQuery)).
			WithArgs(fmt.Sprintf("%d", ws.ID)).
			WillReturnRows(exerciseRows)

		const getSetEntries = `SELECT * FROM "set_entries" WHERE "set_entries"."exercise_id" IN ($1,$2) AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntries)).
			WithArgs(ws.Exercises[0].ID, ws.Exercises[1].ID).
			WillReturnRows(setEntryRows)

		var resp GetExercisesResp
		gqlQuery := fmt.Sprintf(`	
			query Exercises {
				exercises(workoutSessionId: "%d") {
					id
					sets {
						weight
						reps
					}
					notes
				}
			}`,
			ws.ID,
		)
		c.MustPost(gqlQuery, &resp, AddContext(u))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Exercises Invalid Token", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		var resp GetExerciseResp
		gqlQuery := fmt.Sprintf(`	
			query Exercises {
				exercises(workoutSessionId: "%d") {
					id
					sets {
						weight
						reps
					}
					notes
				}
			}`,
			ws.ID,
		)
		err = c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Exercises: Invalid Token\",\"path\":[\"exercises\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Exercises Access Denied", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnError(gorm.ErrRecordNotFound)

		var resp GetExerciseResp
		gqlQuery := fmt.Sprintf(`	
			query Exercises {
				exercises(workoutSessionId: "%d") {
					id
					sets {
						weight
						reps
					}
					notes
				}
			}`,
			ws.ID,
		)
		err = c.Post(gqlQuery, &resp, AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Getting Exercises: Access Denied\",\"path\":[\"exercises\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Exercise Success", func(t *testing.T) {
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
			WithArgs(ws.Exercises[0].ID).
			WillReturnRows(setEntryRows)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		var resp GetExerciseResp
		gqlQuery := fmt.Sprintf(`	
			query Exercise {
				exercise(exerciseId: "%d") {
					id
					sets {
						weight
						reps
					}
					notes
				}
			}`,
			e.ID,
		)
		c.MustPost(gqlQuery, &resp, AddContext(u))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Exercise Invalid Token", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		var resp GetExerciseResp
		gqlQuery := fmt.Sprintf(`	
			query Exercise {
				exercise(exerciseId: "%d") {
					id
					sets {
						weight
						reps
					}
					notes
				}
			}`,
			e.ID,
		)
		err := c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Exercise: Invalid Token\",\"path\":[\"exercise\"]}]")
	})

	t.Run("Get Exercise Access Denied", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
			AC: &accesscontrol.AccessController{DB: gormDB},
		}})))

		exerciseId := 788

		exerciseRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExercisesQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExercisesQuery)).
			WithArgs(exerciseId).
			WillReturnRows(exerciseRow)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"})
		for _, s := range e.Sets {
			setEntryRows.AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		}
		const getSetEntries = `SELECT * FROM "set_entries" WHERE "set_entries"."exercise_id" = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntries)).
			WithArgs(ws.Exercises[0].ID).
			WillReturnRows(setEntryRows)

		mock.ExpectQuery(regexp.QuoteMeta(WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnError(gorm.ErrRecordNotFound)

		var resp GetExerciseResp
		gqlQuery := fmt.Sprintf(`	
			query Exercise {
				exercise(exerciseId: "%d") {
					id
					sets {
						weight
						reps
					}
					notes
				}
			}`,
			exerciseId,
		)
		err = c.Post(gqlQuery, &resp, AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Getting Exercise: Access Denied\",\"path\":[\"exercise\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}
