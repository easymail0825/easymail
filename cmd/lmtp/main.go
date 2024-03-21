package main

import (
	"easymail/internal/lmtp"
	"easymail/internal/storage"
	"log"
)

func main() {
	server := lmtp.New("tcp", "0.0.0.0:10025", 1024*1024*50, []string{"8BITMIME", "ENHANCEDSTATUSCODES", "PIPELINING"}...)

	if server == nil {
		log.Panicln("Failed to create server")
	}

	// add storage
	localStorage := storage.NewLocalStorage("./data")
	server.SetStorage(localStorage)

	if err := server.Run(); err != nil {
		log.Panicln(err)
	}
}
