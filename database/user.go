package database

func (db UntilFailureDB) CreateUser(user User) error {
	result := db.DB.Create(&user)
	return result.Error
}

func (db UntilFailureDB) GetUser(id string) (User, error) {
	u := User{}
	result := db.DB.First(&u, "id = ?", id)
	if result.Error != nil {
		return u, result.Error
	}
	return u, nil
}
