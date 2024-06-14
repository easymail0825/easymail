package easydns

import (
	"context"
	"easymail/internal/database"
	"easymail/internal/model"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"strings"
	"time"
)

const (
	ExpireTime = 3600 * time.Second
)

type Resolver struct {
	NameServer string
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

func New(nameServer string) *Resolver {
	return &Resolver{
		NameServer: nameServer,
	}
}

func CreateDefaultResolver() *Resolver {
	nameserver := "8.8.8.8" // default dns server
	c, err := model.GetConfigureByNames("network", "dns", "nameserver")
	if err != nil {
		log.Println("get configure error:", err)
	} else {
		if c.DataType == model.DataTypeString {
			nameserver = c.Value
		}
	}
	return New(nameserver)
}

func (r *Resolver) DomainExist(domain string) (bool, error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	//query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:domain:%s", domain)
	value, err := rdb.Get(context.Background(), key).Result()
	if err == nil {
		return value == "true", nil
	}

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return false, err
	}

	if msg.Rcode != dns.RcodeSuccess {
		return false, fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	// save cache
	rdb.Set(context.Background(), key, "true", ExpireTime)
	return true, nil
}

func (r *Resolver) LookupIPAddr(domain string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	//query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:ip:%s", domain)
	value, err := rdb.Get(context.Background(), key).Result()
	if err == nil {
		return strings.Split(value, ","), nil
	}

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return nil, err
	}

	if msg.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	for _, rr := range msg.Answer {
		if a, ok := rr.(*dns.A); ok {
			data = append(data, a.A.String())
		}
	}

	// save cache
	rdb.Set(context.Background(), key, strings.Join(data, ","), ExpireTime)

	return data, nil

}

func (r *Resolver) LookupTXT(domain string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	//query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:txt:%s", domain)
	value, err := rdb.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
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

		cnames := make([]string, 0)
		for _, rr := range msg.Answer {
			if txt, ok := rr.(*dns.TXT); ok {
				data = append(data, strings.Join(txt.Txt, ""))
			}
			if cname, ok := rr.(*dns.CNAME); ok {
				cnames = append(cnames, cname.Target)
			}
		}

		// query cnames and append them to data
		for _, cname := range cnames {
			m := new(dns.Msg)
			m.SetQuestion(cname, dns.TypeTXT)

			msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
			if err != nil {
				continue
			}

			if msg.Rcode != dns.RcodeSuccess {
				continue
			}

			for _, rr := range msg.Answer {
				if txt, ok := rr.(*dns.TXT); ok {
					data = append(data, strings.Join(txt.Txt, ""))
				}
			}

			// save cache
			rdb.Set(context.Background(), key, strings.Join(data, ","), ExpireTime)

			return data, nil
		}

		// save cache
		rdb.Set(context.Background(), key, strings.Join(data, ","), ExpireTime)

		return data, nil
	}

	if err == nil {
		return strings.Split(value, ","), nil
	}
	return nil, err
}

func (r *Resolver) LookupPtr(ip string) (data []string, err error) {
	//query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:ptr:%s", ip)
	value, err := rdb.Get(context.Background(), key).Result()
	if err == nil {
		return strings.Split(value, ","), nil
	}

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

	// save cache
	rdb.Set(context.Background(), key, strings.Join(data, ","), ExpireTime)

	return data, nil
}

func (r *Resolver) LookupMX(domain string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	// query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:mx:%s", domain)
	value, err := rdb.Get(context.Background(), key).Result()
	if err == nil {
		return strings.Split(value, ","), nil
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

	// save cache
	rdb.Set(context.Background(), key, strings.Join(data, ","), ExpireTime)

	return data, nil
}

func (r *Resolver) LookupSPF(domain string) (record string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	// query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:spf:%s", domain)
	value, err := rdb.Get(context.Background(), key).Result()
	if err == nil {
		return value, nil
	}

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeTXT)

	msg, _, err := c.Exchange(m, net.JoinHostPort(r.NameServer, "53"))
	if err != nil {
		return "", err
	}

	// fixme
	if msg.Rcode != dns.RcodeSuccess {
		return "", fmt.Errorf("DNS query failed with code: %d", msg.Rcode)
	}

	records := []string{}
	for _, rr := range msg.Answer {
		if txt, ok := rr.(*dns.TXT); ok {
			if strings.HasPrefix(strings.ToLower(txt.Txt[0]), "v=spf1 ") || strings.ToLower(txt.Txt[0]) == "v=spf1" {
				records = append(records, txt.String())
			}
		}
	}

	// 0 records is ok, handled by the parent.
	// 1 record is what we expect, return the record.
	// More than that, it's a permanent error:
	// https://tools.ietf.org/html/rfc7208#section-4.5
	l := len(records)
	if l == 0 {
		record = ""
	} else if l == 1 {
		record = records[0]
	}

	// save cache
	rdb.Set(context.Background(), key, record, ExpireTime)

	return record, nil
}

func (r *Resolver) LookupDKIM(domain, selector string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	// query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:dkim:%s", domain)
	value, err := rdb.Get(context.Background(), key).Result()
	if err == nil {
		return strings.Split(value, ","), nil
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

	// save cache
	rdb.Set(context.Background(), key, strings.Join(data, ","), ExpireTime)

	return data, nil
}

func (r *Resolver) LookupDMARC(domain string) (data []string, err error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	// query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("dns:dmarc:%s", domain)
	value, err := rdb.Get(context.Background(), key).Result()
	if err == nil {
		return strings.Split(value, ","), nil
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

	// save cache
	rdb.Set(context.Background(), key, strings.Join(data, ","), ExpireTime)

	return data, nil
}
