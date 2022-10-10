package test

import (
	"regexp"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/common/database"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/utils/config"
	"github.com/neilZon/workout-logger-api/utils/token"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type WorkoutRoutineResp struct {
	CreateWorkoutRoutine struct {
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

type GetWorkoutRoutineResp struct {
	WorkoutRoutines []struct {
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

func TestWorkoutRoutineResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../../.env")
	if err != nil {
		panic("Error loading .env file")
	}
	u := &token.Claims{
		Name: "test",
		ID:   28,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(config.ACCESS_TTL * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
			Issuer:    "neil:)",
			Subject:   "test@test.com",
		},
	}

	wr := &database.WorkoutRoutine{
		Name: "Legs",
		ExerciseRoutines: []database.ExerciseRoutine{
			{
				Model: gorm.Model{
					ID:        3,
					CreatedAt: time.Now(),
					DeletedAt: gorm.DeletedAt{
						Time:  time.Time{},
						Valid: true,
					},
					UpdatedAt: time.Now(),
				},
				Name:             "squat",
				Sets:             4,
				Reps:             6,
				WorkoutRoutineID: 8,
			},
			{
				Model: gorm.Model{
					ID:        4,
					CreatedAt: time.Now(),
					DeletedAt: gorm.DeletedAt{
						Time:  time.Time{},
						Valid: true,
					},
					UpdatedAt: time.Now(),
				},
				Name:             "leg extensions",
				Sets:             4,
				Reps:             6,
				WorkoutRoutineID: 8,
			},
		},
		UserID: 28,
		Model: gorm.Model{
			ID:        8,
			CreatedAt: time.Now(),
			DeletedAt: gorm.DeletedAt{
				Time:  time.Time{},
				Valid: true,
			},
			UpdatedAt: time.Now(),
		},
	}

	t.Run("Create workout routine success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		mock.ExpectBegin()
		const createWorkoutRoutineStmnt = `INSERT INTO "workout_routines" ("created_at","updated_at","deleted_at","name","user_id") VALUES ($1,$2,$3,$4,$5) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createWorkoutRoutineStmnt)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), wr.Name, wr.UserID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wr.ID))
		const createExerciseRoutineStmt = `INSERT INTO "exercise_routines" ("created_at","updated_at","deleted_at","name","sets","reps","workout_routine_id") VALUES ($1,$2,$3,$4,$5,$6,$7),($8,$9,$10,$11,$12,$13,$14) ON CONFLICT ("id") DO UPDATE SET "workout_routine_id"="excluded"."workout_routine_id" RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createExerciseRoutineStmt)).WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			wr.ExerciseRoutines[0].Name,
			wr.ExerciseRoutines[0].Sets,
			wr.ExerciseRoutines[0].Reps,
			wr.ExerciseRoutines[0].WorkoutRoutineID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			wr.ExerciseRoutines[1].Name,
			wr.ExerciseRoutines[1].Sets,
			wr.ExerciseRoutines[1].Reps,
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
			AddContext(u))
		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Create workout routine invalid data", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

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
			AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Invalid Routine Name Length\",\"path\":[\"createWorkoutRoutine\"]}]")
	})

	t.Run("Create workout routine no token", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

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
		require.EqualError(t, err, "[{\"message\":\"Error Creating Workout: Invalid Token\",\"path\":[\"createWorkoutRoutine\"]}]")
	})

	t.Run("Get Workout Routines Success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		workoutRoutineRow := sqlmock.
			NewRows([]string{"id", "name", "created_at", "deleted_at", "updated_at"}).
			AddRow(wr.ID, wr.Name, wr.CreatedAt, wr.DeletedAt, wr.UpdatedAt)

		const workoutRoutineQuery = `
		SELECT workout_routines.id, workout_routines.name, workout_routines.created_at, workout_routines.updated_at, workout_routines.deleted_at 
		FROM "users" left join workout_routines on workout_routines.user_id = users.id 
		WHERE users.email = $1 AND "users"."deleted_at" IS NULL`
		mock.ExpectQuery(regexp.QuoteMeta(workoutRoutineQuery)).WithArgs(u.Subject).WillReturnRows(workoutRoutineRow)

		var resp GetWorkoutRoutineResp
		c.MustPost(`query WorkoutRoutines {
			workoutRoutines {
				id
				name
			}
		}`,
			&resp,
			AddContext(u))
		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Get Workout Routine No Token", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		var resp GetWorkoutRoutineResp
		err := c.Post(`query WorkoutRoutines {
			workoutRoutines {
				id
				name
			}
		}`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Error Getting Workout Routine: Invalid Token\",\"path\":[\"workoutRoutines\"]}]")
	})

	t.Run("Get Workout Routine DB Error", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		var resp GetWorkoutRoutineResp
		err := c.Post(`query WorkoutRoutines {
			workoutRoutines {
				id
				name
			}
		}`,
			&resp, AddContext(u))
		require.EqualError(t, err, "[{\"message\":\"Error Getting Workout Routine\",\"path\":[\"workoutRoutines\"]}]")
	})
}
