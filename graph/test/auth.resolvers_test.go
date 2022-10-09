package test

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/neilZon/workout-logger-api/common/database"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/utils/config"
	"github.com/neilZon/workout-logger-api/utils/token"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestAuthResolvers(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../../.env")
	if err != nil {
		panic("Error loading .env file")
	}
	ACCESS_SECRET := []byte(os.Getenv(config.ACCESS_SECRET))
	REFRESH_SECRET := []byte(os.Getenv(config.REFRESH_SECRET))

	u := database.User{
		Model: gorm.Model{
			ID:        23,
			CreatedAt: time.Now(),
			DeletedAt: gorm.DeletedAt{
				Time:  time.Time{},
				Valid: true,
			},
			UpdatedAt: time.Now(),
		},
		Name:     "testname",
		Email:    "test@com",
		Password: "$2a$10$0EGP2OywIngzJKu.GoKS8eG/08tGSbZi5sMbDoJ..nWVgvQQlaDcC",
	}

	t.Run("Login resolver success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		userRow := sqlmock.
			NewRows([]string{"id", "name", "email", "password", "created_at", "deleted_at", "updated_at"}).
			AddRow(u.ID, u.Name, u.Email, u.Password, u.CreatedAt, u.DeletedAt, u.UpdatedAt)

		const userQuery = `SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(u.Email).WillReturnRows(userRow)

		var resp struct {
			Login struct {
				AccessToken  string
				RefreshToken string
			}
		}
		c.MustPost(`mutation Login {
			login(
			  email: "test@com",
			  password: "password123",
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		assert.True(t, token.Validate(resp.Login.AccessToken, ACCESS_SECRET))
		assert.True(t, token.Validate(resp.Login.RefreshToken, REFRESH_SECRET))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Login resolver wrong password", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		rows := sqlmock.
			NewRows([]string{"id", "name", "email", "password", "created_at", "deleted_at", "updated_at"}).
			AddRow(u.ID, u.Name, u.Email, u.Password, u.CreatedAt, u.DeletedAt, u.UpdatedAt)

		const userQuery = `SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(u.Email).WillReturnRows(rows)

		var resp struct {
			Login struct {
				Message string
			}
		}
		err = c.Post(`mutation Login {
			login(
			  email: "test@com",
			  password: "NOTCORRECTHEHEHE",
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Incorrect Password\",\"path\":[\"login\"]}]")

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Login resolver not an email", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		// empty response struct since we know we are going to return an error
		var resp struct{}
		err = c.Post(`mutation Login {
			login(
			  email: "this_is_def_not_an_email_WTFFFFF",
			  password: "password123",
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Not a valid email\",\"path\":[\"login\"]}]")

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Signup resolver success", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		nullUser := sqlmock.
			NewRows([]string{"id", "name", "email", "password", "created_at", "deleted_at", "updated_at"}).
			AddRow(0, "", "", "", time.Time{}, time.Time{}, time.Time{})

		const userQuery = `SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(u.Email).WillReturnRows(nullUser)

		mock.ExpectBegin()
		const createQuery = `INSERT INTO "users" ("created_at","updated_at","deleted_at","name","email","password") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
		mock.ExpectQuery(regexp.QuoteMeta(createQuery)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), u.Name, u.Email, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(u.ID))
		mock.ExpectCommit()

		var resp struct {
			Signup struct {
				AccessToken  string
				RefreshToken string
			}
		}
		c.MustPost(`mutation Signup{
			signup(
			  email: "test@com",
			  name: "testname",
			  password: "password123",
			  confirmPassword: "password123"
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)

		assert.True(t, token.Validate(resp.Signup.AccessToken, ACCESS_SECRET))
		assert.True(t, token.Validate(resp.Signup.RefreshToken, REFRESH_SECRET))

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Signup resolver with email already exists", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		userRow := sqlmock.
			NewRows([]string{"id", "name", "email", "password", "created_at", "deleted_at", "updated_at"}).
			AddRow(u.ID, u.Name, u.Email, u.Password, u.CreatedAt, u.DeletedAt, u.UpdatedAt)
		const userQuery = `SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(u.Email).WillReturnRows(userRow)

		// empty struct since we not use it
		var resp struct{}
		err := c.Post(`mutation Signup{
			signup(
			  email: "test@com",
			  name: "testname",
			  password: "password123",
			  confirmPassword: "password123"
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Email already exists\",\"path\":[\"signup\"]}]")
	})

	t.Run("Signup resolver with invalid email", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		// empty response struct since we know we are going to return an error
		var resp struct{}
		err = c.Post(`mutation Signup{
			signup(
			  email: "@notanemail:)",
			  name: "testname",
			  password: "password123",
			  confirmPassword: "password123"
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Not a valid email\",\"path\":[\"signup\"]}]")

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Signup resolver with confirm not match password", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		// empty response struct since we know we are going to return an error
		var resp struct{}
		err = c.Post(`mutation Signup{
			signup(
			  email: "test@com",
			  name: "testname",
			  password: "NOPE",
			  confirmPassword: "password123"
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Passwords don't match\",\"path\":[\"signup\"]}]")

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Signup resolver weak password no number", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		// empty response struct since we know we are going to return an error
		var resp struct{}
		err = c.Post(`mutation Signup{
			signup(
			  email: "test@com",
			  name: "testname",
			  password: "passwords",
			  confirmPassword: "passwords"
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Password needs at least 1 number and 8 - 16 characters\",\"path\":[\"signup\"]}]")

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Signup resolver weak password length", func(t *testing.T) {
		mock, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		// empty response struct since we know we are going to return an error
		var resp struct{}
		err = c.Post(`mutation Signup{
			signup(
			  email: "test@com",
			  name: "testname",
			  password: "bowo",
			  confirmPassword: "bowo"
			) {
			  ... on AuthSuccess {
				refreshToken,
				accessToken
			  }
			}
		  }`,
			&resp)
		require.EqualError(t, err, "[{\"message\":\"Password needs at least 1 number and 8 - 16 characters\",\"path\":[\"signup\"]}]")

		err = mock.ExpectationsWereMet() // make sure all expectations were met
		if err != nil {
			panic(err)
		}
	})

	t.Run("Refresh resolver refreshes access token", func(t *testing.T) {
		_, gormDB := SetupMockDB()
		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
			DB: gormDB,
		}})))

		cred := &token.Credentials{
			ID:    12,
			Name:  "testname",
			Email: "test@com",
		}

		refreshToken := token.Sign(cred, REFRESH_SECRET, 5)

		// send request and get back refresh token
		var resp struct {
			RefreshAccessToken struct {
				AccessToken string
			}
		}
		refreshAccessTokenMutation := fmt.Sprintf(`
		mutation RefreshAccessToken {
			refreshAccessToken(
			  refreshToken: "Bearer %s",
			) {
				  accessToken
			}
		  }`, refreshToken)
		c.MustPost(refreshAccessTokenMutation, &resp)
	})
}
