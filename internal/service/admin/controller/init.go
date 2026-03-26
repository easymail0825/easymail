package controller

import (
	"easymail/internal/easydns"
	"sync"
)

var resolver *easydns.Resolver
var resolverOnce sync.Once

func getResolver() *easydns.Resolver {
	resolverOnce.Do(func() {
		resolver = easydns.CreateDefaultResolver()
	})
	return resolver
}
