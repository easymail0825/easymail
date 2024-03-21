package lmtp

import (
	"bytes"
	"net/textproto"
)

type session struct {
	helo         []byte
	sender       []byte
	receipts     [][]byte
	conn         *textproto.Conn
	data         *bytes.Buffer
	commandStage commandStage
	clientStage  clientStage
}

func (s *session) reset() {
	s.helo = nil
	s.sender = nil
	s.receipts = make([][]byte, 0)
	s.data.Reset()
}

func (s *session) writeResponse(response smtpResponse) error {
	if response.class <= 0 {
		for i := 0; i < len(response.message)-1; i++ {
			if err := s.conn.PrintfLine("%d-%v", response.code, response.message[i]); err != nil {
				return err
			}
		}
		err := s.conn.PrintfLine("%d %v", response.code, response.message[len(response.message)-1])
		if err != nil {
			return err
		}
	} else {
		err := s.conn.PrintfLine("%d %v.%v.%v %v", response.code, response.class, response.subject, response.detail, response.message[len(response.message)-1])
		if err != nil {
			return err
		}
	}
	return nil
}
