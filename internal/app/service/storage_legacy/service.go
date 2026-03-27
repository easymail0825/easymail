package storage

import (
	oldstorage "easymail/internal/service/storage"
	"gorm.io/gorm"
)

func NewLocal(root, dataPath string, db *gorm.DB) *oldstorage.LocalStorage {
	return oldstorage.NewLocalStorage(root, dataPath, db)
}
