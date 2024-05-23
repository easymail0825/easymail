// Modifier instance is provided to milter handlers to modify email messages

package milter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/textproto"
)

// postfix wants LF lines endings. Using CRLF results in double CR sequences.
func crlfToLF(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte{'\r', '\n'}, []byte{'\n'})
}

// Modifier provides access to Macros, Headers and Body data to callback handlers. It also defines a
// number of functions that can be used by callback handlers to modify processing of the email message
type Modifier struct {
	Macros  map[string]string
	Headers textproto.MIMEHeader

	WritePacket func(*Message) error
}

// AddRecipient appends a new envelope recipient for current message
func (m *Modifier) AddRecipient(r string) error {
	data := []byte(fmt.Sprintf("<%s>", r) + Null)
	return m.WritePacket(NewResponse('+', data).Response())
}

// DeleteRecipient removes an envelope recipient address from message
func (m *Modifier) DeleteRecipient(r string) error {
	data := []byte(fmt.Sprintf("<%s>", r) + Null)
	return m.WritePacket(NewResponse('-', data).Response())
}

// ReplaceBody substitutes message body with provided body
func (m *Modifier) ReplaceBody(body []byte) error {
	body = crlfToLF(body)
	return m.WritePacket(NewResponse('b', body).Response())
}

// AddHeader appends a new email message header the message
func (m *Modifier) AddHeader(name, value string) error {
	var buffer bytes.Buffer
	buffer.WriteString(name + Null)
	buffer.Write(crlfToLF([]byte(value)))
	buffer.WriteString(Null)
	return m.WritePacket(NewResponse('h', buffer.Bytes()).Response())
}

// Quarantine a message by giving a reason to hold it
func (m *Modifier) Quarantine(reason string) error {
	return m.WritePacket(NewResponse('q', []byte(reason+Null)).Response())
}

// ChangeHeader replaces the header at the specified position with a new one.
// The index is per name.
func (m *Modifier) ChangeHeader(index int, name, value string) error {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, uint32(index)); err != nil {
		return err
	}
	buffer.WriteString(name + Null)
	buffer.Write(crlfToLF([]byte(value)))
	buffer.WriteString(Null)
	return m.WritePacket(NewResponse('m', buffer.Bytes()).Response())
}

// InsertHeader inserts the header at the specified position
func (m *Modifier) InsertHeader(index int, name, value string) error {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, uint32(index)); err != nil {
		return err
	}
	buffer.WriteString(name + Null)
	buffer.Write(crlfToLF([]byte(value)))
	buffer.WriteString(Null)
	return m.WritePacket(NewResponse('i', buffer.Bytes()).Response())
}

// ChangeFrom replaces the FROM envelope header with a new one
func (m *Modifier) ChangeFrom(value string) error {
	data := []byte(value + Null)
	return m.WritePacket(NewResponse('e', data).Response())
}
