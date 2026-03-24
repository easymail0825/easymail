package adminapi

import oldadmin "easymail/internal/service/admin"

func NewServer(family, listen, root, cookiePassword, cookieTag string) *oldadmin.Server {
	return oldadmin.New(family, listen, root, cookiePassword, cookieTag)
}

