package lmtp

import (
	"bytes"
	"easymail/internal/account"
	"github.com/jhillyerd/enmime"
	"time"
)

func ParseMail(raw []byte) (email *account.Email, err error) {
	var e *enmime.Envelope
	p := enmime.NewParser()
	e, err = p.ReadEnvelope(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	email = &account.Email{
		Subject:  e.GetHeader("Subject"),
		JobID:    e.GetHeader("X-Queue-ID"),
		Size:     int64(len(raw)),
		MailTime: time.Now(),
	}
	return email, nil
}
