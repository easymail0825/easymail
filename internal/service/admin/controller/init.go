package controller

import (
	_ "easymail/internal/database"
	"easymail/internal/easydns"
)

var resolver *easydns.Resolver

func init() {
	resolver = easydns.CreateDefaultResolver()
}
