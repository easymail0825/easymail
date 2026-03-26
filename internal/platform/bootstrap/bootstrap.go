package bootstrap

import (
	"easymail/internal/database"
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/observability/sessiontrace"
	"easymail/internal/observability/sessiontrace/dbsink"
	"easymail/internal/observability/sessiontrace/filejsonl"
	"easymail/internal/platform/config"
	"fmt"
	"os"
	"path/filepath"
)

type Runtime struct {
	Config *config.Config
	Logger *easylog.Logger

	// Resolved values from persistent configuration store.
	StorageRoot string
	StorageData string

	Tracer sessiontrace.Tracer
}

func Start(configPath string) (*Runtime, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	if err = database.Initialize(configPath); err != nil {
		return nil, err
	}
	if cfg.Raw.InitDB {
		if err := model.AutoMigrate(database.GetDB()); err != nil {
			return nil, fmt.Errorf("init_db auto migrate failed: %w", err)
		}
		// minimal defaults for first-run (prevents lmtp failing on empty config store)
		_ = model.SeedDefaults("./storage", "./data")
	}
	easylog.InitializeDefaults()

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

	// session tracer (optional)
	var tracer sessiontrace.Tracer = sessiontrace.NewNoop()
	if cfg.Raw.Observability.SessionTrace.Enabled {
		sinks := make([]sessiontrace.Sink, 0, 2)
		switch cfg.Raw.Observability.SessionTrace.Sink {
		case "db":
			s, err := dbsink.New(dbsink.Config{Enabled: true})
			if err == nil && s != nil {
				sinks = append(sinks, s)
			}
		case "file", "":
			s, err := filejsonl.New(filejsonl.Config{
				Path:        cfg.Raw.Observability.SessionTrace.FilePath,
				RotateDaily: true,
			})
			if err == nil && s != nil {
				sinks = append(sinks, s)
			}
		}
		tracer = sessiontrace.NewAsync(sessiontrace.Config{
			Enabled:   len(sinks) > 0,
			QueueSize: cfg.Raw.Observability.SessionTrace.QueueSize,
		}, sinks...)
	}

	storageData, storageRoot := "", ""
	if c, err := model.GetConfigureByNames("easymail", "storage", "data"); err == nil && c != nil {
		storageData = c.Value
	}
	if r, err := model.GetConfigureByNames("easymail", "storage", "root"); err == nil && r != nil {
		storageRoot = r.Value
	}
	if storageRoot == "" && cfg.Raw.InitDB {
		storageRoot = "./storage"
	}
	if storageData == "" && cfg.Raw.InitDB {
		storageData = "./data"
	}
	return &Runtime{
		Config:      cfg,
		Logger:      logger,
		StorageRoot: storageRoot,
		StorageData: storageData,
		Tracer:      tracer,
	}, nil
}
