package database

import (
	"fmt"
	"sync"
)

var initOnce sync.Once
var initErr error

func Initialize(configPath string) error {
	initOnce.Do(func() {
		appConfig, err := ReadAppConfig(configPath)
		if err != nil {
			initErr = fmt.Errorf("read app config failed: %w", err)
			return
		}
		// initialize database first
		if err = initMySQL(appConfig.Mysql); err != nil {
			initErr = fmt.Errorf("connect mysql failed: %w", err)
			return
		}
		initRedis(appConfig.Redis)
	})
	return initErr
}

func init() {
	_ = Initialize("easymail.yaml")
}
