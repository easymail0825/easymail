package storage

import (
	"easymail/internal/account"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LocalStorage is a Storager implementation that saves files locally.
type LocalStorage struct {
	BaseDir string
	lock    sync.Mutex
}

func (s *LocalStorage) Query(accID int64, page int, pageSize int) ([]account.Email, error) {
	//TODO implement me
	panic("implement me")
}

// NewLocalStorage creates a new LocalStorage instance.
func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{
		BaseDir: baseDir,
		lock:    sync.Mutex{},
	}
}

// Save saves the content of the file at the given filename.
func (s *LocalStorage) Save(accountName string, email *account.Email, content io.Reader) (identify string, err error) {
	// check account status
	if len(accountName) < 1 {
		return "", errors.New("account name is empty")
	}
	var acc *account.Account
	acc, err = account.FindAccountByName(accountName)
	if err != nil {
		return "", errors.New("account not found")
	}
	if !account.ValidateAccount(*acc) {
		return "", errors.New("account is disabled")
	}

	// check domain status
	var domain *account.Domain
	domain, err = account.FindDomainByID(acc.DomainID)
	if err != nil {
		return "", errors.New("domain not found")
	}
	if !account.ValidateDomain(*domain) {
		return "", errors.New("domain is disabled")
	}

	// postfix queue-id must exists
	if len(email.JobID) < 8 || len(email.JobID) > 12 {
		return "", errors.New("email job id is invalid")
	}

	// generate identify,
	// domain/account/jobid_first_char/jobid_second_char/jobid
	filePath := filepath.Join(s.BaseDir, fmt.Sprintf("%s/%s/%s/%s", domain.Name, acc.Username,
		string(email.JobID[0]), string(email.JobID[2])))

	// make directory if not exists
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return "", err
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	// save file
	filePath = filepath.Join(filePath, email.JobID)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	if err != nil {
		return "", err
	}
	email.Identity = filePath
	email.AccountID = acc.ID
	email.SaveTime = time.Now()

	// update mail database
	err = acc.SaveEmail(email)
	if err != nil {
		return "", err
	}
	return filePath, nil
}
