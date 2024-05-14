package storage

import (
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
	"os"
	"testing"
)

func TestReadMail(t *testing.T) {
	fd, err := os.Open("/home/bobxiao/tmp/test.eml")
	msg, err := enmime.ReadEnvelope(fd)
	if err != nil {
		t.Fatal(err)
	}
	sender := msg.GetHeaderValues("To")
	t.Log(sender)
}

func TestUuid(t *testing.T) {
	u := uuid.New()
	t.Log(u.String())
	t.Log(u.Time())
	t.Log(u.ID())

	u2, err := uuid.Parse("21efddc6-7422-4064-9be3-1e98ea6b7e5d")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u2)
}
