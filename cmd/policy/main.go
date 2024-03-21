package main

import (
	"easymail/internal/policy"
	"log"
)

func main() {
	server := policy.NewCheckPolicyServer("tcp", "0.0.0.0:10026")

	if server == nil {
		log.Panicln("Failed to create server")
	}

	if err := server.Run(); err != nil {
		log.Panicln(err)
	}
}
