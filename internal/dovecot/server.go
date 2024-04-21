package dovecot

import (
	"bufio"
	"bytes"
	"easymail/internal/account"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

/*
Server dovecot server, accept client and handle authorize
*/
type Server struct {
	family  string
	address string
	// atomic counter for cuid
	count int64
}

func New(family, address string) *Server {
	if family != "tcp" && family != "unix" {
		return nil
	}

	fields := strings.Split(address, ":")
	if len(fields) != 2 {
		return nil
	}

	return &Server{
		family:  family,
		address: address,
	}
}

func (s *Server) Run() (err error) {
	listener, err := net.Listen(s.family, s.address)
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
				log.Println("handle connection failed:", err)
			}
		}()
	}
}

func (s *Server) Handle(conn net.Conn) (err error) {
	atomic.AddInt64(&s.count, 1)

	log.Printf("new client connected, from %s\n", conn.RemoteAddr().String())
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Println("close connection failed:", err)
		}
	}(conn)

	// handle client
	// create a new client session, and wrap it with scanner
	// we can read by line
	sess := NewSession(conn)
	scanner := bufio.NewScanner(sess.socket)

	// step 1: receive handShake
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, "\t")
		if len(tokens) < 2 {
			return errors.New(fmt.Sprintf("invalid command, %s", line))
		}

		cmd := tokens[0]
		argv := tokens[1:]
		switch cmd {
		case "VERSION":
			if argv[0] != "1" {
				return errors.New(fmt.Sprintf("invalid version: %s", argv[0]))
			}
		case "CPID":
			sess.cpid = argv[0]
			sess.handshakeOk = true
		}
		if sess.handshakeOk {
			log.Printf("handshake recieved, cpid: %s\n", sess.cpid)
			break
		}
	}
	if err = scanner.Err(); err != nil {
		return errors.New(fmt.Sprintf("read error"))
	}

	if !sess.handshakeOk {
		return errors.New(fmt.Sprintf("handshake failed"))
	}

	// step 2: response handshake, must keep this order
	if err = sess.sendLine("VERSION", "1", "1"); err != nil {
		return errors.New(fmt.Sprintf("send VERSION failed"))
	}
	if err = sess.sendLine("MECH", "PLAIN", "plaintext"); err != nil {
		return errors.New(fmt.Sprintf("send MECH failed"))
	}
	if err = sess.sendLine("SPID", strconv.Itoa(os.Getgid())); err != nil {
		return errors.New(fmt.Sprintf("send SPID failed"))
	}
	if err = sess.sendLine("CUID", strconv.FormatInt(s.count, 10)); err != nil {
		return errors.New(fmt.Sprintf("send CUID failed"))
	}
	if err = sess.sendLine("COOKIE", sess.cookie); err != nil {
		return errors.New(fmt.Sprintf("send COOKIE failed"))
	}
	if err = sess.sendLine("DONE"); err != nil {
		return errors.New(fmt.Sprintf("send DONE failed"))
	}
	log.Println("handshake finished, session id ", sess.id)

	// step 3: process authorization, and continue read from client
	data := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			data["error"] = fmt.Sprintf("invalid command: %s", line)
			_ = sess.sendData(false, data)
			continue
		}

		cmd := fields[0]
		argv := fields[1:]
		if cmd != "AUTH" {
			continue
		}
		if len(argv) < 4 {
			data["error"] = "invalid arguments"
			_ = sess.sendData(false, data)
			continue
		}
		sess.id = argv[0]
		if argv[1] != "PLAIN" {
			data["error"] = "invalid auth type"
			_ = sess.sendData(false, data)
			continue
		}

		// get auth parameters
		param := make(map[string]string)
		for _, v := range argv[2:] {
			d := strings.SplitN(v, "=", 2)
			if len(d) == 1 {
				param[d[0]] = ""
			} else if len(d) == 2 {
				param[d[0]] = d[1]
			}
		}

		// param must exists
		if _, ok := param["service"]; !ok {
			data["error"] = "service not found"
			_ = sess.sendData(false, data)
			continue
		}
		if _, ok := param["resp"]; !ok {
			data["error"] = "resp not found"
			_ = sess.sendData(false, data)
			continue
		}

		// decode username and password from resp
		auth, err := base64.StdEncoding.DecodeString(param["resp"])
		if err != nil {
			data["error"] = "failed to decode resp"
			_ = sess.sendData(false, data)
			continue
		}
		authPair := bytes.SplitN(auth, []byte{0}, 3)
		if len(authPair) != 3 {
			data["error"] = "invalid auth resp"
			_ = sess.sendData(false, data)
			continue
		}
		sess.username = string(authPair[1])
		password := string(authPair[2])

		// authorize username with the password
		_, err = account.Authorize(sess.username, password)
		if err != nil {
			data["error"] = "invalid username or password"
			_ = sess.sendData(false, data)
			continue
		}

		// auth successfully
		sess.done = true
		delete(data, "error")
		data["user"] = sess.username
		_ = sess.sendData(true, data)
		log.Println("authorized:", sess.username)
	}
	return nil
}
