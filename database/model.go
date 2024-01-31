package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	BaseModel struct {
		CreatedAt time.Time      `json:"CreatedAt"`
		UpdatedAt time.Time      `json:"updatedAt"`
		DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	}

	User struct {
		BaseModel
		ID       string    `gorm:"primarykey"`
		Email    string    `gorm:"unique"`
		Routines []Routine `gorm:"constraint:OnDelete:CASCADE"`
	}

	Routine struct {
		BaseModel
		ID               uuid.UUID         `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Name             string            `gorm:"not null;size:32"`
		ExerciseRoutines []ExerciseRoutine `gorm:"constraint:OnDelete:CASCADE"`
		Workouts         []Workout         `gorm:"constraint:OnDelete:CASCADE"`
		Active           bool              `gorm:"default:true"`
		UserID           string
	}

	ExerciseRoutine struct {
		BaseModel
		ID        uuid.UUID  `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Name      string     `gorm:"not null;size:32"`
		Sets      uint       `gorm:"not null"`
		Reps      uint       `gorm:"not null"`
		Exercises []Exercise `gorm:"constraint:OnDelete:CASCADE"`
		Active    bool       `gorm:"default:true"`
		RoutineID uuid.UUID
	}

	Workout struct {
		BaseModel
		ID        uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Start     time.Time `gorm:"not null"`
		End       *time.Time
		Routine   Routine
		Exercises []Exercise `gorm:"constraint:OnDelete:CASCADE"`
		RoutineID uuid.UUID
		UserID    uuid.UUID
	}

	Exercise struct {
		BaseModel
		ID                uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Workout           Workout
		ExerciseRoutine   ExerciseRoutine
		Sets              []SetEntry `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
		Notes             string     `gorm:"size:512"`
		ExerciseRoutineID uuid.UUID
		WorkoutID         uuid.UUID
	}

	SetEntry struct {
		BaseModel
		ID         uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Weight     *float32  `sql:"type:decimal(10,2);"`
		Reps       *uint
		ExerciseID uuid.UUID
	}
)
