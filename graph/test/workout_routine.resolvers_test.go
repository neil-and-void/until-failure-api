package test

import (
	"testing"
)


func TestWorkoutRoutineResolvers(t *testing.T) {
	// // access_token := token.Sign(&token.Credentials{
	// // 	ID: 28,
	// // 	Name: "test",
	// // 	Email: "test@test.com",
	// // }, []byte("supersecret"), config.ACCESS_TTL)

	// t.Run("Workout Routine resolver success", func(t *testing.T) {
	// 	mock, gormDB := tests.SetupMockDB()
	// 	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
	// 		DB: gormDB,
	// 	}})))

	// 	nullUser := sqlmock.
	// 		NewRows([]string{"id", "name", "user_id", "created_at", "deleted_at", "updated_at"}).
	// 		AddRow(28, "", "", time.Time{}, time.Time{}, time.Time{})

	// 	const userQuery = `SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`
	// 	mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(u.Email).WillReturnRows(nullUser)

	// 	mock.ExpectBegin()
	// 	const createQuery = `INSERT INTO "users" ("created_at","updated_at","deleted_at","name","email","password") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
	// 	mock.ExpectQuery(regexp.QuoteMeta(createQuery)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), u.Name, u.Email, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(u.ID))
	// 	mock.ExpectCommit()

	// 	var resp struct {
	// 		Signup struct {
	// 			AccessToken  string
	// 			RefreshToken string
	// 		}
	// 	}
	// 	c.MustPost(`mutation CreateWorkoutRoutine {
	// 		createWorkoutRoutine(
	// 		  routine: {
	// 			name: "Legs",
	// 			exerciseRoutines:[]
	// 		  }
	// 		) {
	// 			  id
	// 			  name
	// 			  exerciseRoutines {
	// 				  id
	// 			  }
	// 		}
	// 	  }`,
	// 		&resp)

	// 	assert.True(t, token.Validate(resp.Signup.AccessToken, ACCESS_SECRET))
	// 	assert.True(t, token.Validate(resp.Signup.RefreshToken, REFRESH_SECRET))

	// 	err = mock.ExpectationsWereMet() // make sure all expectations were met
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// })		
}
