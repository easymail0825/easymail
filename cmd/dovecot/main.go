package main

import (
	"easymail/internal/dovecot"
	"log"
)

func main() {
	server := dovecot.New("tcp", "0.0.0.0:10028")
	if server == nil {
		log.Panicln("Failed to create server")
	}

	if err := server.Run(); err != nil {
		log.Panicln(err)
	}
}
