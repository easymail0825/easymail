package dkim

import "easymail/internal/easydns"

var resolver *easydns.Resolver

func init() {
	//resolver = easydns.New("114.114.114.114")
	resolver = easydns.CreateDefaultResolver()
}
