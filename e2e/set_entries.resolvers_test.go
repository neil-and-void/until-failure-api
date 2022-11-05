package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/accesscontroller/accesscontrol"
	"github.com/neilZon/workout-logger-api/e2e/testdata"
	"github.com/neilZon/workout-logger-api/helpers"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

	u := testdata.User
	e := testdata.WorkoutSession.Exercises[0]
	ws := testdata.WorkoutSession
	s := testdata.WorkoutSession.Exercises[0].Sets[0]

	t.Run("Add Set Entry Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

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
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

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
			helpers.AddContext(u),
		)

		err := mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Set Invalid Token", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 225.0, reps: 8 })
			}
			`,
			&resp,
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Set: Invalid Token\",\"path\":[\"addSet\"]}]")
	})

	t.Run("Add Set Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

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
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 225.0, reps: 8 })
			}
			`,
			&resp,
			helpers.AddContext(u),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Set\",\"path\":[\"addSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Set Error Getting Exercise", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		const getExercisesQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExercisesQuery)).
			WithArgs(e.ID).
			WillReturnError(gorm.ErrRecordNotFound)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 225.0, reps: 8 })
			}
			`,
			&resp,
			helpers.AddContext(u),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Set record not found\",\"path\":[\"addSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Set Error Adding Set", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

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
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()
		addSetEntriesQuery := `INSERT INTO "set_entries" ("created_at","updated_at","deleted_at","weight","reps","exercise_id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addSetEntriesQuery)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), s.Weight, s.Reps, s.ExerciseID).
			WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 225.0, reps: 8 })
			}
			`,
			&resp,
			helpers.AddContext(u),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Set\",\"path\":[\"addSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Set Entries Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

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
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

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
			helpers.AddContext(u),
		)

		err := mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Set Entries Invalid Token", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp GetSetEntriesResp
		err := c.Post(`
			query GetSets {
				sets(exerciseId: "44") {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Sets\",\"path\":[\"sets\"]}]")
	})

	t.Run("Get Set Entries Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

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

		var resp GetSetEntriesResp
		err := c.Post(`
			query GetSets {
				sets(exerciseId: "44") {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Sets: Access Denied\",\"path\":[\"sets\"]}]")
	})
}
