package database

import (
	"fmt"
	"time"

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

func GetWorkoutRoutine(db *gorm.DB, workoutRoutineId string) (*WorkoutRoutine, error) {
	var wr WorkoutRoutine
	result := db.First(&wr, "id = ?", workoutRoutineId)
	return &wr, result.Error
}

// Workout Routine
func GetWorkoutRoutines(db *gorm.DB, userId string, cursor string, limit int) ([]WorkoutRoutine, error) {
	var workoutRoutines []WorkoutRoutine
	if len(cursor) == 0 {
		db = db.Where("user_id = ?", userId)
	} else {
		db = db.Where("user_id = ? AND id > ?", userId, cursor)
	}
	result := db.Order("id").Limit(limit).Find(&workoutRoutines)
	return workoutRoutines, result.Error
}

func UpdateWorkoutRoutine(db *gorm.DB, workoutRoutineId string, workoutRoutineName string, exerciseRoutines []*ExerciseRoutine) error {
	tx := db.Begin()

	if err := tx.Model(&WorkoutRoutine{}).Where("id = ?", workoutRoutineId).Update("name", workoutRoutineName).Error; err != nil {
		tx.Rollback()
		return err
	}

	// exercise routines that are not present in this array are to be deleted
	var exerciseRoutineIds []uint

	// upsert exercise routines
	for _, er := range exerciseRoutines {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"reps", "sets", "name", "active"}),
		}).Clauses(clause.Returning{}).Create(er)

		exerciseRoutineIds = append(exerciseRoutineIds, er.ID)

		if err := result.Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Where("workout_routine_id = ? AND id NOT IN ?", workoutRoutineId, exerciseRoutineIds).Delete(&ExerciseRoutine{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
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
func AddExerciseRoutine(db *gorm.DB, exerciseRoutine *ExerciseRoutine) error {
	result := db.Create(exerciseRoutine)
	return result.Error
}

func UpdateExerciseRoutine(db *gorm.DB, exerciseRoutineId string, exerciseRoutine *ExerciseRoutine) error {
	result := db.Model(exerciseRoutine).Clauses(clause.Returning{}).Where("id = ?", exerciseRoutineId).Updates(exerciseRoutine)
	return result.Error
}

func GetExerciseRoutines(db *gorm.DB, workoutRoutineId string) (*[]ExerciseRoutine, error) {
	exerciseRoutines := []ExerciseRoutine{}

	err := db.
		Where("workout_routine_id = ?", workoutRoutineId).
		Find(&exerciseRoutines).Error

	return &exerciseRoutines, err
}

func GetExerciseRoutineIdsByExercises(db *gorm.DB, exerciseIds []string) (*[]string, error) {
	exerciseRoutineIds := []string{}
	err := db.Preload("ExerciseRoutine").Model(Exercise{}).Where("id in ?", exerciseIds).Pluck("exercise_routine.id", exerciseRoutineIds).Error
	return &exerciseRoutineIds, err
}

func GetExerciseRoutinesByWorkoutRoutineId(db *gorm.DB, workoutRoutineIds []string) (*[]ExerciseRoutine, error) {
	exerciseRoutine := []ExerciseRoutine{}
	err := db.Where("workout_routine_id IN ?", workoutRoutineIds).Find(&exerciseRoutine).Error
	return &exerciseRoutine, err
}

func GetExerciseRoutine(db *gorm.DB, exerciseRoutineId string, er *ExerciseRoutine) error {
	result := db.Model(ExerciseRoutine{}).Where("id = ?", exerciseRoutineId).First(er)
	return result.Error
}

func GetExercisesById(db *gorm.DB, ids []string) (*[]Exercise, error) {
	exercise := []Exercise{}
	err := db.Preload("ExerciseRoutine").Where("id IN ?", ids).Find(&exercise).Error
	return &exercise, err
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

func GetWorkoutSession(db *gorm.DB, workoutSessionId string) (*WorkoutSession, error) {
	workoutSession := WorkoutSession{}
	err := db.Where("id = ?", workoutSessionId).First(&workoutSession).Error
	return &workoutSession, err
}

func GetUsersWorkoutSession(db *gorm.DB, workoutSessionId string, userId string) (*WorkoutSession, error) {
	workoutSession := WorkoutSession{}
	err := db.Where("id = ? AND user_id = ?", workoutSessionId, userId).First(&workoutSession).Error
	return &workoutSession, err
}

func GetWorkoutSessions(db *gorm.DB, userId string, cursor string, limit int) ([]WorkoutSession, error) {
	var workoutSessions []WorkoutSession
	if len(cursor) == 0 {
		db = db.Where("user_id = ?", userId)
	} else {
		db = db.Where("user_id = ? AND id > ?", userId, cursor)
	}
	result := db.Order("id").Limit(limit).Find(&workoutSessions)
	return workoutSessions, result.Error
}

func GetWorkoutSessionsById(db *gorm.DB, ids []string) (*[]WorkoutSession, error) {
	workoutSessions := []WorkoutSession{}
	err := db.Preload("WorkoutRoutine").Where("id IN ?", ids).Find(&workoutSessions).Error
	return &workoutSessions, err
}

func GetPreviousWorkoutSessionsByWorkoutRoutineId(db *gorm.DB, workoutRoutineIds string, before time.Time) ([]WorkoutSession, error) {
	workoutSessions := []WorkoutSession{}
	err := db.
		Preload("Exercises").
		Where("workout_routine_id IN ? AND end < ?", workoutRoutineIds, before).
		Find(&workoutSessions).Error
	return workoutSessions, err
}

func UpdateWorkoutSession(db *gorm.DB, workoutSessionId string, updatedWorkoutSession *WorkoutSession) error {
	result := db.Model(updatedWorkoutSession).Clauses(clause.Returning{}).Where("id = ?", workoutSessionId).Updates(updatedWorkoutSession)
	return result.Error
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

func AddExercise(db *gorm.DB, exercise *Exercise) error {
	result := db.Create(exercise)
	return result.Error
}

func GetExercise(db *gorm.DB, exercise *Exercise, preloadSets bool) error {
	if preloadSets {
		db = db.Preload("Sets")
	}
	result := db.First(exercise)
	return result.Error
}

func GetExercises(db *gorm.DB, exercises *[]Exercise, workoutSessionId string) error {
	result := db.Where("workout_session_id = ?", workoutSessionId).Find(&exercises)
	return result.Error
}

func GetPrevExercisesByWorkoutRoutineId(db *gorm.DB, workoutRoutineId string, before time.Time) ([]Exercise, error) {
	exercises := []Exercise{}
	err := db.Raw(`
		SELECT * from (
			SELECT exercises.*,
				ROW_NUMBER() OVER (PARTITION BY exercises.exercise_routine_id ORDER BY workout_sessions.end DESC) AS rows
			FROM workout_sessions JOIN exercises ON exercises.workout_session_id = workout_sessions.id
			WHERE workout_sessions.end < ? and workout_sessions."end" IS NOT NULL AND workout_sessions.workout_routine_id = ?
		) TBLE where TBLE.rows = 1`,
		before, workoutRoutineId,
	).Scan(&exercises).Error
	fmt.Println("err: ", err)
	fmt.Printf("%+v", exercises)
	return exercises, err
}

func GetExercisesByWorkoutSessionId(db *gorm.DB, workoutSessionIds []string) (*[]Exercise, error) {
	exercises := []Exercise{}
	err := db.
		Where("workout_session_id IN ?", workoutSessionIds).
		Find(&exercises).Error
	return &exercises, err
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

func GetSetsByExerciseId(db *gorm.DB, exerciseIds []string) (*[]SetEntry, error) {
	setEntries := []SetEntry{}
	err := db.
		Where("exercise_id IN ?", exerciseIds).
		Find(&setEntries).Error
	return &setEntries, err
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
