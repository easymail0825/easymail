package runtime

import (
	"easymail/internal/service"
	"sync"
)

type Manager struct {
	lock     sync.Mutex
	services []service.Manager
}

func NewManager() *Manager {
	return &Manager{services: make([]service.Manager, 0)}
}

func (m *Manager) Add(s service.Manager) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.services = append(m.services, s)
}

func (m *Manager) StartAll() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, s := range m.services {
		if err := s.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) StopAll() {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, s := range m.services {
		_ = s.Stop()
	}
}

