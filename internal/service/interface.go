package service

import (
	"easymail/internal/easylog"
)

// Manager Interface
type Manager interface {
	SetLogger(logger *easylog.Logger) error
	Start() error
	Stop() error
	Name() string
}
