package database

import (
	"github.com/neilZon/workout-logger-api/utils/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDb() (*gorm.DB, error) {
	databaseUrl := config.GetEnvVariable("DATABASE_URL")

	var err error
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(User{}, WorkoutRoutine{}, ExerciseRoutine{})
	return db, nil
}
