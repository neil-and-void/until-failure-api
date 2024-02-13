package database

func (db UntilFailureDB) CreateExerciseRoutine(exerciseroutine *ExerciseRoutine) error {
	result := db.DB.Create(exerciseroutine)
	return result.Error
}
