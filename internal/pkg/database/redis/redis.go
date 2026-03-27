package database

import (
	"github.com/redis/go-redis/v9"
)

var rc *redis.Client // 全局变量

func initRedis(dsn string) {
	rc = redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: "",
		DB:       0,
	})
}

func GetRedisClient() *redis.Client {
	return rc
}
