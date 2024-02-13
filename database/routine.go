package database

func (db UntilFailureDB) CreateRoutine(routine *Routine) error {
	result := db.DB.Create(routine)
	return result.Error
}

func (db UntilFailureDB) GetRoutines(userId string) ([]Routine, error) {
	routines := []Routine{}
	result := db.DB.Where("user_id = ?", userId).Order("created_at desc").Find(&routines)
	return routines, result.Error
}

func (db UntilFailureDB) GetRoutine(routineId string) (Routine, error) {
	routine := Routine{}
	result := db.DB.Preload("ExerciseRoutines").Where("id = ?", routineId).First(&routine)
	return routine, result.Error
}
