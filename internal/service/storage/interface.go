package storage

import (
	"easymail/internal/model"
	"io"
)

// Storager defines the interface for a storage provider.
type Storager interface {
	// Save writes the content of the reader to the specified file.
	// The identify parameter is used to determine the file name or id.
	// Email must have model id
	Save(accountName string, email *model.Email, content io.Reader) (identify string, err error)
}
