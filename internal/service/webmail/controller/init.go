package controller

import (
	_ "easymail/internal/database"
	"easymail/internal/model"
	"easymail/internal/service/storage"
	"log"
)

var localStorage *storage.LocalStorage

func init() {
	c, err := model.GetConfigure("easymail", "storage", "data")
	if err != nil {
		log.Fatal("mail storage data is not defined")
	}
	r, err := model.GetConfigure("easymail", "configure", "root")
	if err != nil {
		log.Fatal("easymail configure root is not defined")
	}
	localStorage = storage.NewLocalStorage(r.Value, c.Value)
}
