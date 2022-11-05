package accesscontrol

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neilZon/workout-logger-api/e2e/testdata"
	"github.com/neilZon/workout-logger-api/helpers"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAccessControl(t *testing.T) {
	wr := testdata.WorkoutRoutine
	ws := testdata.WorkoutSession

	t.Run("Test Can Access Workout Routine Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()

		userId := fmt.Sprintf("%d", wr.UserID)
		workoutRoutineId := fmt.Sprintf("%d", wr.ID)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "user_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(wr.ID, wr.Name, wr.UserID, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt)

		const userQuery = `SELECT * FROM "workout_routines" WHERE (user_id = $1 AND id = $2) AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(userId, workoutRoutineId).WillReturnRows(workoutRoutineRow)

		ac := &AccessController{DB: gormDB}
		err := ac.CanAccessWorkoutRoutine(userId, workoutRoutineId)
		require.Nil(t, err, "Should be no error for accessing workout routine")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Test Can Access Workout Routine Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()

		userId := fmt.Sprintf("%d", wr.UserID)
		workoutRoutineId := fmt.Sprintf("%d", wr.ID)

		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(userId, workoutRoutineId).WillReturnError(gorm.ErrRecordNotFound)

		ac := &AccessController{DB: gormDB}
		err := ac.CanAccessWorkoutRoutine(userId, workoutRoutineId)
		require.Equal(t, err.Error(), "Access Denied")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Test Can Access Workout Session Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()

		userId := fmt.Sprintf("%d", ws.UserID)
		workoutSessionId := fmt.Sprintf("%d", ws.ID)

		workoutSessionRow := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)

		const workoutSessionQuery = `SELECT * FROM "workout_sessions" WHERE (user_id = $1 AND id = $2) AND "workout_sessions"."deleted_at" IS NULL ORDER BY "workout_sessions"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(workoutSessionQuery)).WithArgs(userId, workoutSessionId).WillReturnRows(workoutSessionRow)

		ac := &AccessController{DB: gormDB}
		err := ac.CanAccessWorkoutSession(userId, workoutSessionId)
		require.Nil(t, err, "Should be no error for accessing workout session")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Test Can Access Workout Session Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()

		userId := fmt.Sprintf("%d", ws.UserID)
		workoutSessionId := fmt.Sprintf("%d", ws.ID)

		const workoutSessionQuery = `SELECT * FROM "workout_sessions" WHERE (user_id = $1 AND id = $2) AND "workout_sessions"."deleted_at" IS NULL ORDER BY "workout_sessions"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(workoutSessionQuery)).WithArgs(userId, workoutSessionId).WillReturnError(gorm.ErrRecordNotFound)

		ac := &AccessController{DB: gormDB}
		err := ac.CanAccessWorkoutSession(userId, workoutSessionId)
		require.Equal(t, err.Error(), "Access Denied")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}
