package easydns

import "testing"

func TestResolver_LookupMX(t *testing.T) {
	r := New("8.8.8.8")

	mxs, err := r.LookupMX("qq.com")

	if err != nil {
		t.Skipf("LookupMX error: %v", err)
	} else {
		t.Logf("MX records for example.com: %v", mxs)
	}
}

func TestResolver_LookupSPF(t *testing.T) {
	r := New("8.8.8.8")

	mxs, err := r.LookupSPF("qq.com")

	if err != nil {
		t.Skipf("LookupSPF error: %v", err)
	} else {
		t.Logf("SPF records for example.com: %v", mxs)
	}
}

func TestResolver_LookupDKIM(t *testing.T) {
	r := New("8.8.8.8")

	mxs, err := r.LookupDKIM("_domainkey.qq.com", "default")

	if err != nil {
		t.Skipf("LookupDKIM error: %v", err)
	} else {
		t.Logf("DKIM records for example.com: %v", mxs)
	}
}

func TestResolver_LookupDMARC(t *testing.T) {
	r := New("8.8.8.8")

	mxs, err := r.LookupDMARC("qq.com")

	if err != nil {
		t.Skipf("LookupDMARC error: %v", err)
	} else {
		t.Logf("DMARC records for example.com: %v", mxs)
	}
}

func TestDomainExist(t *testing.T) {
	r := New("8.8.8.8")

	exist, err := r.DomainExist("baidu.com")

	if err != nil {
		t.Skipf("DomainExist error: %v", err)
	} else {
		t.Logf("Domain exist: %v", exist)
	}
}
