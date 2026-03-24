package bootstrap

import (
	"easymail/internal/database"
	"easymail/internal/easylog"
	"easymail/internal/platform/config"
	"fmt"
	"os"
	"path/filepath"
)

type Runtime struct {
	Config *config.Config
	Logger *easylog.Logger
}

func Start(configPath string) (*Runtime, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	if err = database.Initialize(configPath); err != nil {
		return nil, err
	}

	logFile := cfg.Raw.LogFile
	if logFile == "" {
		logFile = "easymail.log"
	}
	fh, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file failed: %w", err)
	}

	logger := easylog.NewLogger(fh, "")
	if err = easylog.SetOutputByName(filepath.Base(logFile)); err != nil {
		return nil, fmt.Errorf("set log output failed: %w", err)
	}
	logger.SetHighlighting(false)
	logger.SetRotateByDay()
	return &Runtime{
		Config: cfg,
		Logger: logger,
	}, nil
}

