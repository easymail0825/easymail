package filter

import (
	"context"
	"easymail/internal/easylog"
	"easymail/internal/service/milter"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type Filter struct {
	milter.Milter
	optAction   milter.OptAction
	optProtocol milter.OptProtocol
}

const timeout = 100000 * time.Microsecond

func (f *Filter) Connect(host string, family string, port uint16, addr net.IP, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	addr = net.ParseIP("211.136.192.6")
	isPrivate := addr.IsPrivate() || addr.IsLoopback()
	feature := make([]milter.Feature, 0)
	// add feature of ip
	feature = append(feature, milter.Feature{Name: "client_ip", Value: addr.String(), ValueType: milter.DataTypeString})
	ctxPtr, cancelPtr := context.WithTimeout(context.Background(), timeout)
	defer cancelPtr()

	// query ip ptr
	if !isPrivate && featureSwitch([]string{"feature", "ip", "ptr"}) {
		go func(ctx context.Context) {
			ptr, err := QueryPtr(addr.String())
			if err == nil && ptr != "" {
				feature = append(feature, milter.Feature{Name: "ip_ptr", Value: "true", ValueType: milter.DataTypeBool})
			} else {
				feature = append(feature, milter.Feature{Name: "ip_ptr", Value: "false", ValueType: milter.DataTypeBool})
			}
		}(ctxPtr)
	}

	// query ip region info
	// only supper GeoLite2-City.mmdb from official https://download.maxmind.com/app/geoip_download_by_token
	ctxRegion, cancelRegion := context.WithTimeout(context.Background(), timeout)
	defer cancelRegion()
	if !isPrivate && featureSwitch([]string{"feature", "ip", "region"}) {
		matched := false
		// query redis first
		ctxRedisQuery, cancelRedisQuery := context.WithTimeout(context.Background(), timeout)
		defer cancelRedisQuery()
		if tmp, err := queryCacheInString(ctxRedisQuery, fmt.Sprintf("ip:country:%s", addr.String())); err == nil {
			matched = true
			if tmp != "" {
				feature = append(feature, milter.Feature{Name: "ip_country", Value: strings.ToLower(tmp), ValueType: milter.DataTypeString})
			}
		}
		if tmp, err := queryCacheInString(ctxRedisQuery, fmt.Sprintf("ip:province:%s", addr.String())); err == nil {
			matched = true
			if tmp != "" {
				feature = append(feature, milter.Feature{Name: "ip_province", Value: strings.ToLower(tmp), ValueType: milter.DataTypeString})
			}
		}
		if tmp, err := queryCacheInString(ctxRedisQuery, fmt.Sprintf("ip:city:%s", addr.String())); err == nil {
			matched = true
			if tmp != "" {
				feature = append(feature, milter.Feature{Name: "ip_city", Value: strings.ToLower(tmp), ValueType: milter.DataTypeString})
			}
		}

		if !isPrivate && !matched && geoip != nil {
			go func(ctx context.Context) {
				country, province, city, err := QueryRegion(addr)
				if err == nil {
					if country != "" {
						feature = append(feature, milter.Feature{Name: "ip_country", Value: strings.ToLower(country), ValueType: milter.DataTypeString})
						if err := setCacheInString(ctx, fmt.Sprintf("ip:country:%s", addr.String()), strings.ToLower(country), time.Hour*24); err != nil {
							log.Println("[ERROR]", err)
						}
					}
					if province != "" {
						feature = append(feature, milter.Feature{Name: "ip_province", Value: strings.ToLower(province), ValueType: milter.DataTypeString})
						if err := setCacheInString(ctx, fmt.Sprintf("ip:province:%s", addr.String()), strings.ToLower(province), time.Hour*24); err != nil {
							log.Println("[ERROR]", err)
						}
					}
					if city != "" {
						feature = append(feature, milter.Feature{Name: "ip_city", Value: strings.ToLower(city), ValueType: milter.DataTypeString})
						if err := setCacheInString(ctx, fmt.Sprintf("ip:city:%s", addr.String()), strings.ToLower(city), time.Hour*24); err != nil {
							log.Println("[ERROR]", err)
						}
					}
				}
			}(ctxRegion)
		}
	}

	ctxRedis, cancelRedis := context.WithTimeout(context.Background(), timeout)
	defer cancelRedis()
	if featureSwitch([]string{"feature", "ip", "10min"}) {
		go func(ctx context.Context) {
			if n, err := increaseCount(ctx, fmt.Sprintf("ip:%s:%s", addr.String(), formatTimeForTenMinutes(time.Now())), time.Minute*10); err == nil {
				feature = append(feature, milter.Feature{Name: "ip_10min_request", Value: fmt.Sprintf("%d", n), ValueType: milter.DataTypeInt})
			}
		}(ctxRedis)
	}

	if featureSwitch([]string{"feature", "ip", "1day"}) {
		go func(ctx context.Context) {
			if n, err := increaseCount(ctx, fmt.Sprintf("ip:%s:%s", addr.String(), time.Now().Format("20060102")), time.Hour*24); err == nil {
				feature = append(feature, milter.Feature{Name: "ip_1day_request", Value: fmt.Sprintf("%d", n), ValueType: milter.DataTypeInt})
			}
		}(ctxRedis)
	}

	select {
	case <-ctxPtr.Done():
	case <-ctxRegion.Done():
	case <-ctxRedis.Done():
	}
	return milter.RespContinue, feature, nil
}

func (f *Filter) Helo(name string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	//add feature of helo argument
	feature := make([]milter.Feature, 0)
	feature = append(feature, milter.Feature{Name: "helo", Value: strings.ToLower(name), ValueType: milter.DataTypeString})
	return milter.RespContinue, feature, nil
}

func (f *Filter) MailFrom(from string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	return milter.RespContinue, nil, nil
}

func (f *Filter) RcptTo(rcptTo string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	return milter.RespContinue, nil, nil
}

func (f *Filter) Header(name string, value string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	return milter.RespContinue, nil, nil
}

func (f *Filter) Headers(h textproto.MIMEHeader, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	return milter.RespContinue, nil, nil
}

func (f *Filter) BodyChunk(chunk []byte, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	return milter.RespContinue, nil, nil
}

func (f *Filter) Body(m *milter.Modifier, macro map[string]string) (milter.Response, []milter.Feature, error) {
	if v, ok := macro["i"]; ok {
		m.AddHeader("X-Queue-ID", v)
	}
	return milter.RespContinue, nil, nil
}

type Server struct {
	name    string
	stopCh  chan struct{}
	started bool
	lock    *sync.Mutex
	family  string
	listen  string
	debug   bool
	_log    *easylog.Logger

	filter *Filter
}

func New(family string, listen string) (server *Server) {
	server = &Server{
		name:    "filter",
		stopCh:  make(chan struct{}),
		lock:    &sync.Mutex{},
		started: false,
		family:  family,
		listen:  listen,

		filter: &Filter{
			optProtocol: milter.OptProtocol(milter.OptChangeBody | milter.OptChangeFrom | milter.OptChangeHeader | milter.OptAddHeader | milter.OptAddRcpt | milter.OptChangeFrom),
			optAction:   0,
		},
	}
	return server
}

func (s *Server) SetDebug(debug bool) {
	s.debug = debug
}

func (s *Server) SetLogger(_log *easylog.Logger) error {
	if _log == nil {
		return fmt.Errorf("%s logger is nil", s.name)
	}
	s._log = _log
	return nil
}

func (s *Server) Start() error {
	if s._log == nil {
		return fmt.Errorf("%s logger is nil", s.name)
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.started {
		return fmt.Errorf("%s server already started", s.name)
	}

	s.started = true
	s._log.Infof("%s server started!", s.name)
	go s.run()
	return nil
}

func (s *Server) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.started {
		return fmt.Errorf("%s server not started", s.name)
	}

	s.started = false
	close(s.stopCh)
	s._log.Infof("%s server stopped!", s.name)
	return nil
}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) run() (err error) {
	listener, err := net.Listen(s.family, s.listen)
	if err != nil {
		return err
	}

	for {
		client, err := listener.Accept()
		if err != nil {
			return err
		}
		session := NewSession(client, s.filter, s._log)
		go session.Handle()
	}
}
