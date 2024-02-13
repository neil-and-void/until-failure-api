package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UntilFailureDB struct {
	DB *gorm.DB
}

func InitDb() (UntilFailureDB, error) {
	DB_HOST := os.Getenv("DB_HOST")
	DB_DBNAME := os.Getenv("DB_DBNAME")
	DB_USERNAME := os.Getenv("DB_USERNAME")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT := os.Getenv("DB_PORT")
	DSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", DB_HOST, DB_USERNAME, DB_PASSWORD, DB_DBNAME, DB_PORT)

	var err error
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  DSN,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		return UntilFailureDB{}, err
	}

	db.AutoMigrate(User{}, Routine{}, ExerciseRoutine{}, Workout{}, Exercise{}, SetEntry{}, SetScheme{}, Tag{})

	untilFailureDB := UntilFailureDB{DB: db}

	return untilFailureDB, nil
}
