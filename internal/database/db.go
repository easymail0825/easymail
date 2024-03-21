package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局变量

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func init() {
	var err error
	dsn := "easymail:easymail@tcp(localhost:3306)/easymail?charset=utf8&parseTime=True&loc=Local"
	DB, err = InitDB(dsn)
	if err != nil {
		panic("failed to connect to database")
	}
}
