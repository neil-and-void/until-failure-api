package test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/accesscontroller/accesscontrol"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/helpers"
	"github.com/neilZon/workout-logger-api/tests/testdata"
	"github.com/neilZon/workout-logger-api/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type AddWorkoutSessionResp struct {
	AddWorkoutSession string
}

type GetWorkoutSession struct {
	WorkoutSessions []struct {
		ID               string
		Start            string
		End              string
		WorkoutRoutineId string
		Exercises        []struct {
			ExerciseRoutineId string
			ID                string
			Sets              []struct {
				ID     string
				Weight float32
				Reps   int
				Notes  string
			}
		}
	}
}

type GetWorkoutSessions struct {
	WorkoutSessions struct {
		Edges []struct {
			Node struct {
				ID    string
				Start string
				End   string
				// WorkoutRoutineId string
				Exercises []struct {
					ExerciseRoutine model.ExerciseRoutine
					Notes           string
					ID              string
					Sets            []struct {
						ID     string
						Weight float32
						Reps   int
					}
				}
			}
			Cursor string
		}
		PageInfo struct{}
	}
}

type UpdateWorkoutSession struct {
	UpdateWorkoutSession struct {
		ID    string
		Start string
		End   string
	}
}

type DeleteWorkoutSessionResp struct {
	DeleteWorkoutSession int
}

func TestWorkoutSessionResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	ws := testdata.WorkoutSession
	// wr := testdata.WorkoutRoutine
	u := testdata.User

	t.Run("Add Workout Session success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(db)
		c := helpers.NewGqlClient(gormDB, acs)

		mock.ExpectBegin()

		const addWorkoutSessionStmnt = `INSERT INTO "workout_sessions" ("created_at","updated_at","deleted_at","start","end","workout_routine_id","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addWorkoutSessionStmnt)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ws.Start, nil, ws.WorkoutRoutineID, ws.UserID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ws.ID))

		const addExerciseStmt = `INSERT INTO "exercises" ("created_at","updated_at","deleted_at","notes","exercise_routine_id","workout_session_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) ON CONFLICT ("id") DO UPDATE SET "workout_session_id"="excluded"."workout_session_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addExerciseStmt)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[0].Notes,
			ws.Exercises[0].ExerciseRoutineID,
			ws.Exercises[0].WorkoutSessionID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[1].Notes,
			ws.Exercises[1].ExerciseRoutineID,
			ws.Exercises[1].WorkoutSessionID,
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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Workout Session Access Invalid Token", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

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
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"addWorkoutSession\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")
	})

	t.Run("Add Workout Session Error (invalid workout routine ID fk constraint)", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		mock.ExpectBegin()

		const addWorkoutSessionStmnt = `INSERT INTO "workout_sessions" ("created_at","updated_at","deleted_at","start","end","workout_routine_id","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addWorkoutSessionStmnt)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ws.Start, nil, 8789, ws.UserID).
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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Workout Session\",\"path\":[\"addWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Add Workout Session Error (invalid exercise ID fk constraint)", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		mock.ExpectBegin()

		const addWorkoutSessionStmnt = `INSERT INTO "workout_sessions" ("created_at","updated_at","deleted_at","start","end","workout_routine_id","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addWorkoutSessionStmnt)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ws.Start, nil, ws.WorkoutRoutineID, ws.UserID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ws.ID))

		const addExerciseStmt = `INSERT INTO "exercises" ("created_at","updated_at","deleted_at","notes","exercise_routine_id","workout_session_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) ON CONFLICT ("id") DO UPDATE SET "workout_session_id"="excluded"."workout_session_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(addExerciseStmt)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[0].Notes,
			ws.Exercises[0].ExerciseRoutineID,
			ws.Exercises[0].WorkoutSessionID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			ws.Exercises[1].Notes,
			ws.Exercises[1].ExerciseRoutineID,
			ws.Exercises[1].WorkoutSessionID,
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
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)
		require.EqualError(t, err, "[{\"message\":\"Error Adding Workout Session\",\"path\":[\"addWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Workout Session", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(utils.UIntToString(ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()

		updatedWorkoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		updateWorkoutSessionStmt := `UPDATE "workout_sessions" SET "updated_at"=$1,"end"=$2 WHERE id = $3 AND "workout_sessions"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(updateWorkoutSessionStmt)).
			WithArgs(sqlmock.AnyArg(), ws.End, utils.UIntToString(ws.ID)).
			WillReturnRows(updatedWorkoutSessionRow)

		mock.ExpectCommit()

		gqlQuery := fmt.Sprintf(`
			mutation UpdateWorkoutSession {
				updateWorkoutSession(workoutSessionId: "%d", updateWorkoutSessionInput: {
					end: "%s",
				}) {
					id
					start
					end
				}
			}`, ws.ID, ws.End.Format(time.RFC3339))
		var resp UpdateWorkoutSession
		c.MustPost(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Workout Session Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		gqlQuery := fmt.Sprintf(`
			mutation UpdateWorkoutSession {
				updateWorkoutSession(workoutSessionId: "%d", updateWorkoutSessionInput: {
					end: "%s",
				}) {
					id
					start
					end
				}
			}`, ws.ID, ws.End.Format(time.RFC3339))
		var resp UpdateWorkoutSession
		err := c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"updateWorkoutSession\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Workout Session Acces Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		badUserId := 1423
		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, badUserId, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(utils.UIntToString(ws.ID)).WillReturnRows(workoutSessionRow)

		gqlQuery := fmt.Sprintf(`
			mutation UpdateWorkoutSession {
				updateWorkoutSession(workoutSessionId: "%d", updateWorkoutSessionInput: {
					end: "%s",
				}) {
					id
					start
					end
				}
			}`, ws.ID, ws.End.Format(time.RFC3339))
		var resp UpdateWorkoutSession
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Updating Workout Session: Access Denied\",\"path\":[\"updateWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Workout Session Error", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(utils.UIntToString(ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()

		updateWorkoutSessionStmt := `UPDATE "workout_sessions" SET "updated_at"=$1,"end"=$2 WHERE id = $3 AND "workout_sessions"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(updateWorkoutSessionStmt)).
			WithArgs(sqlmock.AnyArg(), ws.End, utils.UIntToString(ws.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)

		mock.ExpectRollback()

		gqlQuery := fmt.Sprintf(`
			mutation UpdateWorkoutSession {
				updateWorkoutSession(workoutSessionId: "%d", updateWorkoutSessionInput: {
					end: "%s",
				}) {
					id
					start
					end
				}
			}`, ws.ID, ws.End.Format(time.RFC3339))
		var resp UpdateWorkoutSession
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Updating Workout Session\",\"path\":[\"updateWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Session Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(utils.UIntToString(ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()
		deleteWorkoutSessionQuery := `UPDATE "workout_sessions" SET "deleted_at"=$1 WHERE id = $2 AND "workout_sessions"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteWorkoutSessionQuery)).WithArgs(sqlmock.AnyArg(), utils.UIntToString(ws.ID)).WillReturnResult(sqlmock.NewResult(1, 1))

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"})
		for _, e := range ws.Exercises {
			exerciseRows.AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		}
		deleteExerciseQuery := `UPDATE "exercises" SET "deleted_at"=$1 WHERE workout_session_id = $2 AND "exercises"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(deleteExerciseQuery)).WithArgs(sqlmock.AnyArg(), utils.UIntToString(ws.ID)).WillReturnRows(exerciseRows)

		deleteSetEntryQuery := `UPDATE "set_entries" SET "deleted_at"=$1 WHERE exercise_id IN ($2,$3) AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteSetEntryQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(ws.Exercises[0].ID), utils.UIntToString(ws.Exercises[1].ID)).
			WillReturnResult(sqlmock.NewResult(1, 2))

		mock.ExpectCommit()

		gqlQuery := fmt.Sprintf(`mutation DeleteWorkoutSession {
			deleteWorkoutSession(workoutSessionId: "%d")
		}`, ws.ID)
		var resp DeleteWorkoutSessionResp
		c.MustPost(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Session Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		gqlQuery := fmt.Sprintf(`mutation DeleteWorkoutSession {
			deleteWorkoutSession(workoutSessionId: "%d")
		}`, ws.ID)
		var resp DeleteWorkoutSessionResp
		err := c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"deleteWorkoutSession\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Session Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		badUserId := 142
		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, badUserId, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(utils.UIntToString(ws.ID)).WillReturnRows(workoutSessionRow)

		gqlQuery := fmt.Sprintf(`mutation DeleteWorkoutSession {
			deleteWorkoutSession(workoutSessionId: "%d")
		}`, ws.ID)
		var resp DeleteWorkoutSessionResp
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Workout Session: Access Denied\",\"path\":[\"deleteWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Session Error", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutSessionAccessQuery)).WithArgs(utils.UIntToString(ws.ID)).WillReturnRows(workoutSessionRow)

		mock.ExpectBegin()
		deleteWorkoutSessionQuery := `UPDATE "workout_sessions" SET "deleted_at"=$1 WHERE id = $2 AND "workout_sessions"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteWorkoutSessionQuery)).WithArgs(sqlmock.AnyArg(), utils.UIntToString(ws.ID)).WillReturnError(gorm.ErrInvalidTransaction)
		mock.ExpectRollback()

		gqlQuery := fmt.Sprintf(`mutation DeleteWorkoutSession {
			deleteWorkoutSession(workoutSessionId: "%d")
		}`, ws.ID)
		var resp DeleteWorkoutSessionResp
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Workout Session\",\"path\":[\"deleteWorkoutSession\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}
