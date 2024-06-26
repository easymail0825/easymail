package lmtp

import (
	"bytes"
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/service/storage"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type commandStage int

const (
	commandStageAccept commandStage = iota
	commandStageHELO
	commandStageMail
	commandStageRcpt
	commandStageData
	commandStageDotData
	commandStageQuit
)

type clientStage int

const (
	clientStageGreeting clientStage = iota
	clientStageCommand
	clientStageData
	clientStageShutdown
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@([a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}|localhost)$`)

type Server struct {
	name    string
	stopCh  chan struct{}
	started bool
	lock    *sync.Mutex
	family  string
	listen  string
	debug   bool
	_log    *easylog.Logger

	hostname      string
	messageLimit  int64
	extension     []string
	readTimeout   time.Duration
	writeTimeout  time.Duration
	handleConnect func(session *session) smtpResponse
	handleHelo    func(session *session, arg string, hostname string, extension []string) smtpResponse
	handleNoop    func(session *session) smtpResponse
	handleMail    func(session *session, arg string) smtpResponse
	handleRcpt    func(session *session, arg string) smtpResponse
	handleReset   func(session *session) smtpResponse
	handleHelp    func(session *session) smtpResponse
	handleData    func(session *session) smtpResponse
	handleRset    func(session *session) smtpResponse
	storager      storage.Storager
}

func New(family, listen string, messageLimit int64, extension ...string) *Server {
	if family != "tcp" && family != "unix" {
		return nil
	}

	fields := strings.Split(listen, ":")
	if len(fields) != 2 {
		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	if messageLimit > 0 {
		extension = append(extension, fmt.Sprintf("SIZE %d", messageLimit))
	}

	return &Server{
		name:   "lmtp",
		lock:   &sync.Mutex{},
		stopCh: make(chan struct{}),
		family: family,
		listen: listen,

		extension:     extension,
		hostname:      hostname,
		messageLimit:  1024 * 1024 * 50, // 50MB
		readTimeout:   time.Duration(300 * time.Second),
		writeTimeout:  time.Duration(300 * time.Second),
		handleConnect: handleConnect,
		handleHelo:    handleHelo,
		handleHelp:    handleHelp,
		handleMail:    handleMail,
		handleRcpt:    handleRcpt,
		handleData:    handleData,
		handleRset:    handleRset,
	}
}

func (s *Server) SetStorage(storager storage.Storager) {
	s.storager = storager
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

func (s *Server) run() error {
	listener, err := net.Listen(s.family, s.listen)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			err := s.Handle(conn)
			if err != nil {
				s._log.Errorf("handle connection failed:％+v\n", err)
			}
		}()
	}
}

func (s *Server) resetTimeout(conn net.Conn, readTimeout, writeTimeout time.Duration) {
	now := time.Now()
	if s.readTimeout > 0 {
		err := conn.SetReadDeadline(now.Add(s.readTimeout))
		if err != nil {
			s._log.Errorf("set read deadline failed:%+v\n", err)
		}
	}
	if s.writeTimeout > 0 {
		err := conn.SetWriteDeadline(now.Add(s.writeTimeout))
		if err != nil {
			s._log.Errorf("set write deadline failed:%+v\n", err)
		}
	}
}

// Handle process lmtp session
func (s *Server) Handle(conn net.Conn) (err error) {
	s.resetTimeout(conn, s.readTimeout, s.writeTimeout)
	if s.debug {
		s._log.Debugf("lmtp client connected, from %s\n", conn.RemoteAddr().String())
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	sess := &session{
		conn:         textproto.NewConn(conn),
		commandStage: commandStageAccept,
		clientStage:  clientStageGreeting,
		receipts:     make([][]byte, 0),
		data:         bytes.NewBuffer([]byte{}), // memory leaks?
	}
	reader := sess.conn.Reader
	writer := sess.conn.Writer

	for sess.clientStage != clientStageShutdown {
		switch sess.clientStage {
		case clientStageGreeting:
			_ = writer.PrintfLine("220 easymail smtp server")
			sess.clientStage = clientStageCommand
		case clientStageCommand:
			s.resetTimeout(conn, s.readTimeout, s.writeTimeout)
			var line []byte
			line, err = reader.ReadLineBytes()
			if err != nil {
				log.Printf("read error: %v\n", err)
				// then close the session
				sess.clientStage = clientStageShutdown
			}
			var cmd, arg string
			cmd, arg, err = parseCommand(string(line))
			if err != nil {
				log.Printf("parse command error: %v\n", err)
				resp := smtpResponse{
					code:    550,
					class:   5,
					subject: 5,
					detail:  2,
					message: []string{"Syntax error"},
				}
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
				}
				continue
			}
			switch cmd {
			case "HELO", "EHLO", "LHLO":
				resp := s.handleHelo(sess, arg, s.hostname, s.extension)
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
					continue
				}
				if resp.code >= 200 && resp.code < 400 {
					sess.commandStage = commandStageHELO
				}
			case "HELP":
				resp := s.handleHelp(sess)
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
					continue
				}
			case "MAIL":
				resp := s.handleMail(sess, arg)
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
					continue
				}
				if resp.code >= 200 && resp.code < 400 {
					sess.commandStage = commandStageMail
				}
			case "RCPT":
				resp := s.handleRcpt(sess, arg)
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
					continue
				}
				if resp.code >= 200 && resp.code < 400 {
					sess.commandStage = commandStageRcpt
				}
			case "RSET":
				resp := s.handleRset(sess)
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
					continue
				}
				if resp.code >= 200 && resp.code < 400 {
					sess.commandStage = commandStageHELO
				}
			case "LOOP":
				_ = writer.PrintfLine("250 noop ok")
			case "DATA":
				resp := s.handleData(sess)
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
					continue
				}
				if resp.code >= 200 && resp.code < 400 {
					sess.clientStage = clientStageData
					sess.commandStage = commandStageData
				}
			case "QUIT":
				_ = writer.PrintfLine("221 bye")
				err := sess.conn.Close()
				if err != nil {
					s._log.Errorf("write response error:%+v\n", err)
				}
				sess.clientStage = clientStageShutdown
				sess.commandStage = commandStageQuit
			default:
				resp := smtpResponse{
					code:    500,
					class:   5,
					subject: 5,
					detail:  1,
					message: []string{"Invalid command"},
				}
				if err := sess.writeResponse(resp); err != nil {
					s._log.Errorf("write response error:%+v\n", err)
				}
			}
		case clientStageData:
			sess.commandStage = commandStageDotData
			mailData, err := reader.ReadDotBytes()
			if err != nil {
				_ = writer.PrintfLine("550 receive data failed")
				break
			}
			sess.data.Write(mailData)

			// parse mailData to get subject and postfix queue id
			var email *model.Email
			email, err = ParseMail(mailData)
			if err != nil {
				s._log.Errorf("parse mail error:%+v\n", err)
				for _, receipt := range sess.receipts {
					_ = writer.PrintfLine("550 <%s> mail parse failed", receipt)
				}
			} else {
				// save mail and index, then reply to every receipt
				for _, receipt := range sess.receipts {
					email.Sender = string(sess.sender)
					email.Recipient = string(bytes.Join(sess.receipts, []byte(",")))
					filePath, err := s.storager.Save(string(receipt), email, sess.data)
					if err == nil {
						s._log.Infof("mail derived for %s successfully to %s\n", receipt, filePath)
						_ = writer.PrintfLine("250 <%s> mail ok", receipt)
					} else {
						s._log.Infof("mail derived for %s failed: %s\n", receipt, err)
						_ = writer.PrintfLine("550 <%s> mail delivered failed", receipt)
					}

				}
			}
			sess.reset()
			sess.clientStage = clientStageCommand
			s.resetTimeout(conn, s.readTimeout, s.writeTimeout)
		} // end switch
	} // end for
	return nil
}
