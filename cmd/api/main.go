package main

import (
	"easymail/internal/app/bootstrap"
	"easymail/internal/app/wire"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	configPath := "easymail.yaml"
	if _, err := os.Stat(configPath); err != nil {
		// repo default
		configPath = filepath.FromSlash("configs/easymail.yaml")
		if _, err := os.Stat(configPath); err != nil {
			// legacy location
			configPath = filepath.FromSlash("cmd/easymail/easymail.yaml")
		}
	}
	rt, err := bootstrap.Start(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if rt.Tracer != nil {
			_ = rt.Tracer.Close()
		}
	}()
	manager, err := wire.Build(rt)
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
