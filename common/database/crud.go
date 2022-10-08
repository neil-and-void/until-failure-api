package database

import (
	"gorm.io/gorm"
)

// User
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var u User
	result := db.First(&u, "email = ?", email)
	return &u, result.Error
}

func CreateWorkoutRoutine(db *gorm.DB, routine *WorkoutRoutine) (*gorm.DB) {
	result := db.Create(routine)
	return result
}

// Workout Routine
func GetWorkoutRoutines(db *gorm.DB, email string) ([]WorkoutRoutine, error) {	
	result := db.Model(&User{}).Select("workout_routines.id, workout_routines.name, workout_routines.created_at, workout_routines.updated_at, workout_routines.deleted_at").Joins("left join workout_routines on workout_routines.user_id = users.id").Where("users.email = ?", email)	
	rows, err := result.Rows()	
	if err != nil {
		return []WorkoutRoutine{}, err
	}
	defer rows.Close()

	workoutRoutines := make([]WorkoutRoutine, 0)
	for rows.Next() {
		var wr WorkoutRoutine
		db.ScanRows(rows, &wr)
		workoutRoutines = append(workoutRoutines, wr)
	}
	return workoutRoutines, nil
}

// Exercise Routine
func GetExerciseRoutines(db *gorm.DB, email string) {}
