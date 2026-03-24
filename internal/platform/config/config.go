package config

import (
	"easymail/internal/database"
	"fmt"
)

type Config struct {
	Raw *database.AppConfig
}

func Load(path string) (*Config, error) {
	raw, err := database.ReadAppConfig(path)
	if err != nil {
		return nil, fmt.Errorf("load app config failed: %w", err)
	}
	return &Config{Raw: raw}, nil
}

