package filtering

import (
	"easymail/internal/easylog"
	oldfilter "easymail/internal/service/filter"
)

// NewServer wraps existing milter filter service for phased migration.
func NewServer(family, listen string, logger *easylog.Logger) *oldfilter.Server {
	return oldfilter.New(family, listen, logger)
}

