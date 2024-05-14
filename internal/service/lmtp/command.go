package lmtp

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

/*
smtpResponse
is a struct that contains the response code, message, including enhanced code.

status-code = class "." subject "." detail
class = "2"/"4"/"5"
subject = 1*3digit
detail = 1*3digit

subject classified
X.0.XXX Other or Undefined Status
X.1.XXX Addressing Status
X.2.XXX Mailbox Status
X.3.XXX Mail System Status
X.4.XXX Network and Routing Status
X.5.XXX Mail Delivery Family Status
X.6.XXX Message Content or Media Status
X.7.XXX Security or Policy Status

Enumerated Status Codes
X.0.0   Other undefined Status

X.1.0   Other address status
X.1.1   Bad destination mailbox address
X.1.2   Bad destination system address
X.1.3   Bad destination mailbox address syntax
X.1.4   Destination mailbox address ambiguous
X.1.5   Destination mailbox address valid
X.1.6   Destination mailbox has moved, No forwarding address
X.1.7   Bad sender's mailbox address syntax
X.1.8   Bad sender's system address

X.2.0   Other or undefined mailbox status
X.2.1   Mailbox disabled, not accepting messages
X.2.2   Mailbox full
X.2.3   Message length exceeds administrative limit
X.2.4   Mailing list expansion problem

X.3.0   Other or undefined mail system status
X.3.1   Mail system full
X.3.2   System not accepting network messages
X.3.3   System not capable of selected features
X.3.4   Message too big for system
X.3.5   System incorrectly configured

X.4.0   Other or undefined network or routing status
X.4.1   No answer from host
X.4.2   Bad connection
X.4.3   Directory server failure
X.4.4   Unable to route message
X.4.5   Mail system congestion
X.4.6   Routing loop detected
X.4.7   Delivery time expired

X.5.0   Other or undefined protocol status
X.5.1   Invalid command
X.5.2   Syntax error
X.5.3   Too many recipients
X.5.4   Invalid command arguments
X.5.5   Command incompatible with protocol state

X.6.0   Other or undefined media error
X.6.1   Media not supported
X.6.2   Conversion required and prohibited
X.6.3   Conversion required but not supported
X.6.4   Conversion with loss performed
X.6.5   Conversion failed

X.7.0   Other or undefined security status
X.7.1   Delivery not authorized, message refused
X.7.2   Mailing list expansion prohibited
X.7.3   Security conversion required but not possible
X.7.4   Security features not supported
X.7.5   Cryptographic failure
X.7.6   Cryptographic algorithm not supported
X.7.7   Message integrity failure
*/
type smtpResponse struct {
	code    int
	message []string
	class   int
	subject int
	detail  int
}

/*
parseCommand Parse a smtp command line sent by client into command and argument.
*/
func parseCommand(line string) (cmd string, arg string, err error) {
	line = strings.TrimRight(line, "\r\n")

	l := len(line)
	switch {
	case strings.HasPrefix(strings.ToUpper(line), "STARTTLS"):
		return "STARTTLS", "", nil
	case l == 0:
		return "", "", nil
	case l < 4:
		return "", "", fmt.Errorf("command too short: %q", line)
	case l == 4:
		return strings.ToUpper(line), "", nil
	case l == 5:
		return "", "", fmt.Errorf("wrong command: %q", line)
	}

	// If we made it here, command is long enough to have args
	if line[4] != ' ' {
		// There wasn't a space after the command?
		return "", "", fmt.Errorf("wrong command: %q", line)
	}

	// mail from:
	cmd = strings.ToUpper(line[0:4])
	if cmd == "MAIL" {
		i := 4
		for ; i < l; i++ {
			if line[i] != ' ' && line[i] != '\t' {
				break
			}
		}
		if l < i+5 || strings.ToUpper(line[i:i+5]) != "FROM:" {
			return "", "", fmt.Errorf("wrong command: %q", line)
		}
		return cmd, strings.TrimSpace(line[i+5:]), nil
	}

	// rcpt to:
	if cmd == "RCPT" {
		i := 4
		for ; i < l; i++ {
			if line[i] != ' ' && line[i] != '\t' {
				break
			}
		}
		if l < i+2 || strings.ToUpper(line[i:i+3]) != "TO:" {
			return "", "", fmt.Errorf("wrong command: %q", line)
		}
		return cmd, strings.TrimSpace(line[i+3:]), nil
	}

	return cmd, strings.TrimSpace(line[5:]), nil
}

