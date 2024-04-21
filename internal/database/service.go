package database

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sync"
)

type Service struct {
	lock        *sync.Mutex
	DB          *gorm.DB
	RedisClient *redis.Client
}

func NewService() *Service {
	return &Service{
		lock:        &sync.Mutex{},
		DB:          db,
		RedisClient: GetRedisClient(),
	}
}

func (s *Service) GetDB() *gorm.DB {
	return db
}
