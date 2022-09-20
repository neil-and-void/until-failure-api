package tests

import (
	"github.com/DATA-DOG/go-sqlmock"
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
