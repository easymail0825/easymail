package dovecot

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"strings"
)

/*
Session client session, store client info
*/
type Session struct {
	socket      net.Conn
	cpid        string
	handshakeOk bool
	cookie      string
	id          string
	username    string
	done        bool
}

/*
NewSession create new session from server accept loop
*/
func NewSession(socket net.Conn) *Session {
	return &Session{
		socket: socket,
		cookie: uuid.New().String(),
	}
}

/*
sendLine send line to client
*/
func (s *Session) sendLine(argv ...string) error {
	resp := strings.Join(argv, "\t")
	_, err := s.socket.Write([]byte(resp + "\n"))
	return err
}

/*
sendData send data to client, indicate success or fail
*/
func (s *Session) sendData(success bool, data map[string]string) error {
	result := "FAIL"
	if success {
		result = "OK"
	}
	resp := []string{
		result,
		s.id,
	}
	for k, v := range data {
		resp = append(resp, fmt.Sprintf("%s=%s", k, v))
	}
	return s.sendLine(resp...)
}
