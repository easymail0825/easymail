package filter

import (
	"net"
	"testing"
)

func TestQueryRegion(t *testing.T) {
	country, province, city, err := QueryRegion(net.ParseIP("120.8.3.3"))
	if err != nil {
		t.Error(err)
	}
	t.Log(country, province, city)
}
