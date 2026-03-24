package app

import (
	"easymail/internal/adminapi"
	"easymail/internal/delivery"
	"easymail/internal/filtering"
	"easymail/internal/model"
	"easymail/internal/platform/bootstrap"
	runtime2 "easymail/internal/platform/runtime"
	"easymail/internal/policy"
	olddovecot "easymail/internal/service/dovecot"
	storage2 "easymail/internal/storage"
	"easymail/internal/webmailapi"
)

func Build(rt *bootstrap.Runtime) (*runtime2.Manager, error) {
	manager := runtime2.NewManager()
	for _, app := range rt.Config.Raw.Apps {
		if !app.Enable {
			continue
		}
		switch app.Name {
		case "dovecot":
			s := olddovecot.New(app.Family, app.Listen)
			if err := s.SetLogger(rt.Logger); err != nil {
				return nil, err
			}
			manager.Add(s)
		case "policy":
			s := policy.NewServer(app.Family, app.Listen)
			if err := s.SetLogger(rt.Logger); err != nil {
				return nil, err
			}
			manager.Add(s)
		case "filter":
			s := filtering.NewServer(app.Family, app.Listen, rt.Logger)
			if err := s.SetLogger(rt.Logger); err != nil {
				return nil, err
			}
			manager.Add(s)
		case "lmtp":
			s := delivery.NewServer(app.Family, app.Listen, 1024*1024*50)
			if err := s.SetLogger(rt.Logger); err != nil {
				return nil, err
			}
			c, err := model.GetConfigureByNames("easymail", "storage", "data")
			if err != nil {
				return nil, err
			}
			r, err := model.GetConfigureByNames("easymail", "configure", "root")
			if err != nil {
				return nil, err
			}
			delivery.AttachStorage(s, storage2.NewLocal(r.Value, c.Value))
			manager.Add(s)
		case "admin":
			s := adminapi.NewServer(app.Family, app.Listen, app.Parameter["root"], app.Parameter["cookie_password"], app.Parameter["cookie_tag"])
			if err := s.SetLogger(rt.Logger); err != nil {
				return nil, err
			}
			manager.Add(s)
		case "webmail":
			s := webmailapi.NewServer(app.Family, app.Listen, app.Parameter["root"], app.Parameter["cookie_password"], app.Parameter["cookie_tag"])
			if err := s.SetLogger(rt.Logger); err != nil {
				return nil, err
			}
			manager.Add(s)
		}
	}
	return manager, nil
}

