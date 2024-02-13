package database

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	BaseModel struct {
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}

	User struct {
		BaseModel
		ID       string    `gorm:"primarykey"`
		Email    string    `gorm:"unique;not null"`
		Routines []Routine `gorm:"constraint:OnDelete:CASCADE"`
	}

	Routine struct {
		BaseModel
		ID               uuid.UUID         `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Name             string            `gorm:"not null;size:32"`
		ExerciseRoutines []ExerciseRoutine `gorm:"constraint:OnDelete:CASCADE"`
		Workouts         []Workout         `gorm:"constraint:OnDelete:CASCADE"`
		Tags             []Tag             `gorm:"many2many:routine_tags;"`
		Active           bool              `gorm:"default:true;not null"`
		Private          bool              `gorm:"default:false;not null"`
		UserID           string
	}

	ExerciseRoutine struct {
		BaseModel
		ID         uuid.UUID   `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Name       string      `gorm:"not null;size:32"`
		Active     bool        `gorm:"default:true;not null"`
		Exercises  []Exercise  `gorm:"constraint:OnDelete:CASCADE"`
		SetSchemes []SetScheme `gorm:"constraint:OnDelete:CASCADE"`
		RoutineID  uuid.UUID
	}

	SetScheme struct {
		BaseModel
		TargetReps        uint            `gorm:"not null"`
		SetType           SetType         `gorm:"default:'WORKING';not null;type:set_type"`
		Measurement       MeasurementType `gorm:"default:'WEIGHT';not null;type:measurement_type"`
		ExerciseRoutineId uuid.UUID
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
		Sets              []SetEntry `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		Notes             string     `gorm:"size:512"`
		ExerciseRoutineID uuid.UUID
		WorkoutID         uuid.UUID
	}

	SetEntry struct {
		BaseModel
		ID          uuid.UUID       `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Weight      *float32        `sql:"type:decimal(10,3);"`
		Reps        *uint           `sql:"type:decimal(10,2);"`
		Seconds     *uint           `sql:"type:decimal(10,3);"`
		SetType     SetType         `gorm:"default:'WORKING';not null;type:set_type"`
		Measurement MeasurementType `gorm:"default:'WEIGHT';not null;type:measurement_type"`
		ExerciseID  uuid.UUID
		SetSchemeID uuid.UUID
	}

	Tag struct {
		BaseModel
		ID  uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
		Tag string    `gorm:"not null;size:32"`
	}

	MeasurementType string
	SetType         string
)

const (
	Weight           MeasurementType = "WEIGHT"
	Duration         MeasurementType = "DURATION"
	BodyWeight       MeasurementType = "BODYWEIGHT"
	WeightedDuration MeasurementType = "WEIGHTED_DURATION"
)

const (
	Warmup  SetType = "WARMUP"
	Working SetType = "WORKING"
	Drop    SetType = "DROP"
	Super   SetType = "SUPER"
)

func (m *MeasurementType) Scan(value interface{}) error {
	*m = MeasurementType(value.([]byte))
	return nil
}

func (m MeasurementType) Value() (driver.Value, error) {
	return string(m), nil
}

func (s *SetType) Scan(value interface{}) error {
	*s = SetType(value.([]byte))
	return nil
}

func (s SetType) Value() (driver.Value, error) {
	return string(s), nil
}

func (SetEntry) TableName() string {
	return "set_entries"
}

func (SetScheme) TableName() string {
	return "set_scheme"
}
