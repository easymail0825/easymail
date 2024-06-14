package agent

import (
	"easymail/internal/easylog"
	"easymail/internal/maillog"
	"easymail/internal/model"
	"fmt"
	"github.com/hpcloud/tail"
	"sync"
)

type Process struct {
	name    string
	stopCh  chan struct{}
	started bool
	lock    *sync.Mutex
	debug   bool
	_log    *easylog.Logger
}

func New() *Process {
	return &Process{
		name:   "agent",
		stopCh: make(chan struct{}),
		lock:   &sync.Mutex{},
		debug:  false,
	}
}

func (p *Process) SetDebug(debug bool) {
	p.debug = debug
}

func (p *Process) SetLogger(_log *easylog.Logger) error {
	if _log == nil {
		return fmt.Errorf("%s logger is nil", p.name)
	}
	p._log = _log
	return nil
}

func (p *Process) Start() error {
	if p._log == nil {
		return fmt.Errorf("%s logger is nil", p.name)
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.started {
		return fmt.Errorf("%s server already started", p.name)
	}

	p.started = true
	p._log.Infof("%s server started!", p.name)
	go p.run()
	return nil
}

func (p *Process) Stop() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if !p.started {
		return fmt.Errorf("%s server not started", p.name)
	}

	p.started = false
	close(p.stopCh)
	p._log.Infof("%s server stopped!", p.name)
	return nil
}

func (p *Process) Name() string {
	return p.name
}

func (p *Process) run() {
	cfg, err := model.GetConfigureByNames("postfix", "log", "mail")
	if err != nil {
		p._log.Fatal(err)
	}
	t, err := tail.TailFile(cfg.Value, tail.Config{
		Follow: true,
		ReOpen: true,
	})
	for line := range t.Lines {
		mailLog, err := maillog.Parse(line.Text)
		if err != nil {
			p._log.Error(err)
			continue
		}
		if err := maillog.Save(mailLog); err != nil {
			p._log.Error(err)
		}
	}
}
