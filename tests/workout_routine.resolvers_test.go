package test

import (
	"fmt"
	"regexp"
	"testing"

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

type WorkoutRoutineResp struct {
	CreateWorkoutRoutine struct {
		ID               string
		Name             string
		ExerciseRoutines []model.ExerciseRoutine
	}
}

type GetWorkoutRoutinesResp struct {
	WorkoutRoutines struct {
		Edges []struct {
			Node model.WorkoutRoutine
		}
		PageInfo struct {
			HasNextPage bool
		}
	}
}

type GetWorkoutRoutineResp struct {
	WorkoutRoutine struct {
		ID               string
		Name             string
		Active           bool
		ExerciseRoutines []struct {
			ID     string
			Name   string
			Active bool
			Sets   int
			Reps   int
		}
	}
}

type UpdateWorkoutRoutine struct {
	UpdateWorkoutRoutine struct {
		ID               string
		Name             string
		ExerciseRoutines []struct {
			ID   string
			Name string
			Sets int
			Reps int
		}
	}
}

type DeleteWorkoutRoutineResp struct {
	DeleteWorkoutRoutine int
}

func TestWorkoutRoutineResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	wr := testdata.WorkoutRoutine
	ws := testdata.WorkoutSession
	u := testdata.User

	t.Run("Create workout routine success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		mock.ExpectBegin()
		const createWorkoutRoutineStmnt = `INSERT INTO "workout_routines" ("created_at","updated_at","deleted_at","name","active","user_id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createWorkoutRoutineStmnt)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), wr.Name, wr.Active, wr.UserID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wr.ID))
		const createExerciseRoutineStmt = `INSERT INTO "exercise_routines" ("created_at","updated_at","deleted_at","name","sets","reps","active","workout_routine_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8),($9,$10,$11,$12,$13,$14,$15,$16) ON CONFLICT ("id") DO UPDATE SET "workout_routine_id"="excluded"."workout_routine_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createExerciseRoutineStmt)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			wr.ExerciseRoutines[0].Name,
			wr.ExerciseRoutines[0].Sets,
			wr.ExerciseRoutines[0].Reps,
			wr.ExerciseRoutines[0].Active,
			wr.ExerciseRoutines[0].WorkoutRoutineID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			wr.ExerciseRoutines[1].Name,
			wr.ExerciseRoutines[1].Sets,
			wr.ExerciseRoutines[1].Reps,
			wr.ExerciseRoutines[1].Active,
			wr.ExerciseRoutines[1].WorkoutRoutineID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wr.ExerciseRoutines[0].ID).AddRow(wr.ExerciseRoutines[1].ID))
		mock.ExpectCommit()

		var resp WorkoutRoutineResp
		c.MustPost(`mutation CreateWorkoutRoutine {
			createWorkoutRoutine(
			  routine: {
				name: "Legs",
				exerciseRoutines:[
					{
						name: "squat",
						sets: 4,
						reps: 6
					},
					{
						name: "leg extensions",
						sets: 4,
						reps: 6
					}
				]
			  }
			) {
				  id
				  name
				  exerciseRoutines {
					  id
				  }
			}
		  }`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Create workout routine invalid data", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp WorkoutRoutineResp
		err = c.Post(`mutation CreateWorkoutRoutine {
			createWorkoutRoutine(
			  routine: {
				name: "a",
				exerciseRoutines:[]
			  }
			) {
				  id
				  name
				  exerciseRoutines {
					  id
				  }
			}
		  }`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Invalid Routine Name Length\",\"path\":[\"createWorkoutRoutine\"]}]")
	})

	t.Run("Create workout routine no token", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp WorkoutRoutineResp
		err := c.Post(`mutation CreateWorkoutRoutine {
			createWorkoutRoutine(
			  routine: {
				name: "Legs",
				exerciseRoutines:[]
			  }
			) {
				  id
				  name
				  exerciseRoutines {
					  id
				  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"createWorkoutRoutine\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")
	})

	t.Run("Get Workout Routines Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt)

		const workoutRoutineQuery = `SELECT * FROM "workout_routines" WHERE user_id = $1 AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(workoutRoutineQuery)).WithArgs(fmt.Sprintf("%d", u.ID)).WillReturnRows(workoutRoutineRow)

		exerciseRoutineRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "name", "sets", "reps", "workout_routine_id"})
		for _, er := range wr.ExerciseRoutines {
			exerciseRoutineRows.AddRow(er.ID, er.CreatedAt, er.DeletedAt, er.UpdatedAt, er.Name, er.Sets, er.Reps, er.WorkoutRoutineID)
		}
		const getExerciseRoutinesQuery = `SELECT * FROM "exercise_routines" WHERE workout_routine_id IN ($1) AND "exercise_routines`
		mock.ExpectQuery(regexp.QuoteMeta(getExerciseRoutinesQuery)).
			WithArgs(utils.UIntToString(wr.ID)).
			WillReturnRows(exerciseRoutineRows)

		var resp GetWorkoutRoutinesResp
		c.MustPost(`
			query WorkoutRoutines {
				workoutRoutines(limit: 6) {
					edges {
						node {
							id
							name
							active
							exerciseRoutines {
								id
								active
								name
								sets
								reps
							}
						}
					}
					pageInfo {
						hasNextPage
					}
				}
			}`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Workout Routines No Token", func(t *testing.T) {
		_, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp GetWorkoutRoutinesResp
		err := c.Post(`
			query WorkoutRoutines {
				workoutRoutines(limit: 6) {
					edges {
						node {
							id
							name
							active
							exerciseRoutines {
								id
								active
								name
								sets
								reps
							}
						}
					}
					pageInfo {
						hasNextPage
					}
				}
			}`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"workoutRoutines\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")
	})

	t.Run("Get Workout Routine", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(utils.UIntToString(wr.ID)).WillReturnRows(workoutRoutineRow)

		workoutRoutineRow = sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		const workoutRoutineQuery = `SELECT * FROM "workout_routines" WHERE id = $1 AND "workout_routines"."deleted_at" IS NULL ORDER BY "workout_routines"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(workoutRoutineQuery)).WithArgs(utils.UIntToString(wr.ID)).WillReturnRows(workoutRoutineRow)

		exerciseRoutineRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "name", "sets", "reps", "workout_routine_id"})
		for _, er := range wr.ExerciseRoutines {
			exerciseRoutineRows.AddRow(er.ID, er.CreatedAt, er.DeletedAt, er.UpdatedAt, er.Name, er.Sets, er.Reps, er.WorkoutRoutineID)
		}
		const getExerciseRoutinesQuery = `SELECT * FROM "exercise_routines" WHERE workout_routine_id IN ($1) AND "exercise_routines"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(getExerciseRoutinesQuery)).
			WithArgs(utils.UIntToString(wr.ID)).
			WillReturnRows(exerciseRoutineRows)

		var resp GetWorkoutRoutineResp
		c.MustPost(`
			query WorkoutRoutine {
				workoutRoutine(workoutRoutineId: "8") {
					id
					name
					active
					exerciseRoutines {
						id
						active
						name
						sets
						reps
					}
				}
			}`,
			&resp,
			helpers.AddContext(u, helpers.NewLoaders(gormDB)),
		)

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Workout Routine No Token", func(t *testing.T) {
	})

	t.Run("Update Workout Routine", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		mock.ExpectBegin()

		updateWorkoutRoutineStmt := `UPDATE "workout_routines" SET "name"=$1,"updated_at"=$2 WHERE id = $3 AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(updateWorkoutRoutineStmt)).
			WithArgs(wr.Name, sqlmock.AnyArg(), utils.UIntToString(wr.ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		exerciseRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "reps", "sets", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(
				wr.ExerciseRoutines[0].ID,
				wr.ExerciseRoutines[0].Name,
				wr.ExerciseRoutines[0].Reps,
				wr.ExerciseRoutines[0].Sets,
				utils.UIntToString(wr.ID),
				wr.ExerciseRoutines[0].CreatedAt,
				wr.ExerciseRoutines[0].DeletedAt,
				wr.ExerciseRoutines[0].UpdatedAt,
			)
		updateExerciseRoutineStmt := `INSERT INTO "exercise_routines" ("created_at","updated_at","deleted_at","name","sets","reps","active","workout_routine_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT ("id") DO UPDATE SET "reps"="excluded"."reps","sets"="excluded"."sets","name"="excluded"."name","active"="excluded"."active" RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(updateExerciseRoutineStmt)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				wr.ExerciseRoutines[0].Name,
				wr.ExerciseRoutines[0].Sets,
				wr.ExerciseRoutines[0].Reps,
				wr.Active,
				wr.ID,
				wr.ExerciseRoutines[0].ID,
			).WillReturnRows(exerciseRoutineRow)

		deleteExerciseRoutinesStmt := `UPDATE "exercise_routines" SET "deleted_at"=$1 WHERE (workout_routine_id = $2 AND id NOT IN ($3)) AND "exercise_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteExerciseRoutinesStmt)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(wr.ID), wr.ExerciseRoutines[0].ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		var resp UpdateWorkoutRoutine
		mutation := fmt.Sprintf(`
			mutation UpdateWorkoutRoutine {
				updateWorkoutRoutine(
					workoutRoutine: {
						id: "%d"
						name: "%s"
						exerciseRoutines: [
							{
								id: "%d",
								name: "%s",
								sets: %d,
								reps: %d
							}
						]
					}
				) {
					id
					name 
					exerciseRoutines {
						id
						name
						reps
						sets
					}
				}
			}`,
			wr.ID,
			wr.Name, wr.ExerciseRoutines[0].ID, wr.ExerciseRoutines[0].Name, wr.ExerciseRoutines[0].Sets, wr.ExerciseRoutines[0].Reps,
		)
		c.MustPost(mutation, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Workout Routine Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp UpdateWorkoutRoutine
		mutation := fmt.Sprintf(`
			mutation UpdateWorkoutRoutine {
				updateWorkoutRoutine(
					workoutRoutine: {
						id: "%d"
						name: "%s"
						exerciseRoutines: [
							{
								id: "%d",
								name: "%s",
								sets: %d,
								reps: %d
							}
						]
					}
				) {
					id
					name 
					exerciseRoutines {
						id
						name
						reps
						sets
					}
				}
			}`,
			wr.ID,
			wr.Name,
			wr.ExerciseRoutines[0].ID,
			wr.ExerciseRoutines[0].Name,
			wr.ExerciseRoutines[0].Sets,
			wr.ExerciseRoutines[0].Reps,
		)
		err := c.Post(mutation, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"updateWorkoutRoutine\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Workout Routine Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		someRandomId := 66
		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, someRandomId, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		var resp UpdateWorkoutRoutine
		mutation := fmt.Sprintf(`
			mutation UpdateWorkoutRoutine {
				updateWorkoutRoutine(
					workoutRoutine: {
						id: "%d"
						name: "%s"
						exerciseRoutines: [
							{
								id: "%d",
								name: "%s",
								sets: %d,
								reps: %d
							}
						]
					}
				) {
					id
					name 
					exerciseRoutines {
						id
						name
						reps
						sets
					}
				}
			}`,
			wr.ID,
			wr.Name,
			wr.ExerciseRoutines[0].ID,
			wr.ExerciseRoutines[0].Name,
			wr.ExerciseRoutines[0].Sets, wr.ExerciseRoutines[0].Reps,
		)
		err := c.Post(mutation, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Updating Workout Routine: Access Denied\",\"path\":[\"updateWorkoutRoutine\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Update Workout Routine Error", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		mock.ExpectBegin()

		updateWorkoutRoutineStmt := `UPDATE "workout_routines" SET "name"=$1,"updated_at"=$2 WHERE id = $3 AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(updateWorkoutRoutineStmt)).
			WithArgs(wr.Name, sqlmock.AnyArg(), utils.UIntToString(wr.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)

		mock.ExpectRollback()

		var resp UpdateWorkoutRoutine
		mutation := fmt.Sprintf(`
			mutation UpdateWorkoutRoutine {
				updateWorkoutRoutine(
					workoutRoutine: {
						id: "%d"
						name: "%s"
						exerciseRoutines: [
							{
								id: "%d",
								name: "%s",
								sets: %d,
								reps: %d
							}
						]
					}
				) {
					id
					name 
					exerciseRoutines {
						id
						name
						reps
						sets
					}
				}
			}`,
			wr.ID,
			wr.Name,
			wr.ExerciseRoutines[0].ID,
			wr.ExerciseRoutines[0].Name,
			wr.ExerciseRoutines[0].Sets, wr.ExerciseRoutines[0].Reps,
		)
		err := c.Post(mutation, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Updating Workout Routine\",\"path\":[\"updateWorkoutRoutine\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Routine Success", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		mock.ExpectBegin()

		deleteWorkoutRoutineQuery := `UPDATE "workout_routines" SET "deleted_at"=$1 WHERE id = $2 AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteWorkoutRoutineQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(wr.ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		deleteExerciseRoutinesQuery := `UPDATE "exercise_routines" SET "deleted_at"=$1 WHERE workout_routine_id = $2 AND "exercise_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteExerciseRoutinesQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(wr.ID)).
			WillReturnResult(sqlmock.NewResult(1, 2))

		workoutSessionRows := sqlmock.
			NewRows([]string{"id", "user_id", "start", "end", "workout_routine_id", "created_at", "deleted_at", "updated_at"}).
			AddRow(ws.ID, ws.UserID, ws.Start, ws.End, ws.WorkoutRoutineID, ws.CreatedAt, ws.DeletedAt, ws.UpdatedAt)
		deleteWorkoutSession := `UPDATE "workout_sessions" SET "deleted_at"=$1 WHERE workout_routine_id = $2 AND "workout_sessions"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(deleteWorkoutSession)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(wr.ID)).
			WillReturnRows(workoutSessionRows)

		exerciseRows := sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "updated_at", "workout_session_id", "exercise_routine_id"})
		for _, e := range ws.Exercises {
			exerciseRows.AddRow(e.ID, e.CreatedAt, e.DeletedAt, e.UpdatedAt, e.WorkoutSessionID, e.ExerciseRoutineID)
		}

		deleteExerciseQuery := `UPDATE "exercises" SET "deleted_at"=$1 WHERE workout_session_id IN ($2) AND "exercises"."deleted_at" IS NULL RETURNING *`
		mock.ExpectQuery(regexp.QuoteMeta(deleteExerciseQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(ws.ID)).
			WillReturnRows(exerciseRows)

		deleteSetQuery := `UPDATE "set_entries" SET "deleted_at"=$1 WHERE exercise_id IN ($2,$3) AND "set_entries"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteSetQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(ws.Exercises[0].ID), utils.UIntToString(ws.Exercises[1].ID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		var resp DeleteWorkoutRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteWorkoutRoutine {
				deleteWorkoutRoutine(workoutRoutineId: "%d")
			}`,
			wr.ID,
		)
		c.MustPost(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Invalid Token", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		var resp DeleteWorkoutRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteWorkoutRoutine {
				deleteWorkoutRoutine(workoutRoutineId: "%d")
			}`,
			wr.ID,
		)
		err := c.Post(gqlQuery, &resp)
		require.EqualError(t, err, "[{\"message\":\"Unauthorized\",\"path\":[\"deleteWorkoutRoutine\"],\"extensions\":{\"code\":\"UNAUTHORIZED\"}}]")

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Access Denied", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		someRandomId := 66
		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, someRandomId, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		var resp DeleteWorkoutRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteWorkoutRoutine {
				deleteWorkoutRoutine(workoutRoutineId: "%d")
			}`,
			wr.ID,
		)

		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Workout Routine: Access Denied\",\"path\":[\"deleteWorkoutRoutine\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})

	t.Run("Delete Workout Error Deleting", func(t *testing.T) {
		mock, gormDB := helpers.SetupMockDB()
		acs := accesscontrol.NewAccessControllerService(gormDB)
		c := helpers.NewGqlClient(gormDB, acs)

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at", "user_id", "active"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt, wr.UserID, wr.Active)
		mock.ExpectQuery(regexp.QuoteMeta(helpers.WorkoutRoutineAccessQuery)).WithArgs(fmt.Sprintf("%d", wr.ID)).WillReturnRows(workoutRoutineRow)

		mock.ExpectBegin()

		deleteWorkoutRoutineQuery := `UPDATE "workout_routines" SET "deleted_at"=$1 WHERE id = $2 AND "workout_routines"."deleted_at" IS NULL`
		mock.ExpectExec(regexp.QuoteMeta(deleteWorkoutRoutineQuery)).
			WithArgs(sqlmock.AnyArg(), utils.UIntToString(wr.ID)).
			WillReturnError(gorm.ErrInvalidTransaction)

		mock.ExpectRollback()

		var resp DeleteWorkoutRoutineResp
		gqlQuery := fmt.Sprintf(`
			mutation DeleteWorkoutRoutine {
				deleteWorkoutRoutine(workoutRoutineId: "%d")
			}`,
			wr.ID,
		)
		err := c.Post(gqlQuery, &resp, helpers.AddContext(u, helpers.NewLoaders(gormDB)))
		require.EqualError(t, err, "[{\"message\":\"Error Deleting Workout Routine\",\"path\":[\"deleteWorkoutRoutine\"]}]")

		err = mock.ExpectationsWereMet()
		if err != nil {
			panic(err)
		}
	})
}
