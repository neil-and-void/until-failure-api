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

type UpdateSetResp struct {
	UpdateSet struct {
		ID     string
		Weight float32
		Reps   int
	}
}

type DeleteSetResp struct {
	DeleteSet int
}

func TestSetEntryResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../.env")
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

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		require.Equal(t, resp.AddSet, utils.UIntToString(s.ID), "Created Id's don't match")

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
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"addSet\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")
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

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 225.0, reps: 8 })
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Set\",\"path\":[\"addSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})


	t.Run("Add Set Entry Too Much Reps", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 100, reps: 293084 })
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Reps needs to be between 0 and 9999\",\"path\":[\"addSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Set Entry Too little Reps", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 225.0, reps: -23 })
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Reps needs to be between 0 and 9999\",\"path\":[\"addSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Set Entry Too Much Weight", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: 423987, reps: 8 })
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Weight needs to be between 0 and 9999\",\"path\":[\"addSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Set Entry Too little Weight", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp AddSetEntryResp
		err := c.Post(`
			mutation AddSet {
				addSet(exerciseId: "44", set: {weight: -423987, reps: 8 })
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Weight needs to be between 0 and 9999\",\"path\":[\"addSet\"]}]")

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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Set: record not found\",\"path\":[\"addSet\"]}]")

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

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
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
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
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
		)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"sets\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")
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

		incorrectUserId := 444
		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, incorrectUserId, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Sets: Access Denied\",\"path\":[\"sets\"]}]")
	})

	t.Run("Update Set Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"}).
			AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		const getSetEntry = `SELECT * FROM "set_entries" WHERE id = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntry)).
			WithArgs(fmt.Sprintf("%d", s.ID)).
			WillReturnRows(setEntryRows)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExerciseQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExerciseQuery)).
			WithArgs(s.ExerciseID).
			WillReturnRows(exerciseRows)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		setEntryRow := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"}).
			AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)

		mock.ExpectBegin()
		updateSetQuery := `UPDATE "set_entries" SET "updated_at"=$1,"weight"=$2 WHERE id = $3 AND "set_entries"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(updateSetQuery)).
			WithArgs(sqlmock.AnyArg(), float64(225), fmt.Sprintf("%d", s.ID)).
			WillReturnRows(setEntryRow)
		mock.ExpectCommit()

		var resp UpdateSetResp
		c.MustPost(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { weight: 225.0 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		err := mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Set Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp UpdateSetResp
		err := c.Post(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { weight: 225.0 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
		)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"updateSet\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Set Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"}).
			AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		const getSetEntry = `SELECT * FROM "set_entries" WHERE id = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntry)).
			WithArgs(fmt.Sprintf("%d", s.ID)).
			WillReturnRows(setEntryRows)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExerciseQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExerciseQuery)).
			WithArgs(s.ExerciseID).
			WillReturnRows(exerciseRows)

		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnError(gorm.ErrRecordNotFound)

		var resp UpdateSetResp
		err := c.Post(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { weight: 225.0 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Updating Set: Access Denied\",\"path\":[\"updateSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Set Entry Too Much Reps", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp UpdateSetResp
		err := c.Post(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { reps: 213908 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		require.EqualError(t, err, "[{\"message\":\"Reps needs to be between 0 and 9999\",\"path\":[\"updateSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Set Entry Too little Reps", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp UpdateSetResp
		err := c.Post(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { reps: -1 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		require.EqualError(t, err, "[{\"message\":\"Reps needs to be between 0 and 9999\",\"path\":[\"updateSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Set Entry Too Much Weight", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp UpdateSetResp
		err := c.Post(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { weight: 213908 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		require.EqualError(t, err, "[{\"message\":\"Weight needs to be between 0 and 9999\",\"path\":[\"updateSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Set Entry Too little Weight", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp UpdateSetResp
		err := c.Post(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { weight: -0.1 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		require.EqualError(t, err, "[{\"message\":\"Weight needs to be between 0 and 9999\",\"path\":[\"updateSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})


	t.Run("Update Set Error Updating Set", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"}).
			AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		const getSetEntry = `SELECT * FROM "set_entries" WHERE id = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntry)).
			WithArgs(fmt.Sprintf("%d", s.ID)).
			WillReturnRows(setEntryRows)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExerciseQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExerciseQuery)).
			WithArgs(s.ExerciseID).
			WillReturnRows(exerciseRows)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()
		updateSetQuery := `UPDATE "set_entries" SET "updated_at"=$1,"weight"=$2 WHERE id = $3 AND "set_entries"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(updateSetQuery)).
			WithArgs(sqlmock.AnyArg(), float64(225), fmt.Sprintf("%d", s.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		var resp UpdateSetResp
		err = c.Post(`
			mutation UpdateSet {
				updateSet(setId: "30", set: { weight: 225.0 }) {
					id
					weight
					reps
				}
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Updating Set\",\"path\":[\"updateSet\"]}]")

		err := mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Set Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"}).
			AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		const getSetEntry = `SELECT * FROM "set_entries" WHERE id = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntry)).
			WithArgs(fmt.Sprintf("%d", s.ID)).
			WillReturnRows(setEntryRows)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExerciseQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExerciseQuery)).
			WithArgs(s.ExerciseID).
			WillReturnRows(exerciseRows)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()
		deleteSetQuery := `UPDATE "set_entries" SET "deleted_at"=$1 WHERE id = $2 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteSetQuery)).
			WithArgs(sqlmock.AnyArg(), fmt.Sprintf("%d", s.ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		var resp DeleteSetResp
		c.MustPost(`
			mutation DeleteSet {
				deleteSet(setId: "30") 
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		err := mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Set Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp DeleteSetResp
		err := c.Post(`
			mutation DeleteSet {
				deleteSet(setId: "30") 
			}
			`,
			&resp,
		)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"deleteSet\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Set Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		setEntryRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "weight", "reps", "exercise_id"}).
			AddRow(s.ID, s.CreatedAt, s.DeletedAt, s.UpdatedAt, s.Weight, s.Reps, s.ExerciseID)
		const getSetEntry = `SELECT * FROM "set_entries" WHERE id = $1 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getSetEntry)).
			WithArgs(fmt.Sprintf("%d", s.ID)).
			WillReturnRows(setEntryRows)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExerciseQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExerciseQuery)).
			WithArgs(s.ExerciseID).
			WillReturnRows(exerciseRows)

		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnError(gorm.ErrRecordNotFound)

		var resp DeleteSetResp
		err := c.Post(`
			mutation DeleteSet {
				deleteSet(setId: "30") 
			}
			`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Set: Access Denied\",\"path\":[\"deleteSet\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}
