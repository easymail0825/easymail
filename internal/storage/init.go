package storage

import (
	"easymail/internal/database"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	service := database.NewService()
	db = service.GetDB()
}
