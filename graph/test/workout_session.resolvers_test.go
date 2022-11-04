package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type AddWorkoutSessionResp struct {
	AddWorkoutSession string
}

type GetWorkoutSession struct {
	WorkoutSessions []struct {
		ID        string
		Start     string
		End       string
		Exercises []struct {
			ID   string
			Sets []struct {
				ID     string
				Weight float32
				Reps   int
				Notes  string
			}
		}
	}
}

func TestWorkoutSessionResolvers(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	ws := WorkoutSession
	u := User

	t.Run("Add Workout Session success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		mock.ExpectBegin()

		const addWorkoutSessionStmnt = `INSERT INTO "workout_sessions" ("created_at","updated_at","deleted_at","start","end","workout_routine_id","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addWorkoutSessionStmnt)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ws.Start, ws.End, ws.WorkoutRoutineID, ws.UserID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ws.ID))

		const addExerciseStmt = `INSERT INTO "exercises" ("created_at","updated_at","deleted_at","workout_session_id","exercise_routine_id","notes") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) ON CONFLICT ("id") DO UPDATE SET "workout_session_id"="excluded"."workout_session_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addExerciseStmt)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[0].WorkoutSessionID,
			ws.Exercises[0].ExerciseRoutineID,
			ws.Exercises[0].Notes,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[1].WorkoutSessionID,
			ws.Exercises[1].ExerciseRoutineID,
			ws.Exercises[1].Notes,
		).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ws.Exercises[0].ID).AddRow(ws.Exercises[1].ID))

		const addSetEntries = `INSERT INTO "set_entries" ("created_at","updated_at","deleted_at","weight","reps","exercise_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12),($13,$14,$15,$16,$17,$18),($19,$20,$21,$22,$23,$24) ON CONFLICT ("id") DO UPDATE SET "exercise_id"="excluded"."exercise_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addSetEntries)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[0].Sets[0].Weight,
			ws.Exercises[0].Sets[0].Reps,
			ws.Exercises[0].ID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[0].Sets[1].Weight,
			ws.Exercises[0].Sets[1].Reps,
			ws.Exercises[0].ID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[1].Sets[0].Weight,
			ws.Exercises[1].Sets[0].Reps,
			ws.Exercises[1].ID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[1].Sets[1].Weight,
			ws.Exercises[1].Sets[1].Reps,
			ws.Exercises[1].ID,
		).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ws.Exercises[0].Sets[0].ID).AddRow(ws.Exercises[0].Sets[1].ID).AddRow(ws.Exercises[1].Sets[0].ID))

		mock.ExpectCommit()

		var resp AddWorkoutSessionResp
		c.MustPost(`
			mutation AddWorkoutSession {
				addWorkoutSession(workout: {
					start: "2022-10-30T12:34:00Z",
					workoutRoutineId: "8",
					exercises: [
						{
							exerciseRoutineId: "3", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is a note"
						},
						{
							exerciseRoutineId: "4", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is another note"
						}
					],
				}) 
			}`,
			&resp,
			AddContext(u),
		)

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Workout Session Access Invalid Token", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		var resp AddWorkoutSessionResp
		err := c.Post(`
			mutation AddWorkoutSession {
				addWorkoutSession(workout: {
					start: "2022-10-30T12:34:00Z",
					workoutRoutineId: "8",
					exercises: [
						{
							exerciseRoutineId: "3", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is a note"
						},
						{
							exerciseRoutineId: "4", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is another note"
						}
					],
				}) 
			}`,
			&resp,
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Workout Session: Invalid Token\",\"path\":[\"addWorkoutSession\"]}]")
	})

	t.Run("Add Workout Session Error (invalid workout routine ID fk constraint)", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		mock.ExpectBegin()

		const addWorkoutSessionStmnt = `INSERT INTO "workout_sessions" ("created_at","updated_at","deleted_at","start","end","workout_routine_id","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addWorkoutSessionStmnt)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ws.Start, ws.End, 8789, ws.UserID).
			WillReturnError(gorm.ErrInvalidValue)

		mock.ExpectRollback()

		var resp AddWorkoutSessionResp
		err := c.Post(`
			mutation AddWorkoutSession {
				addWorkoutSession(workout: {
					start: "2022-10-30T12:34:00Z",
					workoutRoutineId: "8789",
					exercises: [
						{
							exerciseRoutineId: "3", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is a note"
						},
						{
							exerciseRoutineId: "4", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is another note"
						}
					],
				}) 
			}`,
			&resp,
			AddContext(u),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Workout Session\",\"path\":[\"addWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Workout Session Error (invalid exercise ID fk constraint)", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		mock.ExpectBegin()

		const addWorkoutSessionStmnt = `INSERT INTO "workout_sessions" ("created_at","updated_at","deleted_at","start","end","workout_routine_id","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addWorkoutSessionStmnt)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ws.Start, ws.End, ws.WorkoutRoutineID, ws.UserID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ws.ID))

		const addExerciseStmt = `INSERT INTO "exercises" ("created_at","updated_at","deleted_at","workout_session_id","exercise_routine_id","notes") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) ON CONFLICT ("id") DO UPDATE SET "workout_session_id"="excluded"."workout_session_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addExerciseStmt)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[0].WorkoutSessionID,
			ws.Exercises[0].ExerciseRoutineID,
			ws.Exercises[0].Notes,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[1].WorkoutSessionID,
			9879,
			ws.Exercises[1].Notes,
		).WillReturnError(gorm.ErrInvalidValue)

		mock.ExpectRollback()

		var resp AddWorkoutSessionResp
		err := c.Post(`
			mutation AddWorkoutSession {
				addWorkoutSession(workout: {
					start: "2022-10-30T12:34:00Z",
					workoutRoutineId: "8",
					exercises: [
						{
							exerciseRoutineId: "3", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is a note"
						},
						{
							exerciseRoutineId: "9879", 
							setEntries: [
								{ weight: 225, reps: 8},
								{ weight: 225, reps: 7},
							],
							notes: "This is another note"
						}
					],
				}) 
			}`,
			&resp,
			AddContext(u),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Workout Session\",\"path\":[\"addWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Workout Sessions success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "start", "end", "workout_routine_id", "user_id"}).
			AddRow(ws.ID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt, ws.Start, ws.End, ws.WorkoutRoutineID, ws.UserID)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"})
		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"})
		for _, e := range ws.Exercises {
			exerciseRows.AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)

			for _, s := range e.Sets {
				setEntryRows.AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
			}
		}

		const getWorkoutSessions = `SELECT * FROM "workout_sessions" WHERE user_id = $1 AND "workout_sessions"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getWorkoutSessions)).
			WithArgs(fmt.Sprintf("%d", u.ID)).
			WillReturnRows(workoutSessionRow)

		const getExercises = `SELECT * FROM "exercises" WHERE "exercises"."workout_session_id" = $1 AND "exercises"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getExercises)).
			WithArgs(ws.ID).
			WillReturnRows(exerciseRows)

		const getSetEntries = `SELECT * FROM "set_entries" WHERE "set_entries"."exercise_id" IN ($1,$2) AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntries)).
			WithArgs(ws.Exercises[0].ID, ws.Exercises[1].ID).
			WillReturnRows(setEntryRows)

		var resp GetWorkoutSession
		c.MustPost(`
			query WorkoutSessions {
				workoutSessions {
					id
					start
					exercises {
						sets {
							weight
							reps
						}
					}
				}
			}`,
			&resp,
			AddContext(u),
		)

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Workout Sessions Invalid Token", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		var resp GetWorkoutSession
		err := c.Post(`
			query WorkoutSessions {
				workoutSessions {
					id
					start
					exercises {
						sets {
							weight
							reps
						}
					}
				}
			}`,
			&resp,
		)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Workout Sessions: Invalid Token\",\"path\":[\"workoutSessions\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Workout Session Success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "start", "end", "workout_routine_id", "user_id"}).
			AddRow(ws.ID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt, ws.Start, ws.End, ws.WorkoutRoutineID, ws.UserID)

		const getWorkoutSession = `SELECT * FROM "workout_sessions" WHERE (user_id = $1 AND id = $2) AND "workout_sessions"."deleted_at" IS NULL ORDER BY "workout_sessions"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getWorkoutSession)).
			WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).
			WillReturnRows(workoutSessionRow)

		var resp GetWorkoutSession
		err := c.Post(`
			query WorkoutSession {
				workoutSession(workoutSessionId: "3") {
					id
					start
					exercises {
						sets {
							weight
							reps
						}
					}
				}
			}`,
			&resp,
			AddContext(u),
		)

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Workout Session Invalid Token", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		var resp GetWorkoutSession
		err := c.Post(`
			query WorkoutSession {
				workoutSession(workoutSessionId: "3") {
					id
					start
					exercises {
						sets {
							weight
							reps
						}
					}
				}
			}`,
			&resp,
		)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Workout Sessions: Invalid Token\",\"path\":[\"workoutSession\"]}]")
	})
}
