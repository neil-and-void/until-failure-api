package graph

import (
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/neilZon/workout-logger-api/common/database"
	"github.com/neilZon/workout-logger-api/utils/token"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestSchemaResolvers(t *testing.T) {
	err := godotenv.Load("../.test.env")
	if err != nil {
		panic("Error loading .env file")
	}
	ACCESS_SECRET := []byte(os.Getenv("ACCESS_SECRET"))
	REFRESH_SECRET := []byte(os.Getenv("REFRESH_SECRET"))

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
		Email:    "test@test.com",
		Password: "$2a$10$0EGP2OywIngzJKu.GoKS8eG/08tGSbZi5sMbDoJ..nWVgvQQlaDcC",
	}

	t.Run("Login resolver success", func(t *testing.T) {
		mockDb, mock, err := sqlmock.New() // mock sql.DB
		if err != nil {
			panic(err)
		}

		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: mockDb,
		}), &gorm.Config{})
		defer mockDb.Close()

		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &Resolver{
			DB: gormDB,
		}})))

		rows := sqlmock.
			NewRows([]string{"id", "name", "email", "password", "created_at", "deleted_at", "updated_at"}).
			AddRow(u.ID, u.Name, u.Email, u.Password, u.CreatedAt, u.DeletedAt, u.UpdatedAt)

		const userQuery = `SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`
		mock.ExpectQuery(regexp.QuoteMeta(userQuery)).WithArgs(u.Email).WillReturnRows(rows)

		var resp struct {
			Login struct {
				AccessToken  string
				RefreshToken string
			}
		}
		c.MustPost(`mutation Login {
			login(
			  email: "test@test.com",
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

	t.Run("Login resolver wrong password", func(t *testing.T) {})

	t.Run("Login resolver not an email", func(t *testing.T) {})

	t.Run("Signup resolver success", func(t *testing.T) {
		mockDb, mock, err := sqlmock.New() // mock sql.DB
		if err != nil {
			panic(err)
		}

		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: mockDb,
		}), &gorm.Config{})
		defer mockDb.Close()

		c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &Resolver{
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
			  email: "test@test.com",
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

	t.Run("Signup resolver with email already exists", func(t *testing.T) {})
	t.Run("Signup resolver with non-email", func(t *testing.T) {})
	t.Run("Signup resolver with confirm not match password", func(t *testing.T) {})

	// t.Run("Refresh resolver", func(t *testing.T) {})
}
