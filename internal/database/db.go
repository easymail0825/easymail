package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB // 全局变量

func InitDB(dsn string) error {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}
