package main

import (
	"easymail/cmd/easymail/registry"
	"easymail/internal/database"
	"easymail/internal/easylog"
	"easymail/internal/maillog"
	"easymail/internal/model"
	"easymail/internal/postfix/command"
	"easymail/internal/postfix/policy"
	"easymail/internal/service/admin"
	"easymail/internal/service/agent"
	"easymail/internal/service/dovecot"
	"easymail/internal/service/filter"
	"easymail/internal/service/lmtp"
	"easymail/internal/service/storage"
	"easymail/internal/service/webmail"
	"errors"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

var appConfig *database.AppConfig
var err error

func init() {
	appConfig, err = database.ReadAppConfig("easymail.yaml")
	if err != nil {
		log.Println("Error reading app config:", err)
		return
	}
}

func initialize() (err error) {
	// check .initialized file, if exists, do nothing
	if _, err := os.Stat(filepath.Join("./", ".initialized")); err == nil {
		log.Println(".initialized file exists, do nothing")
		return nil
	}
	db := database.GetDB()
	err = db.AutoMigrate(&model.Account{}, &model.Domain{}, &model.Admin{}, &model.Email{}, &model.Configure{}, &maillog.MailLog{})
	if err != nil {
		return err
	}

	// check default domain(localhost)
	localhost := model.Domain{
		Name:        "localhost",
		Description: "create from init",
		Active:      true,
		Deleted:     false,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	err = db.Model(&localhost).Where("name=?", "localhost").First(&localhost).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&localhost).Error
	}

	// check default model(root@localhost)
	var root model.Account
	err = db.Model(&root).Where("username = ?", "root").Where("domain_id=?", localhost.ID).First(&root).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		password, err := model.GeneratePassword("easymail")
		if err != nil {
			return err
		}
		now := time.Now()
		// create it
		root = model.Account{
			Username:           "root",
			Password:           password,
			DomainID:           localhost.ID,
			Active:             true,
			UpdateTime:         now,
			CreateTime:         now,
			PasswordExpireTime: now.Add(70 * 365 * 24 * time.Hour),
		}
		err = db.Create(&root).Error
		if err != nil {
			return err
		}
	}

	// check default super admin
	administrator := model.Admin{}
	err = db.Model(&administrator).Where("domain_id=?", localhost.ID).Where("account_id=?", root.ID).First(&administrator).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = db.Create(&model.Admin{
			AccountID:  root.ID,
			DomainID:   localhost.ID,
			IsSuper:    true,
			CreateTime: time.Now(),
		}).Error
		if err != nil {
			return err
		}
	}
	log.Printf("Account initialized, please login\nusername:root@localhost\npassword:easymail\n")

	// create default config
	type ConfigureItem struct {
		names    []string
		value    string
		dataType model.DataType
	}

	// get current workdir
	workdir, _ := os.Getwd()
	log.Println("workdir:", workdir)
	defaultConfigureItems := []ConfigureItem{
		{[]string{"easymail", "storage", "data"}, "./data", model.DataTypeString},
		{[]string{"easymail", "configure", "root"}, workdir, model.DataTypeString},
		{[]string{"network", "dns", "nameserver"}, "8.8.8.8", model.DataTypeString},
		{[]string{"postfix", "log", "mail"}, "/var/log/mail.log", model.DataTypeString},
		{[]string{"postfix", "configure", "virtual_mailbox_domains"}, path.Join(workdir, "virtual_domains"), model.DataTypeString},
		{[]string{"postfix", "execute", "postmap"}, "/usr/sbin/postmap", model.DataTypeString},
		{[]string{"postfix", "execute", "postcat"}, "/usr/sbin/postcat", model.DataTypeString},
		{[]string{"postfix", "execute", "postqueue"}, "/usr/sbin/postqueue", model.DataTypeString},
		{[]string{"postfix", "execute", "postsuper"}, "/usr/sbin/postsuper", model.DataTypeString},
		{[]string{"postfix", "execute", "postconf"}, "/usr/sbin/postconf", model.DataTypeString},
		{[]string{"postfix", "execute", "postfix"}, "/usr/sbin/postfix", model.DataTypeString},
	}
	for _, cfg := range defaultConfigureItems {
		if _, err := model.GetConfigure(cfg.names...); errors.Is(err, gorm.ErrRecordNotFound) {
			_, err = model.CreateConfigure(cfg.value, cfg.dataType, cfg.names...)
			if err != nil {
				return err
			}
		}
	}

	// initialize postfix configure
	// init dovecot configure
	log.Println("initialize postfix configure")
	for _, app := range appConfig.Apps {
		if app.Name == "dovecot" && app.Enable {
			url, err := command.MakeServicePath(app.Family, app.Listen)
			if err != nil {
				return err
			}
			cl := []command.PostfixConfigure{
				{Name: "smtpd_sasl_auth_enable", Value: "yes"},
				{Name: "smtpd_sasl_type", Value: "dovecot"},
				{Name: "smtpd_sasl_path", Value: url},
				{Name: "smtpd_relay_restrictions", Value: "permit_mynetworks,permit_sasl_authenticated,defer_unauth_destination"},
			}
			if err = command.FlushPostfixConfig(cl); err != nil {
				return err
			}
		} else if app.Name == "policy" && app.Enable {
			url, err := command.MakeServicePath(app.Family, app.Listen)
			if err != nil {
				return err
			}
			cl := []command.PostfixConfigure{
				{Name: "smtpd_recipient_restrictions", Value: "check_policy_service " + url},
			}
			if err = command.FlushPostfixConfig(cl); err != nil {
				return err
			}
		} else if app.Name == "filter" && app.Enable {
			url, err := command.MakeServicePath(app.Family, app.Listen)
			if err != nil {
				return err
			}
			cl := []command.PostfixConfigure{
				{Name: "smtpd_milters", Value: url},
			}
			if err = command.FlushPostfixConfig(cl); err != nil {
				return err
			}
		} else if app.Name == "lmtp" && app.Enable {
			c, err := model.GetConfigure("postfix", "configure", "virtual_mailbox_domains")
			if err != nil {
				return err
			}
			if c.DataType != model.DataTypeString {
				return errors.New("wrong data type")
			}

			url, err := command.MakeServicePath(app.Family, app.Listen)
			if err != nil {
				return err
			}
			cl := []command.PostfixConfigure{
				{Name: "virtual_mailbox_domains", Value: "hash:" + c.Value},
				{Name: "virtual_transport", Value: "lmtp:" + url},
				{Name: "virtual_mailbox_base", Value: "/dev/null"},
				{Name: "default_transport", Value: "smtp"},
				{Name: "relay_transport", Value: "relay"},
			}
			if err = command.FlushPostfixConfig(cl); err != nil {
				return err
			}
		}

	}

	// create .initialized file, mark the initialization completed
	err = os.WriteFile(".initialized", []byte(""), 0644)
	if err != nil {
		log.Println("Error writing .initialized file:", err)
		return
	}

	return nil
}

