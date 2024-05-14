package filter

import (
	"easymail/internal/easylog"
	milter "easymail/internal/service/milter"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"sync"
)

type Filter struct {
	milter.Milter
	optAction   milter.OptAction
	optProtocol milter.OptProtocol
}

func (f *Filter) Connect(host string, family string, port uint16, addr net.IP, m *milter.Modifier) (milter.Response, error) {
	log.Println("connect from ", host, family, port, addr)
	return milter.RespContinue, nil
}

func (f *Filter) Helo(name string, m *milter.Modifier) (milter.Response, error) {
	log.Println("hello ", name)
	return milter.RespContinue, nil
}

func (f *Filter) MailFrom(from string, m *milter.Modifier) (milter.Response, error) {
	log.Println("mail from", from)
	return milter.RespContinue, nil
}

func (f *Filter) RcptTo(rcptTo string, m *milter.Modifier) (milter.Response, error) {
	log.Println("rcpt to", rcptTo)
	return milter.RespContinue, nil
}

func (f *Filter) Header(name string, value string, m *milter.Modifier) (milter.Response, error) {
	log.Println("header", name, value)
	return milter.RespContinue, nil
}

func (f *Filter) Headers(h textproto.MIMEHeader, m *milter.Modifier) (milter.Response, error) {
	log.Println("header", h)
	return milter.RespContinue, nil
}

func (f *Filter) BodyChunk(chunk []byte, m *milter.Modifier) (milter.Response, error) {

	log.Printf("body chunk %s\n", chunk)
	return milter.RespContinue, nil
}

func (f *Filter) Body(m *milter.Modifier, macro map[string]string) (milter.Response, error) {
	log.Println("macro", macro)
	if v, ok := macro["i"]; ok {
		m.AddHeader("X-Queue-ID", v)
	}
	log.Println("body")
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
