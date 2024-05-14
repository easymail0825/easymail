package lmtp

import (
	"bytes"
	"easymail/internal/model"
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
	"time"
)

func ParseMail(raw []byte) (email *model.Email, err error) {
	var e *enmime.Envelope
	p := enmime.NewParser()
	e, err = p.ReadEnvelope(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	email = &model.Email{
		Subject:  e.GetHeader("Subject"),
		JobID:    uuid.New().String(),
		QueueID:  e.GetHeader("X-Queue-ID"),
		Size:     int64(len(raw)),
		MailTime: time.Now(),
	}
	return email, nil
}
