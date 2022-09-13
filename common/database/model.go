package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name            string `gorm:"not null"`
	Email           string `gorm:"unique;not null"`
	Password        string `gorm:"not null"`
	WorkoutRoutines []WorkoutRoutine
}

type WorkoutRoutine struct {
	gorm.Model
	Name             string `gorm:"not null"`
	ExerciseRoutines []ExerciseRoutine
	UserID           uint
}

type ExerciseRoutine struct {
	gorm.Model
	Name             string `gorm:"not null"`
	Sets             uint   `gorm:"not null"`
	Reps             uint   `gorm:"not null"`
	WorkoutRoutineID uint
}
