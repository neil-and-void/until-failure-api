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
	WorkoutSessions  []WorkoutSession
	Active           bool `gorm:"default:true"`
	UserID           uint
}

type ExerciseRoutine struct {
	gorm.Model
	Name             string `gorm:"not null"`
	Sets             uint   `gorm:"not null"`
	Reps             uint   `gorm:"not null"`
	Exercises        []Exercise
	Active           bool `gorm:"default:true"`
	WorkoutRoutineID uint
}

type WorkoutSession struct {
	gorm.Model
	Start            time.Time `gorm:"not null"`
	End              *time.Time
	WorkoutRoutine	 WorkoutRoutine
	Exercises        []Exercise
	WorkoutRoutineID uint
	UserID           uint
}

type Exercise struct {
	gorm.Model
	WorkoutSession WorkoutSession
	ExerciseRoutine ExerciseRoutine
	Sets              []SetEntry `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Notes             string     `gorm:"size:512"`
	ExerciseRoutineID uint
	WorkoutSessionID  uint
}

type SetEntry struct {
	gorm.Model
	Weight     float32 `gorm:"not null" sql:"type:decimal(10,2);"`
	Reps       uint    `gorm:"not null"`
	ExerciseID uint
}
