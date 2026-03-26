package controller

import (
	"errors"
	"easymail/internal/model"
	"easymail/internal/service/storage"
	"sync"
)

var localStorage *storage.LocalStorage
var localStorageOnce sync.Once
var localStorageErr error

func getLocalStorage() (*storage.LocalStorage, error) {
	localStorageOnce.Do(func() {
		c, err := model.GetConfigureByNames("easymail", "storage", "data")
		if err != nil {
			localStorageErr = errors.New("mail storage data is not defined")
			return
		}
		r, err := model.GetConfigureByNames("easymail", "configure", "root")
		if err != nil {
			localStorageErr = errors.New("easymail configure root is not defined")
			return
		}
		db, err := model.DB()
		if err != nil {
			localStorageErr = err
			return
		}
		localStorage = storage.NewLocalStorage(r.Value, c.Value, db)
	})
	if localStorageErr != nil {
		return nil, localStorageErr
	}
	return localStorage, nil
}
