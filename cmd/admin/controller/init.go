package controller

import (
	"easymail/internal/config"
	_ "easymail/internal/database"
	"easymail/internal/dns"
	"log"
)

var resolver *dns.Resolver

func init() {
	nameserver := "8.8.8.8" // 默认使用8.8.8.8作为DNS服务器
	c, err := config.GetConfigure("network", "dns", "nameserver")
	if err != nil {
		log.Println("get configure error:", err)
	} else {
		if c.DataType == config.DataTypeString {
			nameserver = c.Value
		}
	}
	resolver = dns.NewResolver(nameserver)
}
