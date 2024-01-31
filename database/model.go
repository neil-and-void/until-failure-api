package database

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt time.Time      `json:"CreatedAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

type User struct {
	BaseModel
	ID    string `gorm:"primarykey"`
	Email string `gorm:"unique"`
}

type Routine struct {
	BaseModel
	Name             string            `gorm:"not null;size:32"`
	ExerciseRoutines []ExerciseRoutine `gorm:"constraint:OnDelete:CASCADE"`
	Workouts         []Workout         `gorm:"constraint:OnDelete:CASCADE"`
	Active           bool              `gorm:"default:true"`
	UserID           uint              `json:"userId"`
}

type ExerciseRoutine struct {
	BaseModel
	Name      string     `gorm:"not null;size:32"`
	Sets      uint       `gorm:"not null"`
	Reps      uint       `gorm:"not null"`
	Exercises []Exercise `gorm:"constraint:OnDelete:CASCADE"`
	Active    bool       `gorm:"default:true"`
	RoutineID uint
}

type Workout struct {
	BaseModel
	Start     time.Time `gorm:"not null"`
	End       *time.Time
	Routine   Routine
	Exercises []Exercise `gorm:"constraint:OnDelete:CASCADE"`
	RoutineID uint
	UserID    uint
}

type Exercise struct {
	BaseModel
	Workout           Workout
	ExerciseRoutine   ExerciseRoutine
	Sets              []SetEntry `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Notes             string     `gorm:"size:512"`
	ExerciseRoutineID uint
	SessionID         uint
}

type SetEntry struct {
	BaseModel
	Weight     *float32 `sql:"type:decimal(10,2);"`
	Reps       *uint
	ExerciseID uint
}
