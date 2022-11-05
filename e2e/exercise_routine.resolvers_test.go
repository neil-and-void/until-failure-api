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
	"gorm.io/gorm"
)

type GetExerciseRoutineResp struct {
	ExerciseRoutines []struct {
		ID   string
		Name string
		Sets int
		Reps int
	}
}

func TestExerciseRoutineResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	u := testdata.User
	wr := testdata.WorkoutRoutine
	er := testdata.WorkoutRoutine.ExerciseRoutines[0]

	t.Run("Get Exercise Routine Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		exerciseRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "sets", "reps", "created_at", "deleted_at", "updated_at"}).
			AddRow(er.ID, er.Name, er.Sets, er.Reps, er.CreatedAt, er.DeletedAt, er.UpdatedAt)
		const exerciseRoutineQuery = `SELECT exercise_routines.id, exercise_routines.name, exercise_routines.sets, exercise_routines.reps, exercise_routines.created_at, exercise_routines.updated_at, exercise_routines.deleted_at FROM "workout_routines" left join exercise_routines on workout_routines.id = exercise_routines.workout_routine_id WHERE exercise_routines.workout_routine_id = $1 AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(exerciseRoutineQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(exerciseRoutineRow)

		var resp GetExerciseRoutineResp
		query := fmt.Sprintf(`query ExerciseRoutines {
			exerciseRoutines(workoutRoutineId: "%d") {
				id
				name
			}
		}`, er.WorkoutRoutineID)
		c.MustPost(query, &resp, helpers.AddContext(u))
	})

	t.Run("Get Exercise Routine Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", 1234)).WillReturnError(gorm.ErrRecordNotFound)

		var resp GetExerciseRoutineResp
		query := fmt.Sprintf(`query ExerciseRoutines {
			exerciseRoutines(workoutRoutineId: "%d") {
				id
				name
			}
		}`, 1234)
		err = c.Post(query, &resp, helpers.AddContext(u))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})
}
