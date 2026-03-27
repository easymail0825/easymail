package dbsink

import (
	"context"
	"easymail/internal/observability/sessiontrace"
	"easymail/internal/runtime/deps"
	"encoding/json"
	"gorm.io/gorm"
)

type Config struct {
	Enabled bool
}

type Sink struct {
	db *gorm.DB
}

func New(cfg Config) (*Sink, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	db, err := deps.DB()
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&SessionEvent{}); err != nil {
		return nil, err
	}
	return &Sink{db: db}, nil
}

func (s *Sink) Write(ctx context.Context, e sessiontrace.Event) error {
	if s == nil || s.db == nil {
		return nil
	}
	fields, _ := json.Marshal(e.Fields)
	tags, _ := json.Marshal(e.Tags)
	row := &SessionEvent{
		TS:        e.Timestamp,
		SessionID: e.SessionID,
		Protocol:  string(e.Protocol),
		Stage:     e.Stage,
		Remote:    e.Remote,
		Local:     e.Local,
		Duration:  int64(e.Duration),
		Err:       e.Err,
		Fields:    string(fields),
		Tags:      string(tags),
	}
	return s.db.WithContext(ctx).Create(row).Error
}

func (s *Sink) Close() error { return nil }
