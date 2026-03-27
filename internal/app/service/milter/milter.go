package milter

import (
	"net"
	"net/textproto"
)

// OptAction sets which actions the milter wants to perform.
// Multiple options can be set using a bitmask.
type OptAction uint32

// Set which actions the milter wants to perform.
const (
	OptAddHeader    OptAction = 1 << 0 // SMFIF_ADDHDRS
	OptChangeBody   OptAction = 1 << 1 // SMFIF_CHGBODY
	OptAddRcpt      OptAction = 1 << 2 // SMFIF_ADDRCPT
	OptRemoveRcpt   OptAction = 1 << 3 // SMFIF_DELRCPT
	OptChangeHeader OptAction = 1 << 4 // SMFIF_CHGHDRS
	OptQuarantine   OptAction = 1 << 5 // SMFIF_QUARANTINE

	// [v6]
	OptChangeFrom      OptAction = 1 << 6 // SMFIF_CHGFROM
	OptAddRcptWithArgs OptAction = 1 << 7 // SMFIF_ADDRCPT_PAR
	OptSetSymList      OptAction = 1 << 8 // SMFIF_SETSYMLIST
)

// OptProtocol masks out unwanted parts of the SMTP transaction.
// Multiple options can be set using a bitmask.
type OptProtocol uint32

type FeatureType uint8

const (
	DataTypeString FeatureType = iota
	DataTypeInt
	DataTypeFloat
	DataTypeBool
)

type Feature struct {
	Name      string
	Value     string
	ValueType FeatureType
}

const (
	OptNoConnect  OptProtocol = 1 << 0 // SMFIP_NOCONNECT
	OptNoHelo     OptProtocol = 1 << 1 // SMFIP_NOHELO
	OptNoMailFrom OptProtocol = 1 << 2 // SMFIP_NOMAIL
	OptNoRcptTo   OptProtocol = 1 << 3 // SMFIP_NORCPT
	OptNoBody     OptProtocol = 1 << 4 // SMFIP_NOBODY
	OptNoHeaders  OptProtocol = 1 << 5 // SMFIP_NOHDRS
	OptNoEOH      OptProtocol = 1 << 6 // SMFIP_NOEOH
	OptNoUnknown  OptProtocol = 1 << 8 // SMFIP_NOUNKNOWN
	OptNoData     OptProtocol = 1 << 9 // SMFIP_NODATA

	// [v6] MTA supports ActSkip
	OptSkip OptProtocol = 1 << 10 // SMFIP_SKIP
	// [v6] Filter wants rejected RCPTs
	OptRcptRej OptProtocol = 1 << 11 // SMFIP_RCPT_REJ

	// Milter will not send action response for the following MTA messages
	OptNoHeaderReply OptProtocol = 1 << 7 // SMFIP_NR_HDR, SMFIP_NOHREPL
	// [v6]
	OptNoConnReply    OptProtocol = 1 << 12 // SMFIP_NR_CONN
	OptNoHeloReply    OptProtocol = 1 << 13 // SMFIP_NR_HELO
	OptNoMailReply    OptProtocol = 1 << 14 // SMFIP_NR_MAIL
	OptNoRcptReply    OptProtocol = 1 << 15 // SMFIP_NR_RCPT
	OptNoDataReply    OptProtocol = 1 << 16 // SMFIP_NR_DATA
	OptNoUnknownReply OptProtocol = 1 << 17 // SMFIP_NR_UNKN
	OptNoEOHReply     OptProtocol = 1 << 18 // SMFIP_NR_EOH
	OptNoBodyReply    OptProtocol = 1 << 19 // SMFIP_NR_BODY

	// [v6]
	OptHeaderLeadingSpace OptProtocol = 1 << 20 // SMFIP_HDR_LEADSPC
)

// Milter is an interface for milter callback handlers.
type Milter interface {
	// Connect is called to provide SMTP connection data for incoming message.
	// Suppress with OptNoConnect.
	Connect(host string, addr net.IP, payload map[string]string, m *Modifier) (Response, []Feature, error)

	// Helo is called to process any HELO/EHLO related filters. Suppress with
	// OptNoHelo.
	Helo(name string, payload map[string]string, m *Modifier) (Response, []Feature, error)

	// MailFrom is called to process filters on envelope FROM address. Suppress
	// with OptNoMailFrom.
	MailFrom(from string, payload map[string]string, m *Modifier) (Response, []Feature, error)

	// RcptTo is called to process filters on envelope TO address. Suppress with
	// OptNoRcptTo.
	RcptTo(rcptTo string, payload map[string]string, m *Modifier) (Response, []Feature, error)

	// Header is called once for each header in incoming message. Suppress with
	// OptNoHeaders.
	Header(name string, value string, payload map[string]string, m *Modifier) (Response, []Feature, error)

	// Headers are called when all message headers have been processed. Suppress
	// with OptNoEOH.
	Headers(h textproto.MIMEHeader, payload map[string]string, m *Modifier) (Response, []Feature, error)

	// BodyChunk is called to process next message body chunk data (up to 64KB
	// in size). Suppress with OptNoBody.
	BodyChunk(chunk []byte, payload map[string]string, m *Modifier) (Response, []Feature, error)

	// Body is called at the end of each message. All changes to message's
	// content & attributes must be done here.
	Body(payload map[string]string, m *Modifier, macro map[string]string) (Response, []Feature, error)

	// Abort is called is the current message has been aborted. All message data
	// should be reset to prior to the Helo callback. Connection data should be
	// preserved.
	Abort(m *Modifier) error
}
