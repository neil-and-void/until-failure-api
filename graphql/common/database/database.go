package db

import (
	"github.com/neilZon/workout-logger-api/graphql/common/database/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dsn = "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
)

func InitDb() (*gorm.DB, error) {
    var err error
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, err
    }
    db.AutoMigrate(&model.User{}, &model.WorkoutRoutine{}, &model.ExerciseRoutine{})
    return db, nil
}
