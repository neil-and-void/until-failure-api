package database

import (
	"time"

	"gorm.io/gorm"
)

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

type WorkoutSession struct {
	gorm.Model
	Start            time.Time `gorm:"not null"`
	End              *time.Time
	WorkoutRoutineID uint
	UserID           uint
	Exercises        []Exercise
}

type Exercise struct {
	gorm.Model
	WorkoutSessionID  uint
	ExerciseRoutineID uint
	Sets              []SetEntry
}

type SetEntry struct {
	gorm.Model
	Weight     float32 `gorm:"not null" sql:"type:decimal(10,2);"`
	Reps       uint    `gorm:"not null"`
	Notes      *string `gorm:"size:512"`
	ExerciseID uint
}
