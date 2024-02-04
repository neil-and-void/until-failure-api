package database

func (db UntilFailureDB) CreateRoutine(routine *Routine) error {
	result := db.DB.Create(routine)
	return result.Error
}
