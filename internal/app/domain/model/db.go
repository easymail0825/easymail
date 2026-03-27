package model

import (
	"easymail/internal/database"
	"errors"
	"gorm.io/gorm"
)

var ErrDBNotInitialized = errors.New("database not initialized")

func getDB() (*gorm.DB, error) {
	d := database.GetDB()
	if d == nil {
		return nil, ErrDBNotInitialized
	}
	return d, nil
}

// DB exposes the initialized DB handle for modules that are not allowed
// to import internal/database directly (architecture boundary enforcement).
func DB() (*gorm.DB, error) {
	return getDB()
}
