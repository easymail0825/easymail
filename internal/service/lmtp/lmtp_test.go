package lmtp

import (
	"github.com/jhillyerd/enmime"
	"strings"
	"testing"
)

// EqualBytes 判断两个[]byte切片是否相等
func EqualBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParseMailbox(t *testing.T) {
	tests := []struct {
		input  []byte
		output []byte
	}{
		//{
		//	input:  []byte("<user@example.com>"),
		//	output: []byte("user@example.com"),
		//},
		//{
		//	input:  []byte("<>"),
		//	output: []byte(""),
		//},
		{
			input:  []byte("<root@localhost> size=386"),
			output: []byte("root@localhost"),
		},
	}
	for _, tt := range tests {
		if output, _ := parseMailbox(tt.input); EqualBytes(output, tt.output) != true {
			t.Errorf("parseMailbox(%q) = %q, want %q", tt.input, output, tt.output)
		}
	}
}

func TestMailboxCheck(t *testing.T) {
	f := emailRegex.MatchString("a@a.com")
	t.Log(f)
}

func TestParseMail(t *testing.T) {
	raw := `Received: from [127.0.0.1] (WorkStation.lan [192.168.1.106])
        by ubuntu.lan (Postfix) with ESMTP id BFD3F407C5
        for <admin@super.com>; Thu, 21 Mar 2024 05:13:09 +0000 (UTC)
Content-Type: multipart/mixed; boundary="===============2732670983689014392=="
MIME-Version: 1.0
From: root@qq.com
To: admin@super.com
Subject: Test Email from outside
X-Queue-ID: BFD3F407C5

--===============2732670983689014392==
Content-Type: text/plain; charset="us-ascii"
MIME-Version: 1.0
Content-Transfer-Encoding: 7bit

This is a test email from outside
--===============2732670983689014392==--
`
	p := enmime.NewParser()
	t.Log(p)
	e, err := p.ReadEnvelope(strings.NewReader(raw))
	if err != nil {
		t.Fail()
	}
	t.Log(e)
}

func TestJobID(t *testing.T) {
	s := "C0A9D400DA"
	t.Log(string(s[0]))
}
