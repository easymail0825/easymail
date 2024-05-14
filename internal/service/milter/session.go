package milter

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"net/textproto"
	"strings"
	"time"
)

var errCloseSession = errors.New("stop current milter processing")
var serverProtocolVersion uint32 = 2

// session keeps session state during MTA communication
type session struct {
	actions  OptAction
	protocol OptProtocol
	conn     net.Conn
	headers  textproto.MIMEHeader
	macros   map[string]string
	milter   Milter
}

func NewSession(conn net.Conn, milter Milter, actions OptAction, protocol OptProtocol) *session {
	return &session{
		actions:  0,
		protocol: 0,
		conn:     conn,
		milter:   milter,
		macros:   make(map[string]string),
		headers:  make(textproto.MIMEHeader),
	}
}

// ReadPacket reads incoming milter packet
func (s *session) ReadPacket() (*Message, error) {
	return readPacket(s.conn, 0)
}

func readPacket(conn net.Conn, timeout time.Duration) (*Message, error) {
	if timeout != 0 {
		err := conn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return nil, err
		}
		defer func(conn net.Conn, t time.Time) {
			err := conn.SetReadDeadline(t)
			if err != nil {
				log.Printf("SetReadDeadline error %v\n", err)
			}
		}(conn, time.Time{})
	}
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}
	message := Message{
		Code: data[0],
		Data: data[1:],
	}
	return &message, nil
}

// WritePacket sends a milter response packet to socket stream
func (s *session) WritePacket(msg *Message) error {
	return writePacket(s.conn, msg, 0)
}

func writePacket(conn net.Conn, msg *Message, timeout time.Duration) error {
	if timeout != 0 {
		err := conn.SetWriteDeadline(time.Now().Add(timeout))
		if err != nil {
			return err
		}
		defer func(conn net.Conn, t time.Time) {
			err := conn.SetWriteDeadline(t)
			if err != nil {
				log.Println("set write deadline:", err)
			}
		}(conn, time.Time{})
	}
	buffer := bufio.NewWriter(conn)
	length := uint32(len(msg.Data) + 1)
	if err := binary.Write(buffer, binary.BigEndian, length); err != nil {
		return err
	}
	if err := buffer.WriteByte(msg.Code); err != nil {
		return err
	}
	if _, err := buffer.Write(msg.Data); err != nil {
		return err
	}
	if err := buffer.Flush(); err != nil {
		return err
	}
	return nil
}

// Process processes incoming milter commands
func (s *session) Process(msg *Message) (Response, error) {
	switch Code(msg.Code) {
	case CodeAbort:
		s.headers = nil
		s.macros = nil
		return nil, nil

	case CodeBody:
		// body chunk
		return s.milter.BodyChunk(msg.Data, newModifier(s))

	case CodeConn:
		// new connection, get hostname
		hostname := readCString(msg.Data)
		msg.Data = msg.Data[len(hostname)+1:]
		// get protocol family
		protocolFamily := msg.Data[0]
		msg.Data = msg.Data[1:]
		// get port
		var port uint16
		if protocolFamily == '4' || protocolFamily == '6' {
			if len(msg.Data) < 2 {
				return RespTempFail, nil
			}
			port = binary.BigEndian.Uint16(msg.Data)
			msg.Data = msg.Data[2:]
		}
		// get address
		address := readCString(msg.Data)
		// convert address and port to human-readable string
		family := map[byte]string{
			'U': "unknown",
			'L': "unix",
			'4': "tcp4",
			'6': "tcp6",
		}
		// run handler and return
		return s.milter.Connect(
			hostname,
			family[protocolFamily],
			port,
			net.ParseIP(address),
			newModifier(s))

	case CodeMacro:
		// define macros
		s.macros = make(map[string]string)
		// convert data to Go strings
		data := decodeCStrings(msg.Data[1:])
		if len(data) != 0 {
			if len(data)%2 == 1 {
				data = append(data, "")
			}

			// store data in a map
			for i := 0; i < len(data); i += 2 {
				s.macros[data[i]] = data[i+1]
			}
		}
		// do not send response
		return nil, nil

	case CodeEOB:
		// call and return milter handler
		return s.milter.Body(newModifier(s), s.macros)

	case CodeHelo:
		// helo command
		name := strings.TrimSuffix(string(msg.Data), null)
		return s.milter.Helo(name, newModifier(s))

	case CodeHeader:
		// make sure headers is initialized
		if s.headers == nil {
			s.headers = make(textproto.MIMEHeader)
		}
		// add new header to headers map
		headerData := decodeCStrings(msg.Data)
		// headers with an empty body appear as `text\x00\x00`, decodeCStrings will drop the empty body
		if len(headerData) == 1 {
			headerData = append(headerData, "")
		}
		if len(headerData) == 2 {
			s.headers.Add(headerData[0], headerData[1])
			// call and return milter handler
			return s.milter.Header(headerData[0], headerData[1], newModifier(s))
		}

	case CodeMail:
		// envelope from address
		from := readCString(msg.Data)
		return s.milter.MailFrom(strings.Trim(from, "<>"), newModifier(s))

	case CodeEOH:
		// end of headers
		return s.milter.Headers(s.headers, newModifier(s))

	case CodeOptNeg:
		// ignore request and prepare response buffer
		var buffer bytes.Buffer
		// prepare response data
		for _, value := range []uint32{serverProtocolVersion, uint32(s.actions), uint32(s.protocol)} {
			if err := binary.Write(&buffer, binary.BigEndian, value); err != nil {
				return nil, err
			}
		}
		// build and send packet
		return NewResponse('O', buffer.Bytes()), nil

	case CodeQuit:
		// client requested session close
		return nil, errCloseSession

	case CodeRcpt:
		// envelope to address
		to := readCString(msg.Data)
		return s.milter.RcptTo(strings.Trim(to, "<>"), newModifier(s))

	case CodeData:
		// data, ignore

	default:
		// print error and close session
		log.Printf("Unrecognized command code: %c", msg.Code)
		return nil, errCloseSession
	}

	// by default continue with next milter message
	return RespContinue, nil
}

// Handle processes all milter commands in the same connection
func (s *session) Handle() {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v\n", err)
		}
	}(s.conn)

	for {
		msg, err := s.ReadPacket()
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading milter command: %v\n", err)
			}
			return
		}

		resp, err := s.Process(msg)
		if err != nil {
			if err != errCloseSession {
				// log error condition
				log.Printf("Error performing milter command: %v\n", err)
			}
			return
		}

		// ignore empty responses
		if resp != nil {
			// send back response message
			if err = s.WritePacket(resp.Response()); err != nil {
				log.Printf("Error writing packet: %v", err)
				return
			}

			if !resp.Continue() {
				// prepare milter for next message
				s.milter = nil
				return
			}
		}
	}
}
