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
		Notes string
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
		Notes string
	}
}

type UpdateExerciseResp struct {
	UpdateExercise struct {
		ID    string
		Notes string
	}
}

type DeleteExerciseResp struct {
	DeleteExercise int
}

func TestExerciseResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	u := testdata.User
	ws := testdata.WorkoutSession
	e := testdata.WorkoutSession.Exercises[0]

	t.Run("Add Exercise Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()

		const createExerciseStmnt = `INSERT INTO "exercises" ("created_at","updated_at","deleted_at","notes","exercise_routine_id","workout_session_id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createExerciseStmnt)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), e.Notes, e.WorkoutSessionID, e.ExerciseRoutineID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(e.ID))

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
			helpers.AddContext(u),
		)

		require.Equal(t, resp.AddExercise, helpers.UIntToString(e.ID))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	// todo
	t.Run("Add Exercise Foreign Key Error", func(t *testing.T) {})

	t.Run("Add Exercise Invalid Token", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

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
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"addExercise\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")
	})

	t.Run("Add Exercise Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		incorrectUserId := 99
		workoutSessionRow := sqlmock.
		NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, incorrectUserId, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)

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
			1233,
		)
		err = c.Post(gqlMutation, &resp, helpers.AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Adding Exercise: Access Denied\",\"path\":[\"addExercise\"]}]")
	})

	t.Run("Get Exercise Success", func(t *testing.T) {
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
		c.MustPost(gqlQuery, &resp, helpers.AddContext(u))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Exercise Invalid Token", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(db)
		c := helpers.NewGqlClient(gormDB, acs)

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
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"exercise\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")
	})

	t.Run("Get Exercise Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		exerciseId := 788

		exerciseRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExercisesQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExercisesQuery)).
			WithArgs(exerciseId).
			WillReturnRows(exerciseRow)

		incorrectUserId := 444
		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, incorrectUserId, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnRows(workoutSessionRow)	

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
		err = c.Post(gqlQuery, &resp, helpers.AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Getting Exercise: Access Denied\",\"path\":[\"exercise\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Exercise Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		updatedNote := "BLAH"

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
		updateExerciseStmt := `UPDATE "exercises" SET "updated_at"=$1,"notes"=$2 WHERE id = $3 AND "exercises"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(updateExerciseStmt)).
			WithArgs(sqlmock.AnyArg(), updatedNote, fmt.Sprintf("%d", e.ID)).
			WillReturnRows(exerciseRow)
		mock.ExpectCommit()

		var resp UpdateExerciseResp
		gqlQuery := fmt.Sprintf(`	
			mutation UpdateExercise {
				updateExercise(exerciseId: "%d", exercise: { notes: "%s" }) {
					id
					notes
				}
			}`,
			e.ID,
			updatedNote,
		)
		c.MustPost(gqlQuery, &resp, helpers.AddContext(u))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Exercise Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		updatedNote := "BLAH"

		var resp UpdateExerciseResp
		gqlQuery := fmt.Sprintf(`	
			mutation UpdateExercise {
				updateExercise(exerciseId: "%d", exercise: { notes: "%s" }) {
					id
					notes
				}
			}`,
			e.ID,
			updatedNote,
		)
		err := c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"updateExercise\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Exercise Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		updatedNote := "BLAH"

		exerciseRow := sqlmock.
			NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"}).
			AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		const getExercisesQuery = `SELECT * FROM "exercises" WHERE "exercises"."deleted_at" IS NULL AND "exercises"."id" = $1 ORDER BY "exercises"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(getExercisesQuery)).
			WithArgs(e.ID).
			WillReturnRows(exerciseRow)

		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnError(gorm.ErrRecordNotFound)

		var resp UpdateExerciseResp
		gqlQuery := fmt.Sprintf(`	
			mutation UpdateExercise {
				updateExercise(exerciseId: "%d", exercise: { notes: "%s" }) {
					id
					notes
				}
			}`,
			e.ID,
			updatedNote,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Updating Exercise: Access Denied\",\"path\":[\"updateExercise\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Exercise db error updating exercise", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		updatedNote := "BLAH"

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
		updateExerciseStmt := `UPDATE "exercises" SET "updated_at"=$1,"notes"=$2 WHERE id = $3 AND "exercises"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(updateExerciseStmt)).
			WithArgs(sqlmock.AnyArg(), updatedNote, fmt.Sprintf("%d", e.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		var resp UpdateExerciseResp
		gqlQuery := fmt.Sprintf(`	
			mutation UpdateExercise {
				updateExercise(exerciseId: "%d", exercise: { notes: "%s" }) {
					id
					notes
				}
			}`,
			e.ID,
			updatedNote,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Updating Exercise\",\"path\":[\"updateExercise\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Success", func(t *testing.T) {
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
		deleteExerciseQuery := `UPDATE "exercises" SET "deleted_at"=$1 WHERE id = $2 AND "exercises"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteExerciseQuery)).
			WithArgs(sqlmock.AnyArg(), helpers.UIntToString(e.ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		deleteSetQuery := `UPDATE "set_entries" SET "deleted_at"=$1 WHERE exercise_id = $2 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteSetQuery)).
			WithArgs(sqlmock.AnyArg(), helpers.UIntToString(e.ID)).
			WillReturnResult(sqlmock.NewResult(1, 2))
		mock.ExpectCommit()

		var resp DeleteExerciseResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExercise {
				deleteExercise(exerciseId: "%d")
			}`,
			e.ID,
		)
		c.MustPost(gqlQuery, &resp, helpers.AddContext(u))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp DeleteExerciseResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExercise {
				deleteExercise(exerciseId: "%d")
			}`,
			e.ID,
		)
		err := c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"deleteExercise\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Access Denied", func(t *testing.T) {
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

		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(fmt.Sprintf("%d", ws.ID)).WillReturnError(gorm.ErrRecordNotFound)

		var resp DeleteExerciseResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExercise {
				deleteExercise(exerciseId: "%d")
			}`,
			e.ID,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Exercise: Access Denied\",\"path\":[\"deleteExercise\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Error, Update exercise tx", func(t *testing.T) {
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
		deleteExerciseQuery := `UPDATE "exercises" SET "deleted_at"=$1 WHERE id = $2 AND "exercises"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteExerciseQuery)).
			WithArgs(sqlmock.AnyArg(), helpers.UIntToString(e.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)

		mock.ExpectRollback()

		var resp DeleteExerciseResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExercise {
				deleteExercise(exerciseId: "%d")
			}`,
			e.ID,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Exercise\",\"path\":[\"deleteExercise\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Exercise Error, Update set entries tx", func(t *testing.T) {
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
		deleteExerciseQuery := `UPDATE "exercises" SET "deleted_at"=$1 WHERE id = $2 AND "exercises"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteExerciseQuery)).
			WithArgs(sqlmock.AnyArg(), helpers.UIntToString(e.ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		deleteSetQuery := `UPDATE "set_entries" SET "deleted_at"=$1 WHERE exercise_id = $2 AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteSetQuery)).
			WithArgs(sqlmock.AnyArg(), helpers.UIntToString(e.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		var resp DeleteExerciseResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteExercise {
				deleteExercise(exerciseId: "%d")
			}`,
			e.ID,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Exercise\",\"path\":[\"deleteExercise\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}
