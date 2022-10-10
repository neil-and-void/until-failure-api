package accesscontrol

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neilZon/workout-logger-api/graph/test"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAccessControl(t *testing.T) {
	wr := test.WorkoutRoutine

	t.Run("Test Access Control Success", func(t *testing.T) {
		mock, gormDB := test.SetupMockDB()

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
	})

	t.Run("Test Access Control Denied", func(t *testing.T) {
		mock, gormDB := test.SetupMockDB()

		userId := fmt.Sprintf("%d", wr.UserID)
		workoutRoutineId := fmt.Sprintf("%d", wr.ID)
		
		const userQuery = `SELECT * FROM "workout_routines" WHERE (user_id = $1 AND id = $2) AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(userId, workoutRoutineId).WillReturnError(gorm.ErrRecordNotFound)

		ac := &AccessController{DB: gormDB}
		err := ac.CanAccessWorkoutRoutine(userId, workoutRoutineId)
		require.Equal(t, err.Error(), "Access Denied")
	})
}
