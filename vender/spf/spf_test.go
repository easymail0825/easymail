package spf

import (
	"easymail/internal/easydns"
	"net"
	"testing"
)

func TestSPF(t *testing.T) {
	resolver := easydns.New("8.8.8.8")
	//result, err := CheckHostWithSender(resolver, net.IP("1.2.3.4"), "example.com", "a@example.com")
	//if err != nil {
	//	t.Errorf("Error: %s", err)
	//}
	//t.Log(result)

	result, err := CheckHostWithSender(resolver, net.ParseIP("162.62.58.69"), "qq.com", "a@qq.com")
	//result, err := CheckHostWithSender(resolver, net.ParseIP("107.173.114.205"), "easypostix.com", "a@easypostix.com")
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	t.Log(result)
}