func main() {
	// initialize model
	err = initialize()
	if err != nil {
		log.Println("Error initializing model:", err)
	}

	// initialize logger
	serviceRegistry := registry.New()
	if appConfig.LogFile == "" {
		appConfig.LogFile = "easymail.log"
	}
	logFile, err := os.OpenFile(appConfig.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	_log := easylog.NewLogger(logFile, "")
	err = easylog.SetOutputByName(filepath.Base(appConfig.LogFile))
	if err != nil {
		panic(err)
	}
	_log.SetHighlighting(false)
	_log.SetRotateByDay()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// start services
	for _, app := range appConfig.Apps {
		if !app.Enable {
			continue
		}
		switch app.Name {
		case "dovecot":
			s := dovecot.New(app.Family, app.Listen)
			if err = s.SetLogger(_log); err != nil {
				panic(err)
			}
			serviceRegistry.Register(s)
			go func() {
				if err := s.Start(); err != nil {
					panic(err)
				}
			}()
		case "policy":
			s := policy.New(app.Family, app.Listen)
			if err = s.SetLogger(_log); err != nil {
				panic(err)
			}
			serviceRegistry.Register(s)
			go func() {
				if err := s.Start(); err != nil {
					panic(err)
				}
			}()
		case "filter":
			s := filter.New(app.Family, app.Listen)
			if err = s.SetLogger(_log); err != nil {
				panic(err)
			}
			serviceRegistry.Register(s)
			go func() {
				if err := s.Start(); err != nil {
					panic(err)
				}
			}()
		case "lmtp":
			s := lmtp.New(app.Family, app.Listen, 1024*1024*50, []string{"8BITMIME", "ENHANCEDSTATUSCODES", "PIPELINING"}...)
			if err = s.SetLogger(_log); err != nil {
				panic(err)
			}
			serviceRegistry.Register(s)
			go func() {
				c, err := model.GetConfigure("easymail", "storage", "data")
				if err != nil {
					log.Fatal("mail storage data is not defined")
				}
				r, err := model.GetConfigure("easymail", "configure", "root")
				if err != nil {
					log.Fatal("easymail configure root is not defined")
				}

				// add storage
				localStorage := storage.NewLocalStorage(r.Value, c.Value)
				s.SetStorage(localStorage)

				if err := s.Start(); err != nil {
					panic(err)
				}
			}()
		case "admin":
			root, ok := app.Parameter["root"]
			if !ok {
				panic("admin template is not defined")
			}
			cookiePassword, ok := app.Parameter["cookie_password"]
			if !ok {
				panic("cookiePassword is not defined")
			}
			cookieTag, ok := app.Parameter["cookie_tag"]
			if !ok {
				panic("cookieTag is not defined")
			}
			s := admin.New(app.Family, app.Listen, root, cookiePassword, cookieTag)
			if err = s.SetLogger(_log); err != nil {
				panic(err)
			}
			serviceRegistry.Register(s)
			go func() {
				if err := s.Start(); err != nil {
					panic(err)
				}
			}()
		case "webmail":
			root, ok := app.Parameter["root"]
			if !ok {
				panic("admin template is not defined")
			}
			cookiePassword, ok := app.Parameter["cookie_password"]
			if !ok {
				panic("cookiePassword is not defined")
			}
			cookieTag, ok := app.Parameter["cookie_tag"]
			if !ok {
				panic("cookieTag is not defined")
			}
			s := webmail.New(app.Family, app.Listen, root, cookiePassword, cookieTag)
			if err = s.SetLogger(_log); err != nil {
				panic(err)
			}
			serviceRegistry.Register(s)
			go func() {
				if err := s.Start(); err != nil {
					panic(err)
				}
			}()
		case "agent":
			s := agent.New()
			if err = s.SetLogger(_log); err != nil {
				panic(err)
			}
			serviceRegistry.Register(s)
			go func() {
				if err := s.Start(); err != nil {
					panic(err)
				}
			}()
		}
	}

	// wait for signal
	sig := <-sigChan
	log.Printf("Received signal: %v, shutting down...\n", sig)

	// stop services
	for _, app := range appConfig.Apps {
		if app.Enable {
			if err = serviceRegistry.Unregister(app.Name); err != nil {
				log.Println("Error stopping service:", err)
			}
		}
	}
	os.Exit(0)

}
