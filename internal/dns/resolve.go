package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"strings"
)

type Resolver struct {
	NameServer string
	// todo cache
}

func ReverseIP(ip string) string {
	parts := strings.Split(ip, ".")
	parts = ReverseString(parts)
	return strings.Join(parts, ".")
}

func ReverseString(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func NewResolver(nameServer string) *Resolver {
	return &Resolver{
		NameServer: nameServer,
	}
}

func (r *Resolver) LookupPtr(ip string) (data []string, err error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(fmt.Sprintf("%s.in-addr.arpa.", ReverseIP(ip)), dns.TypePTR)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return nil, err
	}

	if msg.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	for _, rr := range msg.Answer {
		if ptr, ok := rr.(*dns.PTR); ok {
			data = append(data, ptr.Ptr)
		}
	}

	return data, nil
}

func (r *Resolver) LookupMX(domain string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeMX)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return nil, err
	}

	if msg.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	for _, rr := range msg.Answer {
		if mx, ok := rr.(*dns.MX); ok {
			data = append(data, mx.String())
		}
	}

	return data, nil
}

func (r *Resolver) LookupSPF(domain string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeTXT)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return nil, err
	}

	if msg.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	for _, rr := range msg.Answer {
		if txt, ok := rr.(*dns.TXT); ok {
			if strings.HasPrefix(strings.ToLower(txt.Txt[0]), "v=spf1") {
				data = append(data, txt.String())
			}
		}
	}

	return data, nil
}
func (r *Resolver) LookupDKIM(domain, selector string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	domain = fmt.Sprintf("%s._domainkey.%s", selector, domain)
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeTXT)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return nil, err
	}

	if msg.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	for _, rr := range msg.Answer {
		if txt, ok := rr.(*dns.TXT); ok {
			if strings.HasPrefix(strings.ToLower(txt.Txt[0]), "v=dkim1") {
				data = append(data, txt.String())
			}
		}
	}

	return data, nil
}

func (r *Resolver) LookupDMARC(domain string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	domain = fmt.Sprintf("_dmarc.%s", domain)
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeTXT)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return nil, err
	}

	if msg.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	for _, rr := range msg.Answer {
		if txt, ok := rr.(*dns.TXT); ok {
			if strings.HasPrefix(strings.ToLower(txt.Txt[0]), "v=dmarc1") {
				data = append(data, txt.String())
			}
		}
	}

	return data, nil
}
