package delivery

import (
	oldlmtp "easymail/internal/service/lmtp"
	"easymail/internal/service/storage"
)

func NewServer(family, listen string, limit int64) *oldlmtp.Server {
	return oldlmtp.New(family, listen, limit, "8BITMIME", "ENHANCEDSTATUSCODES", "PIPELINING")
}

func AttachStorage(s *oldlmtp.Server, st storage.Storager) {
	s.SetStorage(st)
}

