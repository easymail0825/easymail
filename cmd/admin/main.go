package main

import (
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/service/admin/router"
	"encoding/gob"
	"os"
)

func main() {
	gob.Register(model.Account{})
	logFile, err := os.OpenFile("test.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	_log := easylog.NewLogger(logFile, "")
	r := router.New(_log, "/home/bobxiao/Projects/golang/easymail/internal/service/admin", "xxxxxxxxxxxxxxxx", "easymail_admin")
	r.Run("127.0.0.1:10088")
}
