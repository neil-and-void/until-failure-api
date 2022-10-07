package database

import (
	"gorm.io/gorm"
)

// User
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	result := db.First(&user, "email = ?", email)
	return &user, result.Error
}

func CreateWorkoutRoutine(db *gorm.DB, routine *WorkoutRoutine) (*gorm.DB) {
	result := db.Create(routine)
	return result
}

// Workout Routine
func GetWorkoutRoutines(db *gorm.DB, email string) {}

// Exercise Routine
func GetExerciseRoutines(db *gorm.DB, email string) {}
