package webmailapi

import oldwebmail "easymail/internal/service/webmail"

func NewServer(family, listen, root, cookiePassword, cookieTag string) *oldwebmail.Server {
	return oldwebmail.New(family, listen, root, cookiePassword, cookieTag)
}

