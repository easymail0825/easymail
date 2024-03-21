package main

import (
	"easymail/internal/filter"
	"easymail/internal/milter"
)

func main() {
	milterServer, err := filter.New(
		"tcp",
		"0.0.0.0:10027",
		milter.OptChangeBody|milter.OptChangeFrom|milter.OptChangeHeader|milter.OptAddHeader|milter.OptAddRcpt|milter.OptChangeFrom,
		0,
	)
	if err != nil {
		panic(err)
	}
	milterServer.Run()
}
