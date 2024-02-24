package database

import "gorm.io/gorm/clause"

func (db UntilFailureDB) CreateSetScheme(setScheme *SetScheme) error {
	result := db.DB.Create(setScheme)
	return result.Error
}

func (db UntilFailureDB) UpdateSetScheme(setScheme *SetScheme) error {
	result := db.DB.Clauses(clause.Returning{}).Model(setScheme).Updates(setScheme)
	return result.Error
}
