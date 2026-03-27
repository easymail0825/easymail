package model

import "gorm.io/gorm"

// AutoMigrate is an opt-in bootstrap helper for first-run environments.
// It should only be called by bootstrap/database initialization, never from init().
func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return gorm.ErrInvalidDB
	}
	return db.AutoMigrate(
		&Configure{},
		&Domain{},
		&Account{},
		&Admin{},
		&Email{},
		&SsdeepHash{},
		&FilterRule{},
		&FilterLog{},
		&FilterField{},
		&FilterMetric{},
	)
}
