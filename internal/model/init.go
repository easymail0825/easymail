package model

import (
	"easymail/internal/database"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	db = database.GetDB()
}
