package policy

import (
	"bufio"
	"easymail/internal/account"
	"log"
	"net"
	"strings"
)

type CheckPolicyServer struct {
	family         string
	address        string
	accountService *account.Service
}

func NewCheckPolicyServer(family, address string) *CheckPolicyServer {
	return &CheckPolicyServer{
		family:         family,
		address:        address,
		accountService: account.NewService(),
	}
}

func (s *CheckPolicyServer) Run() error {
	listener, err := net.Listen(s.family, s.address)
	if err != nil {
		return err
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Println("Error closing listener:", err)
		}
	}(listener)
	log.Println("check policy server waiting for connections...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("Accepted connection from %s\n", conn.RemoteAddr())
		go s.handleClient(conn)
	}
}

func (s *CheckPolicyServer) handleClient(conn net.Conn) {
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Println("Error closing connection:", err)
		}
	}(conn)

	var err error
	var sender string
	var recipient string

	// read all content from the client
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		log.Println("Received line from client:", line)
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
		log.Println("Error reading from client:", err)
		return
	}

	// then send the CPS response, action=dunno is allow, and action=reject is rejected
	log.Printf("found sender %s and recipient %s\n", sender, recipient)
	if recipient == "" {
		_, err = conn.Write([]byte("action=reject\n\n"))
	} else {
		_, err = s.accountService.FindAccountByName(recipient)
		if err == nil {
			_, err = conn.Write([]byte("action=dunno\n\n"))
		} else {
			_, err = conn.Write([]byte("action=reject\n\n"))
		}
	}

	if err != nil {
		log.Println("Error sending response to client:", err)
		return
	}
}