func parseMailbox(m []byte) (mailbox []byte, err error) {
	start := 0
	end := 0
	for i := 0; i < len(m); i++ {
		if m[i] == ' ' {
			continue
		} else if m[i] == '<' {
			start = i
		} else if m[i] == '>' {
			end = i
			break
		}
	}
	if start > 2 {
		return mailbox, errors.New("start with too many blank")
	}
	if start >= 0 && end > 0 && start < end {
		mailbox = m[start+1 : end]
		return
	}
	return mailbox, errors.New("syntax error")
}

func handleConnect(sess *session) smtpResponse {
	return smtpResponse{
		code:    220,
		message: []string{"lmtp server ready"},
	}
}

func handleHelo(sess *session, arg string, hostname string, extension []string) smtpResponse {
	message := []string{hostname}
	message = append(message, extension...)
	message = append(message, "OK")

	return smtpResponse{
		code:    250,
		message: message,
	}
}

func handleHelp(sess *session) smtpResponse {
	return smtpResponse{
		code:    250,
		message: []string{"ok"},
	}
}

func handleMail(sess *session, arg string) smtpResponse {
	if sess.commandStage < commandStageHELO {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  5,
			message: []string{"Command incompatible with protocol state, helo first"},
		}
	}

	sender, err := parseMailbox([]byte(arg))
	if err != nil {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  1,
			message: []string{"Invalid command"},
		}
	}
	if len(sender) > 0 && !emailRegex.MatchString(string(sender)) {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  2,
			message: []string{"Syntax error"},
		}
	}
	sess.sender = sender
	log.Printf("sender is %s\n", sender)
	return smtpResponse{
		code:    250,
		message: []string{"mail ok"},
	}
}

func handleRcpt(sess *session, arg string) smtpResponse {
	if sess.commandStage < commandStageMail {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  5,
			message: []string{"Command incompatible with protocol state, mail first"},
		}
	}

	rcpt, err := parseMailbox([]byte(arg))
	if err != nil {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  1,
			message: []string{"Invalid command"},
		}
	}

	if len(rcpt) == 0 {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  4,
			message: []string{"Invalid command arguments"},
		}
	}

	if len(rcpt) > 100 {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  3,
			message: []string{"Too many recipients"},
		}
	}

	if !emailRegex.MatchString(string(rcpt)) {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  2,
			message: []string{"Syntax error"},
		}
	}
	log.Printf("rcpt is %s\n", rcpt)
	sess.receipts = append(sess.receipts, rcpt)
	return smtpResponse{
		code:    250,
		message: []string{"rcpt ok"},
	}
}

func handleData(sess *session) smtpResponse {
	if sess.commandStage < commandStageRcpt {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  5,
			message: []string{"Command incompatible with protocol state, rcpt first"},
		}
	}

	if len(sess.receipts) == 0 {
		return smtpResponse{
			code:    550,
			class:   5,
			subject: 5,
			detail:  4,
			message: []string{"Invalid command arguments"},
		}
	}
	return smtpResponse{
		code:    354,
		message: []string{"End data with <CR><LF>.<CR><LF>"},
	}
}

func handleRset(sess *session) smtpResponse {
	sess.reset()
	return smtpResponse{
		code:    250,
		message: []string{"reset ok"},
	}
}

/*
The DATA command

In the LMTP protocol, there is one additional restriction placed on
the DATA command, and one change to how replies to the final "." are
sent.

The additional restriction is that when there have been no successful
RCPT commands in the mail transaction, the DATA command MUST fail
with a 503 reply code.

The change is that after the final ".", the server returns one reply
for each previously successful RCPT command in the mail transaction,
in the order that the RCPT commands were issued.  Even if there were
multiple successful RCPT commands giving the same forward-path, there
must be one reply for each successful RCPT command.

When one of these replies to the final "." is a Positive Completion
reply, the server is accepting responsibility for delivering or
relying the message to the corresponding recipient.  It must take
this responsibility seriously, i.e., it MUST NOT lose the message for
frivolous reasons, e.g., because the host later crashes or because of
a predictable resource shortage.

A multiline reply is still considered a single reply and corresponds
to a single RCPT command.



*/
