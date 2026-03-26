package maillog

import (
	"errors"
	"easymail/internal/database"
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

