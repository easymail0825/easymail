package policy

import (
	"bufio"
	"context"
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/observability/sessiontrace"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Server struct {
	name    string
	stopCh  chan struct{}
	started bool
	lock    *sync.Mutex
	family  string
	listen  string
	debug   bool
	_log    *easylog.Logger

	tracer sessiontrace.Tracer
}

func New(family, listen string) *Server {
	if family != "tcp" && family != "unix" {
		return nil
	}

	fields := strings.Split(listen, ":")
	if len(fields) != 2 {
		return nil
	}

	return &Server{
		name:    "policy",
		stopCh:  make(chan struct{}),
		lock:    &sync.Mutex{},
		started: false,
		family:  family,
		listen:  listen,
	}
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

func (s *Server) SetTracer(t sessiontrace.Tracer) {
	s.tracer = t
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

func (s *Server) run() error {
	listener, err := net.Listen(s.family, s.listen)
	if err != nil {
		return err
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			s._log.Errorf("Error closing listener: %+v\n", err)
		}
	}(listener)
	for {
		conn, err := listener.Accept()
		if err != nil {
			s._log.Errorf("Error accepting connection:%+v", err)
			continue
		}
		if s.debug {
			s._log.Debugf("Policy Server accepted connection from %s\n", conn.RemoteAddr())
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	var span sessiontrace.Session
	if s.tracer != nil {
		span = s.tracer.NewSession(context.Background(), sessiontrace.SessionMeta{
			Protocol: sessiontrace.ProtocolPolicy,
			Remote:   conn.RemoteAddr().String(),
			Local:    conn.LocalAddr().String(),
		})
		span.Event("connect", nil)
		defer span.End("disconnect", nil)
	}
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			s._log.Errorf("Error closing connection:%+v", err)
		}
	}(conn)

	var err error
	var sender string
	var recipient string

	// read all content from the client
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			if parts[0] == "sender" {
				sender = parts[1]
			}
			if parts[0] == "recipient" {
				recipient = parts[1]
			}
		}
	}

	if err = scanner.Err(); err != nil {
		s._log.Errorf("Error reading from client:%+v", err)
		if span != nil {
			span.Error("read", err, nil)
		}
		return
	}

	// then send the CPS response, action=dunno is allow, and action=reject is rejected
	action := "reject"
	defer func() {
		_, _ = conn.Write([]byte("action=" + action + "\n\n"))
		s._log.Infof("policy check done, sender=%s, recipient=%s, action=%s\n", sender, recipient, action)
		if span != nil {
			span.Event("decision", map[string]any{
				"sender":    sessiontrace.MaskEmail(sender),
				"recipient": sessiontrace.MaskEmail(recipient),
				"action":    action,
			})
		}
	}()

	if sender == "" {
		return
	}
	if _, err = model.FindAccountByName(sender); err != nil {
		return
	}

	if recipient == "" {
		return
	}

	recipient = strings.ToLower(recipient)
	d := strings.SplitN(recipient, "@", 2)
	if len(d) != 2 {
		return
	}
	_, domainErr := model.FindDomainByName(d[1])

	// internal domain mail must target existing account; external mail is allowed.
	if domainErr == nil {
		if _, err = model.FindAccountByName(recipient); err != nil {
			return
		}
	}
	action = "dunno"
}
