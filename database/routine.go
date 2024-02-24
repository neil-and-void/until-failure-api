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

func (db UntilFailureDB) GetRoutineCount(userId string) (uint, error) {
	var count int64
	err := db.DB.Model(&User{}).Where("name = ?", "jinzhu").Count(&count).Error
	return uint(count), err
}

func (db UntilFailureDB) GetRoutine(routineId string) (Routine, error) {
	routine := Routine{}
	result := db.DB.Preload("ExerciseRoutines.SetSchemes").Where("id = ?", routineId).First(&routine)
	return routine, result.Error
}
