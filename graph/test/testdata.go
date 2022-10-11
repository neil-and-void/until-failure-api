package test

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/neilZon/workout-logger-api/common/database"
	"github.com/neilZon/workout-logger-api/utils/config"
	"github.com/neilZon/workout-logger-api/utils/token"
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
