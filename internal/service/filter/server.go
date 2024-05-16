package filter

import (
	"context"
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/service/milter"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"sync"
	"time"
)

type Filter struct {
	milter.Milter
	optAction   milter.OptAction
	optProtocol milter.OptProtocol
}

func (f *Filter) Connect(host string, family string, port uint16, addr net.IP, m *milter.Modifier, feature []milter.Feature) (milter.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
	defer cancel()

	// query ip ptr
	if featureSwitch([]string{"feature", "ip", "ptr"}) {
		go func(ctx context.Context) {
			ptr, err := QueryPtr(addr.String())
			if err == nil {
				feature = append(feature, milter.Feature{Name: "ip-ptr", Value: ptr, DataType: model.DataTypeString})
			}
		}(ctx)
	}

	// query ip region info
	// only supper GeoLite2-City.mmdb from official https://download.maxmind.com/app/geoip_download_by_token
	if featureSwitch([]string{"feature", "ip", "region"}) {
		if geoip != nil {
			go func(ctx context.Context) {
				country, province, city, err := QueryRegion(addr)
				if err == nil {
					feature = append(feature, milter.Feature{Name: "ip-country", Value: country, DataType: model.DataTypeString})
					feature = append(feature, milter.Feature{Name: "ip-province", Value: province, DataType: model.DataTypeString})
					feature = append(feature, milter.Feature{Name: "ip-city", Value: city, DataType: model.DataTypeString})
				}
			}(ctx)
		}
	}

	if err := increaseCount(ctx, fmt.Sprintf("%s:10min-request", addr.String()), time.Minute*10); err != nil {
		log.Println("Error updating IP count in Redis:", err)
	}
	return milter.RespContinue, nil
}

func (f *Filter) Helo(name string, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (f *Filter) MailFrom(from string, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (f *Filter) RcptTo(rcptTo string, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (f *Filter) Header(name string, value string, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (f *Filter) Headers(h textproto.MIMEHeader, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (f *Filter) BodyChunk(chunk []byte, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (f *Filter) Body(m *milter.Modifier, macro map[string]string) (milter.Response, error) {
	if v, ok := macro["i"]; ok {
		m.AddHeader("X-Queue-ID", v)
	}
	return milter.RespContinue, nil
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
		session := milter.NewSession(client, s.filter, s.filter.optAction, s.filter.optProtocol)
		go session.Handle()
	}
}
