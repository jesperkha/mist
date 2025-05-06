package database

func (db *Database) GetServiceByID(id uint) (Service, error) {
	var s Service
	tx := db.conn.Where("id = ?", id).First(&s)
	return s, tx.Error
}

func (db *Database) GetServiceByName(name string) (Service, error) {
	var s Service
	tx := db.conn.Where("name = ?", name).First(&s)
	return s, tx.Error
}

func (db *Database) GetAllServices() ([]Service, error) {
	var s []Service
	tx := db.conn.Find(&s)
	return s, tx.Error
}

func (db *Database) NewService(s Service) (id uint, err error) {
	tx := db.conn.Create(&s)
	return s.ID, tx.Error
}

func (db *Database) RemoveService(id uint) error {
	tx := db.conn.Where("id = ?", id).Delete(&Service{})
	return tx.Error
}
