package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/accesscontroller/accesscontrol"
	"github.com/neilZon/workout-logger-api/helpers"
	"github.com/neilZon/workout-logger-api/tests/testdata"
	"github.com/neilZon/workout-logger-api/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type AddExerciseRoutine struct {
	AddExerciseRoutine string
}

type GetExerciseRoutineResp struct {
	ExerciseRoutines []struct {
		ID   string
		Name string
		Sets int
		Reps int
	}
}

type UpdateExerciseRoutineResp struct {
	UpdateExerciseRoutine struct {
		ID   string
		Name string
		Sets int
		Reps int
	}
}

type DeleteExerciseRoutineResp struct {
	DeleteExerciseRoutine int
}

func TestExerciseRoutineResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	u := testdata.User
	wr := testdata.WorkoutRoutine
	ws := testdata.WorkoutSession
	er := testdata.WorkoutRoutine.ExerciseRoutines[0]

	t.Run("Add Exercise Routine", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		mock.ExpectBegin()
		createExerciseRoutineStmt := `INSERT INTO "exercise_routines" ("created_at","updated_at","deleted_at","name","sets","reps","active","workout_routine_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createExerciseRoutineStmt)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), er.Name, er.Sets, er.Reps, er.Active, er.WorkoutRoutineID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(er.ID))
		mock.ExpectCommit()

		var resp AddExerciseRoutine
		mutation := fmt.Sprintf(`
			mutation AddExerciseRoutine {
				addExerciseRoutine(workoutRoutineId: "%d", exerciseRoutine: {
					sets: %d,
					reps: %d,
					name: "%s"
				}) 
			}
			`,
			er.WorkoutRoutineID, er.Sets, er.Reps, er.Name,
		)
		c.MustPost(mutation, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Exercise Routine Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp AddExerciseRoutine
		mutation := fmt.Sprintf(`
			mutation AddExerciseRoutine {
				addExerciseRoutine(workoutRoutineId: "%d", exerciseRoutine: {
					sets: %d,
					reps: %d,
					name: "%s"
				}) 
			}
			`,
			er.WorkoutRoutineID, er.Sets, er.Reps, er.Name,
		)
		err := c.Post(mutation, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"addExerciseRoutine\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Exercise Routine Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnError(gorm.ErrRecordNotFound)

		var resp AddExerciseRoutine
		mutation := fmt.Sprintf(`
			mutation AddExerciseRoutine {
				addExerciseRoutine(workoutRoutineId: "%d", exerciseRoutine: {
					sets: %d,
					reps: %d,
					name: "%s"
				}) 
			}
			`,
			er.WorkoutRoutineID, er.Sets, er.Reps, er.Name,
		)
		err := c.Post(mutation, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Adding Exercise Routine: Access Denied\",\"path\":[\"addExerciseRoutine\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Exercise Routine Error Creating", func(t *testing.T) {})

	t.Run("Get Exercise Routines Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		exerciseRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "sets", "reps", "created_at", "deleted_at", "updated_at"}).
			AddRow(er.ID, er.Name, er.Sets, er.Reps, er.CreatedAt, er.DeletedAt, er.UpdatedAt)
		const exerciseRoutineQuery = `SELECT * FROM "exercise_routines" WHERE workout_routine_id = $1 AND "exercise_routines"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(exerciseRoutineQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(exerciseRoutineRow)

		var resp GetExerciseRoutineResp
		query := fmt.Sprintf(`
			query ExerciseRoutines {
				exerciseRoutines(workoutRoutineId: "%d") {
					id
					name
				}
			}`,
			er.WorkoutRoutineID,
		)
		c.MustPost(query, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Exercise Routines Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		incorrectUserId := 66
		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, incorrectUserId, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		var resp GetExerciseRoutineResp
		query := fmt.Sprintf(`query ExerciseRoutines {
			exerciseRoutines(workoutRoutineId: "%d") {
				id
				name
			}
		}`, er.WorkoutRoutineID)
		err = c.Post(query, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Exercise Routines Error", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		const exerciseRoutineQuery = `SELECT * FROM "exercise_routines" WHERE workout_routine_id = $1 AND "exercise_routines"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(exerciseRoutineQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnError(gorm.ErrInvalidTransaction)

		var resp GetExerciseRoutineResp
		query := fmt.Sprintf(`
			query ExerciseRoutines {
				exerciseRoutines(workoutRoutineId: "%d") {
					id
					name
				}
			}`,
			er.WorkoutRoutineID,
		)
		err := c.Post(query, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Getting Exercise Routine\",\"path\":[\"exerciseRoutines\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Routine Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		exerciseRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "sets", "reps", "created_at", "deleted_at", "updated_at", "workout_routine_id"}).
			AddRow(er.ID, er.Name, er.Sets, er.Reps, er.CreatedAt, er.DeletedAt, er.UpdatedAt, er.WorkoutRoutineID)
		const exerciseRoutineQuery = `SELECT * FROM "exercise_routines" WHERE id = $1 AND "exercise_routines"."deleted_at" IS NULL ORDER BY "exercise_routines"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(exerciseRoutineQuery)).WithArgs(fmt.Sprintf("%d", er.ID)).WillReturnRows(exerciseRoutineRow)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		mock.ExpectBegin()
		deleteExerciseRoutineQuery := `UPDATE "exercise_routines" SET "deleted_at"=$1 WHERE id = $2 AND "exercise_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteExerciseRoutineQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(er.ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		exerciseRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"})
		for _, e := range ws.Exercises {
			exerciseRow.AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		}
		deleteExerciseQuery := `UPDATE "exercises" SET "deleted_at"=$1 WHERE exercise_routine_id = $2 AND "exercises"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(deleteExerciseQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(er.ID)).
			WillReturnRows(exerciseRow)

		deleteSetQuery := `UPDATE "set_entries" SET "deleted_at"=$1 WHERE exercise_id IN ($2,$3) AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteSetQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(ws.Exercises[0].ID), utils.UIntToString(ws.Exercises[1].ID)).
			WillReturnResult(sqlmock.NewResult(1, 2))
		mock.ExpectCommit()

		var resp DeleteExerciseRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExerciseRoutine {
				deleteExerciseRoutine(exerciseRoutineId: "%d")
			}`,
			er.ID,
		)
		c.MustPost(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Routine Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp DeleteExerciseRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExerciseRoutine {
				deleteExerciseRoutine(exerciseRoutineId: "%d")
			}`,
			er.ID,
		)
		err := c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"deleteExerciseRoutine\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Routine Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		exerciseRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "sets", "reps", "created_at", "deleted_at", "updated_at", "workout_routine_id"}).
			AddRow(er.ID, er.Name, er.Sets, er.Reps, er.CreatedAt, er.DeletedAt, er.UpdatedAt, er.WorkoutRoutineID)
		const exerciseRoutineQuery = `SELECT * FROM "exercise_routines" WHERE id = $1 AND "exercise_routines"."deleted_at" IS NULL ORDER BY "exercise_routines"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(exerciseRoutineQuery)).WithArgs(fmt.Sprintf("%d", er.ID)).WillReturnRows(exerciseRoutineRow)

		incorrectUserId := 66
		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, incorrectUserId, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		var resp DeleteExerciseRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExerciseRoutine {
				deleteExerciseRoutine(exerciseRoutineId: "%d")
			}`,
			er.ID,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Exercise Routine: Access Denied\",\"path\":[\"deleteExerciseRoutine\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Routine Error", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		exerciseRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "sets", "reps", "created_at", "deleted_at", "updated_at", "workout_routine_id"}).
			AddRow(er.ID, er.Name, er.Sets, er.Reps, er.CreatedAt, er.DeletedAt, er.UpdatedAt, er.WorkoutRoutineID)
		const exerciseRoutineQuery = `SELECT * FROM "exercise_routines" WHERE id = $1 AND "exercise_routines"."deleted_at" IS NULL ORDER BY "exercise_routines"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(exerciseRoutineQuery)).WithArgs(fmt.Sprintf("%d", er.ID)).WillReturnRows(exerciseRoutineRow)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		mock.ExpectBegin()
		deleteExerciseRoutineQuery := `UPDATE "exercise_routines" SET "deleted_at"=$1 WHERE id = $2 AND "exercise_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteExerciseRoutineQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(er.ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		exerciseRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"})
		for _, e := range ws.Exercises {
			exerciseRow.AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		}
		deleteExerciseQuery := `UPDATE "exercises" SET "deleted_at"=$1 WHERE exercise_routine_id = $2 AND "exercises"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(deleteExerciseQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(er.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		var resp DeleteExerciseRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExerciseRoutine {
				deleteExerciseRoutine(exerciseRoutineId: "%d")
			}`,
			er.ID,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Exercise Routine\",\"path\":[\"deleteExerciseRoutine\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}
