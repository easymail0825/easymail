package model

import (
	"errors"
	"gorm.io/gorm"
	"testing"
)

func TestGetRootConfigureRootNodes(t *testing.T) {
	nodes, err := GetRootConfigureRootNodes()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(nodes)

	for _, node := range nodes {
		t.Log(node.Name)
	}
}

func TestGetConfigureByParentId(t *testing.T) {
	nodes, err := GetConfigureByParentId(15)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(nodes)

	for _, node := range nodes {
		t.Log(node.Name)
	}
}

func TestAddConfigure(t *testing.T) {
	type ConfigureItem struct {
		names    []string
		value    string
		dataType DataType
		describe string
	}

	defaultConfigureItems := []ConfigureItem{
		//{[]string{"feature", "ip", "ptr"}, "true", DataTypeBool, "lookup ptr record of ip"},
		//{[]string{"feature", "ip", "region"}, "true", DataTypeBool, "lookup region info of ip"},
		//{[]string{"feature", "ip", "region-city-mmdb"}, "GeoLite2-City.mmdb", DataTypeString, "path of mmdb"},
		//{[]string{"feature", "ip", "rbl"}, "true", DataTypeBool, "query rbl for ip"},
		//{[]string{"network", "rbl", "sorbs"}, "rbl/sorbs", DataTypeString, "sorbs rbl"},
		//{[]string{"feature", "ip", "10min"}, "true", DataTypeBool, "ip total requests in 10 min"},
		//{[]string{"feature", "ip", "1day"}, "true", DataTypeBool, "ip total requests in 1 day"},
		//{[]string{"feature", "ip", "10min-ham"}, "true", DataTypeBool, "ip total ham in 10 min"},
		//{[]string{"feature", "ip", "1day-ham"}, "true", DataTypeBool, "ip total ham in 1 day"},
		//{[]string{"feature", "ip", "10min-spam"}, "true", DataTypeBool, "ip total ham in 10 min"},
		//{[]string{"feature", "ip", "1day-spam"}, "true", DataTypeBool, "ip total ham in 1 day"},
		//{[]string{"feature", "ip", "10min-sender"}, "true", DataTypeBool, "ip total sender in 10 min"},
		//{[]string{"feature", "ip", "1day-sender"}, "true", DataTypeBool, "ip total sender in 1 day"},
		//{[]string{"feature", "ip", "10min-ham-sender"}, "true", DataTypeBool, "ip total ham sender in 10 min"},
		//{[]string{"feature", "ip", "1day-ham-sender"}, "true", DataTypeBool, "ip total ham sender in 1 day"},
		//{[]string{"feature", "ip", "10min-spam-sender"}, "true", DataTypeBool, "ip total spam sender in 10 min"},
		//{[]string{"feature", "ip", "1day-spam-sender"}, "true", DataTypeBool, "ip total spam sender in 1 day"},
		//{[]string{"feature", "ip", "1hour-spam-rate"}, "true", DataTypeBool, "ip spam rate in 1 hour"},
		//{[]string{"feature", "domain", "a"}, "true", DataTypeBool, "lookup a record of domain"},
		//{[]string{"feature", "domain", "mx"}, "true", DataTypeBool, "lookup mx record of domain"},
		//{[]string{"feature", "domain", "spf"}, "true", DataTypeBool, "lookup spf record of domain, and validate"},
		//{[]string{"feature", "domain", "dkim"}, "true", DataTypeBool, "lookup dkim record of domain, and validate"},
		//{[]string{"feature", "domain", "dmarc"}, "true", DataTypeBool, "lookup dmarc record of domain, and validate"},
		//{[]string{"feature", "domain", "10min"}, "true", DataTypeBool, "domain total requests in 10 min"},
		//{[]string{"feature", "domain", "1day"}, "true", DataTypeBool, "domain total requests in 1 day"},
		//{[]string{"feature", "domain", "10min-ham"}, "true", DataTypeBool, "domain total ham in 10 min"},
		//{[]string{"feature", "domain", "1day-ham"}, "true", DataTypeBool, "domain total ham in 1 day"},
		//{[]string{"feature", "domain", "10min-spam"}, "true", DataTypeBool, "domain total spam in 10 min"},
		//{[]string{"feature", "domain", "1day-spam"}, "true", DataTypeBool, "domain total spam in 1 day"},
		//{[]string{"feature", "domain", "1hour-spam-rate"}, "true", DataTypeBool, "domain spam rate in 1 hour"},
		//{[]string{"feature", "sender", "10min"}, "true", DataTypeBool, "sender requests in 10 min"},
		//{[]string{"feature", "sender", "1day"}, "true", DataTypeBool, "sender requests in 1 day"},
		//{[]string{"feature", "sender", "10min-ham"}, "true", DataTypeBool, "sender total ham in 10 min"},
		//{[]string{"feature", "sender", "1day-ham"}, "true", DataTypeBool, "sender total ham in 1 day"},
		//{[]string{"feature", "sender", "10min-spam"}, "true", DataTypeBool, "sender total spam in 10 min"},
		//{[]string{"feature", "sender", "1day-spam"}, "true", DataTypeBool, "sender total spam in 1 day"},
		//{[]string{"feature", "sender", "auto-whitelist"}, "true", DataTypeBool, "sender and receipt match auto whitelist"},
		//{[]string{"feature", "sender", "relation"}, "true", DataTypeBool, "relation between sender and receipt"},
		//{[]string{"feature", "sender", "1hour-spam-rate"}, "true", DataTypeBool, "sender spam rate in 1 hour"},
		//{[]string{"feature", "attach-md5", "10min"}, "true", DataTypeBool, "attach md5 requests in 10 min"},
		//{[]string{"feature", "attach-md5", "1day"}, "true", DataTypeBool, "attach md5 requests in 1 day"},
		//{[]string{"feature", "attach-md5", "10min-ham"}, "true", DataTypeBool, "attach md5 total ham in 10 min"},
		//{[]string{"feature", "attach-md5", "1day-ham"}, "true", DataTypeBool, "attach md5 total ham in 1 day"},
		//{[]string{"feature", "attach-md5", "10min-spam"}, "true", DataTypeBool, "attach md5 total spam in 10 min"},
		//{[]string{"feature", "attach-md5", "1day-spam"}, "true", DataTypeBool, "attach md5 total spam in 1 day"},
		//{[]string{"feature", "attach-md5", "1hour-spam-rate"}, "true", DataTypeBool, "attach md5 spam rate in 1 hour"},
		//{[]string{"feature", "attach-hash", "10min"}, "true", DataTypeBool, "attach hash requests in 10 min"},
		//{[]string{"feature", "attach-hash", "1day"}, "true", DataTypeBool, "attach hash requests in 1 day"},
		//{[]string{"feature", "attach-hash", "10min-ham"}, "true", DataTypeBool, "attach hash total ham in 10 min"},
		//{[]string{"feature", "attach-hash", "1day-ham"}, "true", DataTypeBool, "attach hash total ham in 1 day"},
		//{[]string{"feature", "attach-hash", "10min-spam"}, "true", DataTypeBool, "attach hash total spam in 10 min"},
		//{[]string{"feature", "attach-hash", "1day-spam"}, "true", DataTypeBool, "attach hash total spam in 1 day"},
		//{[]string{"feature", "attach-hash", "1hour-spam-rate"}, "true", DataTypeBool, "attach hash spam rate in 1 hour"},
		//{[]string{"feature", "text-hash", "10min"}, "true", DataTypeBool, "text hash requests in 10 min"},
		//{[]string{"feature", "text-hash", "1day"}, "true", DataTypeBool, "text hash requests in 1 day"},
		//{[]string{"feature", "text-hash", "10min-ham"}, "true", DataTypeBool, "text hash total ham in 10 min"},
		//{[]string{"feature", "text-hash", "1day-ham"}, "true", DataTypeBool, "text hash total ham in 1 day"},
		//{[]string{"feature", "text-hash", "10min-spam"}, "true", DataTypeBool, "text hash total spam in 10 min"},
		//{[]string{"feature", "text-hash", "1day-spam"}, "true", DataTypeBool, "text hash total spam in 1 day"},
		//{[]string{"feature", "text-hash", "1hour-spam-rate"}, "true", DataTypeBool, "text hash spam rate in 1 hour"},
	}
	for _, cfg := range defaultConfigureItems {
		if _, err := GetConfigure(cfg.names...); errors.Is(err, gorm.ErrRecordNotFound) {
			_, err = CreateConfigure(cfg.value, cfg.describe, cfg.dataType, cfg.names...)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
