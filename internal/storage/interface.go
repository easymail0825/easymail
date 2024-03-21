package storage

import (
	"easymail/internal/account"
	"io"
)

// Storager defines the interface for a storage provider.
type Storager interface {
	// Save writes the content of the reader to the specified file.
	// The identify parameter is used to determine the file name or id.
	// Email must have account id
	Save(accountName string, email *account.Email, content io.Reader) (identify string, err error)
	Query(accID int64, page int, pageSize int) ([]account.Email, error)
}
