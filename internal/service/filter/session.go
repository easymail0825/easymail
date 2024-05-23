package filter

import (
	"bufio"
	"bytes"
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/service/milter"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"io"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

var errCloseSession = errors.New("stop current milter processing")
var serverProtocolVersion uint32 = 2

type AntispamResult struct {
	Action model.AntispamAction `json:"action"`
	RuleID int                  `json:"ruleID"`
}

// Session keeps Session state during MTA communication
type Session struct {
	actions  milter.OptAction
	protocol milter.OptProtocol
	conn     net.Conn
	headers  textproto.MIMEHeader
	macros   map[string]string
	filter   *Filter
	features []milter.Feature
	_log     *easylog.Logger
}

func NewSession(conn net.Conn, filter *Filter, _log *easylog.Logger) *Session {
	return &Session{
		actions:  0,
		protocol: 0,
		conn:     conn,
		filter:   filter,
		macros:   make(map[string]string),
		headers:  make(textproto.MIMEHeader),
		features: make([]milter.Feature, 0),
		_log:     _log,
	}
}

// ReadPacket reads incoming milter packet
func (s *Session) ReadPacket() (*milter.Message, error) {
	return readPacket(s.conn, 0)
}

func readPacket(conn net.Conn, timeout time.Duration) (*milter.Message, error) {
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
	message := milter.Message{
		Code: data[0],
		Data: data[1:],
	}
	return &message, nil
}

// WritePacket sends a milter response packet to socket stream
func (s *Session) WritePacket(msg *milter.Message) error {
	return writePacket(s.conn, msg, 0)
}

func writePacket(conn net.Conn, msg *milter.Message, timeout time.Duration) error {
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

// NewModifier creates a new Modifier instance from Session
func (s *Session) newModifier() *milter.Modifier {
	return &milter.Modifier{
		Macros:      s.macros,
		Headers:     s.headers,
		WritePacket: s.WritePacket,
	}
}

// Process processes incoming milter commands
func (s *Session) Process(msg *milter.Message) (milter.Response, []milter.Feature, error) {
	switch milter.Code(msg.Code) {
	case milter.CodeAbort:
		fmt.Println("abort")
		s.headers = nil
		s.macros = nil
		s.features = nil
		return nil, nil, nil

	case milter.CodeBody:
		fmt.Println("body")
		return s.filter.BodyChunk(msg.Data, s.newModifier())

	case milter.CodeConn:
		// new connection, get hostname
		fmt.Println("conn")
		hostname := milter.ReadCString(msg.Data)
		msg.Data = msg.Data[len(hostname)+1:]
		// get protocol family
		protocolFamily := msg.Data[0]
		msg.Data = msg.Data[1:]
		// get port
		var port uint16
		if protocolFamily == '4' || protocolFamily == '6' {
			if len(msg.Data) < 2 {
				return milter.RespTempFail, nil, nil
			}
			port = binary.BigEndian.Uint16(msg.Data)
			msg.Data = msg.Data[2:]
		}
		// get address
		address := milter.ReadCString(msg.Data)
		// convert address and port to human-readable string
		family := map[byte]string{
			'U': "unknown",
			'L': "unix",
			'4': "tcp4",
			'6': "tcp6",
		}
		// run handler and return
		return s.filter.Connect(
			hostname,
			family[protocolFamily],
			port,
			net.ParseIP(address),
			s.newModifier())

	case milter.CodeMacro:
		// define macros
		fmt.Println("macro")
		s.macros = make(map[string]string)
		// convert data to Go strings
		data := milter.DecodeCStrings(msg.Data[1:])
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
		return nil, nil, nil

	case milter.CodeEOB:
		// call and return milter handler
		fmt.Println("eob")
		return s.filter.Body(s.newModifier(), s.macros)

	case milter.CodeHelo:
		// helo command
		fmt.Println("helo")
		name := strings.TrimSuffix(string(msg.Data), milter.Null)
		return s.filter.Helo(name, s.newModifier())

	case milter.CodeHeader:
		// make sure headers is initialized
		fmt.Println("header")
		if s.headers == nil {
			s.headers = make(textproto.MIMEHeader)
		}
		// add new header to headers map
		headerData := milter.DecodeCStrings(msg.Data)
		// headers with an empty body appear as `text\x00\x00`, decodeCStrings will drop the empty body
		if len(headerData) == 1 {
			headerData = append(headerData, "")
		}
		if len(headerData) == 2 {
			s.headers.Add(headerData[0], headerData[1])
			// call and return milter handler
			return s.filter.Header(headerData[0], headerData[1], s.newModifier())
		}

	case milter.CodeMail:
		// envelope from address
		fmt.Println("mail")
		from := milter.ReadCString(msg.Data)
		return s.filter.MailFrom(strings.Trim(from, "<>"), s.newModifier())

	case milter.CodeEOH:
		// end of headers
		fmt.Println("eoh")
		return s.filter.Headers(s.headers, s.newModifier())

	case milter.CodeOptNeg:
		// ignore request and prepare response buffer
		fmt.Println("optneg")
		var buffer bytes.Buffer
		// prepare response data
		for _, value := range []uint32{serverProtocolVersion, uint32(s.actions), uint32(s.protocol)} {
			if err := binary.Write(&buffer, binary.BigEndian, value); err != nil {
				return nil, nil, err
			}
		}
		// build and send packet
		return milter.NewResponse('O', buffer.Bytes()), nil, nil

	case milter.CodeQuit:
		// client requested session close
		fmt.Println("quit")
		return nil, nil, errCloseSession

	case milter.CodeRcpt:
		// envelope to address
		fmt.Println("rcpt")
		to := milter.ReadCString(msg.Data)
		return s.filter.RcptTo(strings.Trim(to, "<>"), s.newModifier())

	case milter.CodeData:
		// data, ignore
		fmt.Println("data")

	default:
		// print error and close session
		log.Printf("Unrecognized command code: %c", msg.Code)
		return nil, nil, errCloseSession
	}

	// by default continue with next milter message
	return milter.RespContinue, nil, nil
}

// Handle processes all milter commands in the same connection
func (s *Session) Handle() {
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

		resp, features, err := s.Process(msg)
		if err != nil {
			if err != errCloseSession {
				// log error condition
				log.Printf("Error performing milter command: %v\n", err)
			}
			return
		} else {
			for _, f := range features {
				s.features = append(s.features, f)
			}
		}

		// check rules
		fmt.Println("features is", s.features)
		result := AntispamResult{}
		if knowledgeInstance != nil {
			featureJson, err := Feature2Json(s.features)
			if err == nil {
				fmt.Printf("featureJson is %s\n", featureJson)
				dataContext := ast.NewDataContext()
				if err := dataContext.AddJSON("feature", featureJson); err == nil {
					if err := dataContext.Add("antispam", &result); err == nil {
						fmt.Println("add ok")
						// execute rules
						err = ruleEngine.Execute(dataContext, knowledgeInstance)
						if err != nil {
							panic(err)
						}
					}
				}
			}
		}
		fmt.Printf("result is %v\n", result)
		if result.RuleID > 0 {
			switch result.Action {
			case model.AntispamActionAccept:
				resp = milter.RespAccept
			case model.AntispamActionTrash:
				resp = milter.RespAccept
			case model.AntispamActionDefer:
				resp = milter.RespTempFail
			case model.AntispamActionReject:
				resp = milter.RespReject
			case model.AntispamActionDiscard:
				resp = milter.RespDiscard
			case model.AntispamActionQuarantine:
				resp = milter.RespReject
			}
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
				s.filter = nil
				return
			}
		}
	}
}

func Feature2Json(features []milter.Feature) (jsonBytes []byte, err error) {
	FeatureMap := make(map[string]any)
	for _, f := range features {
		name := strings.Replace(f.Name, "-", "_", -1)
		switch f.ValueType {
		case milter.DataTypeString:
			FeatureMap[name] = f.Value
		case milter.DataTypeInt:
			FeatureMap[name], _ = strconv.Atoi(f.Value)
		case milter.DataTypeFloat:
			FeatureMap[name], _ = strconv.ParseFloat(f.Value, 64)
		case milter.DataTypeBool:
			FeatureMap[name], _ = strconv.ParseBool(f.Value)
		}
	}

	bs, err := json.Marshal(FeatureMap)
	if err != nil {
		return nil, err
	}

	return bs, nil
}
