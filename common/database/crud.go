package database

import "gorm.io/gorm"

// User
func GetUserByEmail(db *gorm.DB, email string) (User, error) {
	var user User
	result := db.Where("email = ?", email, &user)
	return user, result.Error
}

// Workout Routine
func GetWorkoutRoutines(db *gorm.DB, email string) {}

// Exercise Routine
func GetExerciseRoutines(db *gorm.DB, email string) {}
