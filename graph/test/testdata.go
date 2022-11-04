package test

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/neilZon/workout-logger-api/config"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/token"
	"gorm.io/gorm"
)

var User = &token.Claims{
	Name: "test",
	ID:   28,
	StandardClaims: jwt.StandardClaims{
		ExpiresAt: time.Now().Add(config.ACCESS_TTL * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
		Issuer:    "neil:)",
		Subject:   "test@test.com",
	},
}

var WorkoutRoutine = &database.WorkoutRoutine{
	Name: "Legs",
	ExerciseRoutines: []database.ExerciseRoutine{
		{
			Model: gorm.Model{
				ID:        3,
				CreatedAt: time.Now(),
				DeletedAt: gorm.DeletedAt{
					Time:  time.Time{},
					Valid: true,
				},
				UpdatedAt: time.Now(),
			},
			Name:             "squat",
			Sets:             4,
			Reps:             6,
			WorkoutRoutineID: 8,
		},
		{
			Model: gorm.Model{
				ID:        4,
				CreatedAt: time.Now(),
				DeletedAt: gorm.DeletedAt{
					Time:  time.Time{},
					Valid: true,
				},
				UpdatedAt: time.Now(),
			},
			Name:             "leg extensions",
			Sets:             4,
			Reps:             6,
			WorkoutRoutineID: 8,
		},
	},
	UserID: 28,
	Model: gorm.Model{
		ID:        8,
		CreatedAt: time.Now(),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Time{},
			Valid: true,
		},
		UpdatedAt: time.Now(),
	},
}

var noteOne = "This is a note"
var noteTwo = "This is another note"

var WorkoutSession = &database.WorkoutSession{
	WorkoutRoutineID: 8,
	UserID:           28,
	Start:            time.Date(2022, time.October, 30, 12, 34, 0, 0, time.UTC),
	Model: gorm.Model{
		ID:        3,
		CreatedAt: time.Now(),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Time{},
			Valid: true,
		},
		UpdatedAt: time.Now(),
	},
	Exercises: []database.Exercise{
		{
			Model: gorm.Model{
				ID:        44,
				CreatedAt: time.Now(),
				DeletedAt: gorm.DeletedAt{
					Time:  time.Time{},
					Valid: true,
				},
				UpdatedAt: time.Now(),
			},
			Notes: &noteOne,
			WorkoutSessionID:  3,
			ExerciseRoutineID: 3,
			Sets: []database.SetEntry{
				{
					Model: gorm.Model{
						ID:        30,
						CreatedAt: time.Now(),
						DeletedAt: gorm.DeletedAt{
							Time:  time.Time{},
							Valid: true,
						},
						UpdatedAt: time.Now(),
					},
					Weight:     float32(225),
					Reps:       uint(8),
					ExerciseID: 44,
				},
				{
					Model: gorm.Model{
						ID:        31,
						CreatedAt: time.Now(),
						DeletedAt: gorm.DeletedAt{
							Time:  time.Time{},
							Valid: true,
						},
						UpdatedAt: time.Now(),
					},
					Weight:     float32(225),
					Reps:       uint(7),
					ExerciseID: 44,
				},
			},
		},
		{
			Model: gorm.Model{
				ID:        45,
				CreatedAt: time.Now(),
				DeletedAt: gorm.DeletedAt{
					Time:  time.Time{},
					Valid: true,
				},
				UpdatedAt: time.Now(),
			},
			WorkoutSessionID:  3,
			ExerciseRoutineID: 4,
			Notes: &noteTwo,
			Sets: []database.SetEntry{
				{
					Model: gorm.Model{
						ID:        32,
						CreatedAt: time.Now(),
						DeletedAt: gorm.DeletedAt{
							Time:  time.Time{},
							Valid: true,
						},
						UpdatedAt: time.Now(),
					},
					Weight:     float32(225),
					Reps:       uint(8),
					ExerciseID: 45,
				},
				{
					Model: gorm.Model{
						ID:        33,
						CreatedAt: time.Now(),
						DeletedAt: gorm.DeletedAt{
							Time:  time.Time{},
							Valid: true,
						},
						UpdatedAt: time.Now(),
					},
					Weight:     float32(225),
					Reps:       uint(7),
					ExerciseID: 45,
				},
			},
		},
	},
}
