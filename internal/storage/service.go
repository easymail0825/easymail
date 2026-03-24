package storage

import oldstorage "easymail/internal/service/storage"

func NewLocal(root, dataPath string) *oldstorage.LocalStorage {
	return oldstorage.NewLocalStorage(root, dataPath)
}

