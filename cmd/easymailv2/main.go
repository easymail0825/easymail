package main

import (
	"easymail/internal/platform/app"
	"easymail/internal/platform/bootstrap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	rt, err := bootstrap.Start("easymail.yaml")
	if err != nil {
		log.Fatal(err)
	}
	manager, err := app.Build(rt)
	if err != nil {
		log.Fatal(err)
	}
	if err = manager.StartAll(); err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	manager.StopAll()
}

