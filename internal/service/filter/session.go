package filter

import (
	"bufio"
	"bytes"
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/preprocessing"
	"easymail/internal/service/milter"
	"easymail/vender/spf"
	"easymail/vender/ssdeep"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/jhillyerd/enmime"
	"io"
	"log"
	"net"
	"net/mail"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

var errCloseSession = errors.New("stop current milter processing")
var serverProtocolVersion uint32 = 2

const minHashTextSize = 128
const sep = "\r\n"

type Result struct {
	Action model.FilterAction `json:"action"`
	RuleID int                `json:"ruleID"`
}

// Session keeps Session state during MTA communication
type Session struct {
	id         uuid.UUID
	actions    milter.OptAction
	protocol   milter.OptProtocol
	conn       net.Conn
	headers    textproto.MIMEHeader
	macros     map[string]string
	filter     *Filter
	features   []milter.Feature
	payload    map[string]string
	headerData []byte // mail headerData
	bodyData   []byte // mail bodyData
	_log       *easylog.Logger
	html2Text  *preprocessing.Html2Text
}

func NewSession(conn net.Conn, filter *Filter, _log *easylog.Logger, html2Text *preprocessing.Html2Text) *Session {
	return &Session{
		id:         uuid.New(),
		actions:    0,
		protocol:   0,
		conn:       conn,
		filter:     filter,
		macros:     make(map[string]string),
		headers:    make(textproto.MIMEHeader),
		features:   make([]milter.Feature, 0),
		payload:    make(map[string]string),
		headerData: make([]byte, 0),
		bodyData:   make([]byte, 0),
		_log:       _log,
		html2Text:  html2Text,
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

	case milter.CodeOptNeg:
		fmt.Println("optneg")
		// ignore request and prepare response buffer
		var buffer bytes.Buffer
		// prepare response bodyData
		for _, value := range []uint32{serverProtocolVersion, uint32(s.actions), uint32(s.protocol)} {
			if err := binary.Write(&buffer, binary.BigEndian, value); err != nil {
				return nil, nil, err
			}
		}
		// build and send packet
		return milter.NewResponse('O', buffer.Bytes()), nil, nil

	case milter.CodeMacro:
		fmt.Println("macro")
		// define macros
		s.macros = make(map[string]string)
		// convert bodyData to Go strings
		data := milter.DecodeCStrings(msg.Data[1:])
		if len(data) != 0 {
			if len(data)%2 == 1 {
				data = append(data, "")
			}

			// store bodyData in a map
			for i := 0; i < len(data); i += 2 {
				s.macros[data[i]] = data[i+1]
			}
		}
		// do not send response
		return nil, nil, nil

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
		//address := fmt.Sprintf("%s:%d", family[protocolFamily], port)
		s.payload["Family"] = family[protocolFamily]
		s.payload["Port"] = fmt.Sprintf("%d", port)
		ip := net.ParseIP(address)

		// generate payload of the session
		if fields, err := model.GetFilterFieldByStage(model.FilterStageConnect); err == nil {
			for _, f := range fields {
				switch f.Name {
				case "ClientIP":
					s.payload["ClientIP"] = address
					break
				case "PTR":
					ptr, err := resolver.LookupPtr(address)
					if err == nil {
						s.payload["PTR"] = strings.Join(ptr, sep)
					}
				case "Region":
					country, province, city, err := QueryRegion(ip)
					if err == nil {
						s.payload["Country"] = country
						s.payload["Province"] = province
						s.payload["City"] = city
					}
				}
			}
		}

		// run handler and return
		return s.filter.Connect(
			hostname,
			ip,
			s.payload,
			s.newModifier())

	case milter.CodeHelo:
		fmt.Println("helo")
		// helo command
		name := strings.TrimSuffix(string(msg.Data), milter.Null)
		// generate payload of the session
		if fields, err := model.GetFilterFieldByStage(model.FilterStageHelo); err == nil {
			for _, f := range fields {
				switch f.Name {
				case "Helo":
					s.payload["Helo"] = name
				}
			}
		}
		return s.filter.Helo(name, s.payload, s.newModifier())

	case milter.CodeHeader:
		fmt.Println("header")
		// make sure headers is initialized
		if s.headers == nil {
			s.headers = make(textproto.MIMEHeader)
		}
		// add new header to headers map
		data := milter.DecodeCStrings(msg.Data)
		// headers with an empty body appear as `text\x00\x00`, decodeCStrings will drop the empty body
		if len(data) == 1 {
			data = append(data, "")
		}

		if len(data) == 2 {
			s.headerData = append(s.headerData, []byte(fmt.Sprintf("%s: %s\r\n", data[0], data[1]))...)
			s.headers.Add(data[0], data[1])
			// call and return milter handler
			return s.filter.Header(data[0], data[1], s.payload, s.newModifier())
		}

	case milter.CodeEOH:
		fmt.Println("eoh")
		// end of headers
		sb := strings.Builder{}
		for key, value := range s.headers {
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
		email, emailErr := enmime.ReadEnvelope(strings.NewReader(sb.String()))

		// generate payload of the session
		if fields, err := model.GetFilterFieldByStage(model.FilterStageHeader); err == nil {
			for _, f := range fields {
				switch f.Name {
				case "HeaderFrom":
					if emailErr == nil {
						if sender, err := mail.ParseAddress(email.GetHeader("From")); err == nil {
							s.payload["HeaderFrom"] = sender.Address
							s.payload["Nick"] = sender.Name
						}
					}
				case "Mailer":
					if emailErr == nil {
						s.payload["Mailer"] = email.GetHeader("X-Mailer")
					}
				case "Subject":
					if emailErr == nil {
						s.payload["Subject"] = email.GetHeader("Subject")
					}
				case "Dmarc":
					// todo
				case "DomainKey":
					//todo
				}
			}
		}
		return s.filter.Headers(s.headers, s.payload, s.newModifier())

	case milter.CodeMail:
		fmt.Println("mail")
		// envelope from address
		sender := milter.ReadCString(msg.Data)
		cleanSender := strings.Trim(sender, "<>")
		// generate payload of the session
		if fields, err := model.GetFilterFieldByStage(model.FilterStageMailFrom); err == nil {
			for _, f := range fields {
				switch f.Name {
				case "Sender":
					s.payload["Sender"] = cleanSender
				case "SPF":
					result, _ := spf.CheckHostWithSender(resolver, net.ParseIP(s.payload["ClientIP"]), s.payload["Helo"], sender)
					s.payload["SPF"] = string(result)
				}
			}
		}

		return s.filter.MailFrom(cleanSender, s.payload, s.newModifier())

	case milter.CodeRcpt:
		fmt.Println("rcpt")
		// envelope to address
		to := milter.ReadCString(msg.Data)
		cleanTo := strings.Trim(to, "<>")
		if fields, err := model.GetFilterFieldByStage(model.FilterStageRcptTo); err == nil {
			for _, f := range fields {
				switch strings.ToLower(f.Name) {
				case "rcpt":
					if len(s.payload["rcpt"]) == 0 {
						s.payload["Rcpt"] = cleanTo
					} else {
						s.payload["Rcpt"] = s.payload["Rcpt"] + "," + cleanTo
					}
				}
			}
		}
		return s.filter.RcptTo(cleanTo, s.payload, s.newModifier())

	case milter.CodeBody:
		fmt.Println("body")
		// append to mail bodyData array only
		s.bodyData = append(s.bodyData, msg.Data...)
		//return s.filter.BodyChunk(msg.Data, s.payload, s.newModifier())
		return milter.RespContinue, nil, nil

	case milter.CodeEOB:
		fmt.Println("eob")
		// parse the mail data
		buf := bytes.Buffer{}
		if len(s.headerData) == 0 {
			buf.Write([]byte("test: test"))
		} else {
			buf.Write(s.headerData)
		}
		buf.Write([]byte("\r\n"))
		buf.Write(s.bodyData)
		mailData := buf.Bytes()
		s.payload["Size"] = strconv.Itoa(len(mailData))
		mailObj, err := enmime.ReadEnvelope(bytes.NewReader(buf.Bytes()))

		if err != nil {
			return milter.RespContinue, nil, nil
		}
		text := mailObj.Text
		html := mailObj.HTML
		textParts, urls := s.html2Text.Parse(html)
		htmlText := strings.Join(textParts, "\n")
		if len(htmlText) > 0 {
			text = htmlText
		}
		s.payload["Text"] = text
		s.payload["Html"] = html
		s.payload["URL"] = strings.Join(urls, sep)

		// compute ssdeep hash
		hashText := ""
		if len(text) >= minHashTextSize {
			hashText = text
		} else if len(html) > minHashTextSize {
			hashText = html
		}
		if len(hashText) >= minHashTextSize {
			if h, err := ssdeep.FuzzyBytes([]byte(hashText)); err == nil {
				s.payload["TextHash"] = h
				if d := strings.SplitN(h, ":", 3); len(d) == 3 {
					if chunkSize, err := strconv.Atoi(d[0]); err == nil {
						if _, err1 := model.CreateSsdeepHash(h, s.id.String(), chunkSize, false); err1 != nil {
							s._log.Error("CreateSsdeepHash", err1)
						}
					}
				}
			} else {
				s.payload["TextHash"] = ""
			}
		}

		// Attachment List
		attachNames := make([]string, 0)
		attachHashes := make([]string, 0)
		attachMd5 := make([]string, 0)
		for _, att := range mailObj.Attachments {
			attachNames = append(attachNames, att.FileName)
			if h, err := computeAttachHash(att.Content); err == nil {
				attachHashes = append(attachHashes, h)
			}
			if m, err := computeAttachMD5(att.Content); err == nil {
				attachMd5 = append(attachMd5, m)
			}
		}

		// Inline List
		for _, att := range mailObj.Inlines {
			attachNames = append(attachNames, att.FileName)
			if h, err := computeAttachHash(att.Content); err == nil {
				attachHashes = append(attachHashes, h)
			}
			if m, err := computeAttachMD5(att.Content); err == nil {
				attachMd5 = append(attachMd5, m)
			}
		}

		// save attachment hashes
		for _, h := range attachHashes {
			if d := strings.SplitN(h, ":", 3); len(d) == 3 {
				if chunkSize, err := strconv.Atoi(d[0]); err == nil {
					if _, err1 := model.CreateSsdeepHash(h, s.id.String(), chunkSize, true); err1 != nil {
						s._log.Error("CreateSsdeepHash", err1)
					}
				}
			}
		}
		s.payload["AttachName"] = strings.Join(attachNames, sep)
		s.payload["AttachHash"] = strings.Join(attachHashes, sep)
		s.payload["AttachMd5"] = strings.Join(attachMd5, sep)

		// Other Part List
		for _, att := range mailObj.OtherParts {
			attachNames = append(attachNames, att.FileName)
		}

		fmt.Println("payload:", s.payload)

		// call and return milter handler
		return s.filter.Body(s.payload, s.newModifier(), s.macros)

	case milter.CodeData:
		// bodyData, ignore
		return nil, nil, nil

	case milter.CodeQuit:
		fmt.Println("quit")
		// client requested session close
		return nil, nil, errCloseSession

	case milter.CodeAbort:
		s.headers = nil
		s.macros = nil
		s.features = nil
		s.payload = nil
		return nil, nil, nil

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
		s._log = nil
		s.headers = nil
		s.macros = nil
		s.filter = nil
		s.features = nil
		s.payload = nil
		s.bodyData = nil
		s._log = nil
		s.html2Text = nil
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

		// check rules, ignore CodeOptNeg,CodeMacro,CodeHeader
		if !(msg.Code == 'O' || msg.Code == 'D' || msg.Code == 'L') {
			result := Result{}
			if knowledgeInstance != nil {
				featureJson, err := Feature2Json(s.features)
				if err == nil {
					dataContext := ast.NewDataContext()
					if err := dataContext.AddJSON("feature", featureJson); err == nil {
						if err := dataContext.Add("result", &result); err == nil {
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
				case model.FilterActionAccept:
					resp = milter.RespAccept
				case model.FilterActionTrash:
					resp = milter.RespAccept
				case model.FilterActionDefer:
					resp = milter.RespTempFail
				case model.FilterActionReject:
					resp = milter.RespReject
				case model.FilterActionDiscard:
					resp = milter.RespDiscard
				case model.FilterActionQuarantine:
					resp = milter.RespReject
				}
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
