package filter

import (
	"easymail/internal/database"
	"easymail/internal/dns"
	"easymail/internal/model"
	mmdbreader "github.com/oschwald/maxminddb-golang"
	"github.com/redis/go-redis/v9"
	"log"
)

var rdb *redis.Client
var resolver *dns.Resolver
var geoip *mmdbreader.Reader

func init() {
	rdb = database.GetRedisClient()
	nameserver := "8.8.8.8"
	c, err := model.GetConfigure("network", "dns", "nameserver")
	if err != nil {
		log.Println("get configure error:", err)
	} else {
		if c.DataType == model.DataTypeString {
			nameserver = c.Value
		}
	}
	resolver = dns.NewResolver(nameserver)

	// init maxmind db
	if featureSwitch([]string{"feature", "ip", "region"}) {
		if c, err := model.GetConfigure("feature", "ip", "region-city-mmdb"); err == nil {
			if c.DataType == model.DataTypeString {
				geoip, err = mmdbreader.Open(c.Value)
				if err != nil {
					log.Println("open mmdb error:", err)
				}
			}
		}

	}
}
