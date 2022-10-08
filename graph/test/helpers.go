package test

import (
	"context"

	"github.com/99designs/gqlgen/client"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/utils/token"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupMockDB() (sqlmock.Sqlmock, *gorm.DB) {
	mockDb, mock, err := sqlmock.New() // mock sql.DB
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDb,
	}), &gorm.Config{})

	return mock, gormDB
}

func AddContext(u *token.Claims) client.Option {
    return func(bd *client.Request) {
        ctx := context.WithValue(bd.HTTP.Context(), middleware.UserCtxKey, u)
        bd.HTTP = bd.HTTP.WithContext(ctx)
    }
}
