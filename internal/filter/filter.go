package filter

import (
	"easymail/internal/milter"
	"log"
	"net"
	"net/textproto"
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
	protocol   string
	listenAddr string
	listener   net.Listener
	filter     *Filter
}

func New(protocol string, listenAddr string, optAction milter.OptAction, optProtocol milter.OptProtocol) (server *Server, err error) {
	server = &Server{
		filter: &Filter{
			optProtocol: optProtocol,
			optAction:   optAction,
		},
	}
	server.listener, err = net.Listen(protocol, listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	return server, err
}

func (s *Server) Run() {
	for {
		client, err := s.listener.Accept()
		if err != nil {
			log.Println(err)
		}
		session := milter.NewSession(client, s.filter, s.filter.optAction, s.filter.optProtocol)
		go session.Handle()
	}
}
