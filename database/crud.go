package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	result := db.Model(&User{}). // todo: change to use .Preload()
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

func DeleteWorkoutRoutine(db *gorm.DB, workoutRoutineId string) error {
	tx := db.Begin()
	if err := tx.Where("id = ?", workoutRoutineId).Delete(&WorkoutRoutine{}).Error; err != nil {
		tx.Rollback()
		return err	
	}

	// Cascade exercise routines
	if err := tx.Where("workout_routine_id = ?", workoutRoutineId).Delete(&ExerciseRoutine{}).Error; err != nil {
		tx.Rollback()
		return err	
	}

	// Cascade workout sessions
	var workoutSessions []*WorkoutSession
	if err := tx.Clauses(clause.Returning{}).Where("workout_routine_id = ?", workoutRoutineId).Delete(&workoutSessions).Error; err != nil {
		tx.Rollback()
		return err	
	}

	var workoutSessionIds []string
	for _, ws := range workoutSessions {
		workoutSessionIds = append(workoutSessionIds, fmt.Sprintf("%d", ws.ID))
	} 

	// Cascade exercises
	var exercises []*Exercise
	if err := tx.Clauses(clause.Returning{}).Where("workout_session_id IN ?", workoutSessionIds).Delete(&exercises).Error; err != nil {
		tx.Rollback()
		return err	
	}
	var exerciseIds []string
	for _, e := range exercises {
		exerciseIds = append(exerciseIds, fmt.Sprintf("%d", e.ID))
	}

	// Cascade sets
	if err := tx.Where("exercise_id IN ?", exerciseIds).Delete(&SetEntry{}).Error; err != nil {
		tx.Rollback()
		return err	
	}

	return tx.Commit().Error	
}

// Exercise Routine
func GetExerciseRoutines(db *gorm.DB, workoutRoutineId string) ([]ExerciseRoutine, error) {
	result := db.Model(&WorkoutRoutine{}). // todo: change to use .Preload()
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

func GetExerciseRoutine(db *gorm.DB, exerciseRoutineId string, er *ExerciseRoutine) error {
	result := db.Model(ExerciseRoutine{}).Where("id = ?", exerciseRoutineId).First(er)
	return result.Error
}

func DeleteExerciseRoutine(db *gorm.DB, exerciseRoutineId string) error {
	tx := db.Begin()
	if err := tx.Where("id = ?", exerciseRoutineId).Delete(&ExerciseRoutine{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Cascade exercises
	var exercises []*Exercise
	if err := tx.Clauses(clause.Returning{}).Where("exercise_routine_id = ?", exerciseRoutineId).Delete(&exercises).Error; err != nil {
		tx.Rollback()
		return err
	}
	var exerciseIds []string
	for _, e := range exercises {
		exerciseIds = append(exerciseIds, fmt.Sprintf("%d", e.ID))
	}

	// Cascade sets
	if err := tx.Where("exercise_id IN ?", exerciseIds).Delete(&SetEntry{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func AddWorkoutSession(db *gorm.DB, workout *WorkoutSession) error {
	result := db.Create(workout)
	return result.Error
}

func GetWorkoutSession(db *gorm.DB, userId string, workoutSessionId string, ws *WorkoutSession) error {
	result := db.First(ws, "user_id = ? AND id = ?", userId, workoutSessionId)
	return result.Error
}

func GetWorkoutSessions(db *gorm.DB, userId string) ([]*WorkoutSession, error) {
	var workoutSessions []*WorkoutSession
	db.Preload("Exercises.Sets").Where("user_id = ?", userId).Find(&workoutSessions)
	return workoutSessions, nil
}

func DeleteWorkoutSession(db *gorm.DB, workoutSessionId string) error {
	tx := db.Begin()
	if err := tx.Where("id = ?", workoutSessionId).Delete(&WorkoutSession{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Cascade exercises
	var exercises []*Exercise
	if err := tx.Clauses(clause.Returning{}).Where("workout_session_id = ?", workoutSessionId).Delete(&exercises).Error; err != nil {
		tx.Rollback()
		return err
	}
	var exerciseIds []string
	for _, e := range exercises {
		exerciseIds = append(exerciseIds, fmt.Sprintf("%d", e.ID))
	}

	// Cascade sets
	if err := tx.Where("exercise_id IN ?", exerciseIds).Delete(&SetEntry{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
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

func UpdateExercise(db *gorm.DB, exerciseId string, updatedExercise *Exercise) error {
	result := db.Model(updatedExercise).Clauses(clause.Returning{}).Where("id = ?", exerciseId).Updates(updatedExercise)
	return result.Error
}

func DeleteExercise(db *gorm.DB, exerciseId string) error {
	tx := db.Begin()
	if err := tx.Where("id = ?", exerciseId).Delete(&Exercise{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// cascade delete on set entry table
	if err := tx.Where("exercise_id = ?", exerciseId).Delete(&SetEntry{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func AddSet(db *gorm.DB, set *SetEntry) error {
	result := db.Create(set)
	return result.Error
}

func GetSets(db *gorm.DB, s *[]SetEntry, exerciseId string) error {
	result := db.Where("exercise_id = ?", exerciseId).Find(&s)
	return result.Error
}

func GetSet(db *gorm.DB, s *SetEntry, setId string) error {
	result := db.Where("id = ?", setId).Find(s)
	return result.Error
}

func UpdateSet(db *gorm.DB, setID string, updatedSet *SetEntry) error {
	result := db.Model(updatedSet).Clauses(clause.Returning{}).Where("id = ?", setID).Updates(updatedSet)
	return result.Error
}

func DeleteSet(db *gorm.DB, setID string) error {
	result := db.Where("id = ?", setID).Delete(&SetEntry{})
	return result.Error
}
