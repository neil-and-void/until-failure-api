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

func CreateWorkoutRoutine(db *gorm.DB, routine *WorkoutRoutine) *gorm.DB {
	result := db.Create(routine)
	return result
}

func GetWorkoutRoutine(db *gorm.DB, userId string, workoutRoutineId string) (*WorkoutRoutine, error) {
	var wr WorkoutRoutine
	result := db.First(&wr, "user_id = ? AND id = ?", userId, workoutRoutineId)
	return &wr, result.Error
}

// Workout Routine
func GetWorkoutRoutines(db *gorm.DB, email string) ([]WorkoutRoutine, error) {
	result := db.Model(&User{}).
		Select("workout_routines.id, workout_routines.name, workout_routines.created_at, workout_routines.updated_at, workout_routines.deleted_at").
		Joins("left join workout_routines on workout_routines.user_id = users.id").
		Where("users.email = ?", email)
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
func GetExerciseRoutines(db *gorm.DB, workoutRoutineId string) ([]ExerciseRoutine, error) {
	result := db.Model(&WorkoutRoutine{}).
		Select("exercise_routines.id, exercise_routines.name, exercise_routines.sets, exercise_routines.reps, exercise_routines.created_at, exercise_routines.updated_at, exercise_routines.deleted_at").
		Joins("left join exercise_routines on workout_routines.id = exercise_routines.workout_routine_id").
		Where("exercise_routines.workout_routine_id = ?", workoutRoutineId)
	rows, err := result.Rows()
	if err != nil {
		return []ExerciseRoutine{}, err
	}
	defer rows.Close()

	exerciseRoutines := make([]ExerciseRoutine, 0)
	for rows.Next() {
		var er ExerciseRoutine
		db.ScanRows(rows, &er)
		exerciseRoutines = append(exerciseRoutines, er)
	}
	return exerciseRoutines, nil
}

func AddWorkoutSession(db *gorm.DB, workout *WorkoutSession) error {
	result := db.Create(workout)
	return result.Error
}

func GetWorkoutSession(db *gorm.DB, userId string, workoutSessionId string, ws *WorkoutSession) ( error) {
	result := db.First(ws, "user_id = ? AND id = ?", userId, workoutSessionId)
	return result.Error
}

func GetWorkoutSessions(db *gorm.DB, userId string) ([]*WorkoutSession, error) {
	var workoutSessions []*WorkoutSession
	db.Preload("Exercises.Sets").Where("user_id = ?", userId).Find(&workoutSessions)
	return workoutSessions, nil
}

func AddExercise(db *gorm.DB, exercise *Exercise, workoutSessionId string) error {
	result := db.Create(exercise)
	return result.Error
}

func GetExercise(db *gorm.DB, exercise *Exercise) error {
	result := db.Preload("Sets").First(exercise)
	return result.Error
}

func GetExercises(db *gorm.DB, exercises *[]Exercise, workoutSessionId string) error {
	result := db.Preload("Sets").Where("workout_session_id = ?", workoutSessionId).Find(&exercises)
	return result.Error
}

func AddSet(db *gorm.DB, set *SetEntry) error {
	result := db.Create(set)
	return result.Error
}

func GetSets(db *gorm.DB, s *[]SetEntry, exerciseId string) error {
	result := db.Where("exercise_id = ?", exerciseId).Find(&s)
	return result.Error
}

func UpdateSet(db *gorm.DB, set *SetEntry) error {
	return nil
}
