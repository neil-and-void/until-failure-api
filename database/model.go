package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name               string           `gorm:"not null;type:varchar(50)"`
	Email              string           `gorm:"unique;not null;type:varchar(80)"`
	Password           string           `gorm:"not null;size:type:varchar(32)"`
	WorkoutRoutines    []WorkoutRoutine `gorm:"constraint:OnDelete:CASCADE"`
	VerificationCode   string           `gorm:"unique"`
	Verified           bool             `gorm:"default:false"`
	VerificationSentAt time.Time
}

type WorkoutRoutine struct {
	gorm.Model
	Name             string            `gorm:"not null;size:32"`
	ExerciseRoutines []ExerciseRoutine `gorm:"constraint:OnDelete:CASCADE"`
	WorkoutSessions  []WorkoutSession  `gorm:"constraint:OnDelete:CASCADE"`
	Active           bool              `gorm:"default:true"`
	UserID           uint
}

type ExerciseRoutine struct {
	gorm.Model
	Name             string     `gorm:"not null;size:32"`
	Sets             uint       `gorm:"not null"`
	Reps             uint       `gorm:"not null"`
	Exercises        []Exercise `gorm:"constraint:OnDelete:CASCADE"`
	Active           bool       `gorm:"default:true"`
	WorkoutRoutineID uint
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
	Weight     float32 `gorm:"not null" sql:"type:decimal(10,2);"`
	Reps       uint    `gorm:"not null"`
	ExerciseID uint
}
