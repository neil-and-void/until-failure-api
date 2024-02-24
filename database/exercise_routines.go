package database

import "gorm.io/gorm/clause"

func (db UntilFailureDB) CreateExerciseRoutine(exerciseroutine *ExerciseRoutine) error {
	result := db.DB.Create(exerciseroutine)
	return result.Error
}

func (db UntilFailureDB) UpdateExericseRoutine(exerciseRoutine *ExerciseRoutine) error {
	result := db.DB.Clauses(clause.Returning{}).Model(exerciseRoutine).Updates(exerciseRoutine)
	return result.Error
}
