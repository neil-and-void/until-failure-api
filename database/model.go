package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
}

type WorkoutRoutine struct {
	gorm.Model
	Name             string            `gorm:"not null;size:32" json:"name"`
	ExerciseRoutines []ExerciseRoutine `gorm:"constraint:OnDelete:CASCADE" json:"exerciseRoutines"`
	WorkoutSessions  []WorkoutSession  `gorm:"constraint:OnDelete:CASCADE" json:"workoutSessions"`
	Active           bool              `gorm:"default:true" json:"active"`
	UserID           uint              `json:"userId"`
}

type ExerciseRoutine struct {
	gorm.Model
	Name             string     `gorm:"not null;size:32" json:"name"`
	Sets             uint       `gorm:"not null" json:"sets"`
	Reps             uint       `gorm:"not null" reps:"reps"`
	Exercises        []Exercise `gorm:"constraint:OnDelete:CASCADE" json:"exercises"`
	Active           bool       `gorm:"default:true" json:"active"`
	WorkoutRoutineID uint       `json:"workoutRoutineId"`
}

type WorkoutSession struct {
	gorm.Model
	Start            time.Time `gorm:"not null"`
	End              *time.Time
	WorkoutRoutine   WorkoutRoutine
	Exercises        []Exercise `gorm:"constraint:OnDelete:CASCADE"`
	WorkoutRoutineID uint
	UserID           uint
}

type Exercise struct {
	gorm.Model
	WorkoutSession    WorkoutSession
	ExerciseRoutine   ExerciseRoutine
	Sets              []SetEntry `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Notes             string     `gorm:"size:512"`
	ExerciseRoutineID uint
	WorkoutSessionID  uint
}

type SetEntry struct {
	gorm.Model
	Weight     float32 `sql:"type:decimal(10,2);"`
	Reps       uint
	ExerciseID uint
}
