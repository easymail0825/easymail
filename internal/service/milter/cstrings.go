package milter

import (
	"bytes"
	"strings"
)

// NULL terminator
const Null = "\x00"

// DecodeCStrings splits a C style strings into a Go slice
func DecodeCStrings(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	return strings.Split(strings.Trim(string(data), Null), Null)
}

// ReadCString reads and returns a C style string from []byte
func ReadCString(data []byte) string {
	pos := bytes.IndexByte(data, 0)
	if pos == -1 {
		return string(data)
	}
	return string(data[0:pos])
}
