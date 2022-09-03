package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name         string `gorm:"not null"`
	Email        string `gorm:"unique;not null"`
	WorkoutRoutines []WorkoutRoutine
}

type WorkoutRoutine struct {
	gorm.Model
	Name         string `gorm:"not null"`
	ExerciseRoutines []ExerciseRoutine
}

type ExerciseRoutine struct {
	gorm.Model
	Name         string `gorm:"not null"`
	Sets uint
	Reps uint
}
