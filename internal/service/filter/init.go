package filter

import (
	"easymail/internal/database"
	"easymail/internal/easydns"
	"easymail/internal/model"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	mmdbreader "github.com/oschwald/maxminddb-golang"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"time"
)

var rdb *redis.Client
var resolver *easydns.Resolver
var geoip *mmdbreader.Reader
var knowledgeLibrary *ast.KnowledgeLibrary
var ruleBuilder *builder.RuleBuilder
var knowledgeInstance *ast.KnowledgeBase
var ruleEngine *engine.GruleEngine

func reloadRules() {
	// load rule first
	ok, okb, orb, ore, err := loadRules()
	if err != nil {
		log.Println("load rules error:", err)
	} else {
		ruleBuilder = orb
		knowledgeLibrary = ok
		knowledgeInstance = okb
		ruleEngine = ore
	}

	// then use ticker execute every 1 minute
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	lastTime := time.Time{}
	for {
		select {
		case <-ticker.C:
			t, err := model.GetLastTimeOfRule()
			if err == nil && t.After(lastTime) {
				k, kb, rb, re, err := loadRules()
				if err != nil {
					log.Println("load rules error:", err)
				} else {
					lastTime = t
					knowledgeLibrary = k
					knowledgeInstance = kb
					ruleBuilder = rb
					ruleEngine = re
				}
			}
		}
	}
}

func init() {
	rdb = database.GetRedisClient()
	nameserver := "8.8.8.8"
	c, err := model.GetConfigureByNames("network", "dns", "nameserver")
	if err != nil {
		log.Println("get configure error:", err)
	} else {
		if c.DataType == model.DataTypeString {
			nameserver = c.Value
		}
	}
	resolver = easydns.New(nameserver)

	// init maxmind db
	if featureSwitch([]string{"feature", "ip", "region"}) {
		if c, err := model.GetConfigureByNames("feature", "ip", "region-city-mmdb"); err == nil {
			if c.DataType == model.DataTypeString {
				geoip, err = mmdbreader.Open(c.Value)
				if err != nil {
					log.Println("open mmdb error:", err)
				}
			}
		}
	}

	// init knowledge library
	l := logrus.New()
	l.Out = os.Stderr
	l.SetLevel(logrus.PanicLevel)
	ast.SetLogger(l)
	engine.SetLogger(l)

	// load filter rules
	go reloadRules()
}
