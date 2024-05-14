package admin

import (
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/service/admin/router"
	"encoding/gob"
	"fmt"
	"sync"
)

type Server struct {
	name           string
	stopCh         chan struct{}
	started        bool
	lock           *sync.Mutex
	family         string
	listen         string
	debug          bool
	_log           *easylog.Logger
	root           string
	cookiePassword string
	cookieTag      string
}

func New(family, listen, root, cookiePassword, cookieTag string) *Server {
	return &Server{
		name:           "admin",
		stopCh:         make(chan struct{}),
		started:        false,
		lock:           &sync.Mutex{},
		family:         family,
		listen:         listen,
		root:           root,
		cookiePassword: cookiePassword,
		cookieTag:      cookieTag,
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

func (s *Server) run() {
	//gin.SetMode(gin.ReleaseMode)
	gob.Register(model.Account{})
	r := router.New(s._log, s.root, s.cookiePassword, s.cookieTag)
	err := r.Run(s.listen)
	if err != nil {
		s._log.Errorf("%s server run error: %v", s.name, err)
	}
}
