package controller

import (
	_ "easymail/internal/database"
	"easymail/internal/dns"
	"easymail/internal/model"
	"log"
)

var resolver *dns.Resolver

func init() {
	nameserver := "8.8.8.8" // 默认使用8.8.8.8作为DNS服务器
	c, err := model.GetConfigure("network", "dns", "nameserver")
	if err != nil {
		log.Println("get configure error:", err)
	} else {
		if c.DataType == model.DataTypeString {
			nameserver = c.Value
		}
	}
	resolver = dns.NewResolver(nameserver)
}
