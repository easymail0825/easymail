package deps

import (
	"errors"
	"easymail/internal/database"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrDBNotInitialized = errors.New("database not initialized")
var ErrRedisNotInitialized = errors.New("redis not initialized")

func DB() (*gorm.DB, error) {
	d := database.GetDB()
	if d == nil {
		return nil, ErrDBNotInitialized
	}
	return d, nil
}

func Redis() (*redis.Client, error) {
	rc := database.GetRedisClient()
	if rc == nil {
		return nil, ErrRedisNotInitialized
	}
	return rc, nil
}

