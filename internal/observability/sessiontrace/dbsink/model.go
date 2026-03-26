package dbsink

import (
	"time"
)

type SessionEvent struct {
	ID        uint      `gorm:"autoIncrement;primaryKey"`
	TS        time.Time `gorm:"index"`
	SessionID string    `gorm:"type:varchar(64);index"`
	Protocol  string    `gorm:"type:varchar(32);index"`
	Stage     string    `gorm:"type:varchar(64);index"`
	Remote    string    `gorm:"type:varchar(128)"`
	Local     string    `gorm:"type:varchar(128)"`
	Duration  int64     `gorm:"type:bigint"` // nanoseconds
	Err       string    `gorm:"type:text"`
	Fields    string    `gorm:"type:longtext"`
	Tags      string    `gorm:"type:longtext"`
}

