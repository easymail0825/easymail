package registry

import (
	"easymail/internal/service"
	"fmt"
	"sync"
)

type ServiceRegistry struct {
	lock     *sync.RWMutex
	services map[string]service.Manager
}

func New() *ServiceRegistry {
	return &ServiceRegistry{
		lock:     &sync.RWMutex{},
		services: make(map[string]service.Manager),
	}
}

func (r *ServiceRegistry) Register(m service.Manager) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.services[m.Name()] = m
}

func (r *ServiceRegistry) Unregister(name string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if s, ok := r.services[name]; !ok {
		return fmt.Errorf("service %s not found", name)
	} else {
		err := s.Stop()
		if err != nil {
			return err
		}
	}
	delete(r.services, name)
	return nil
}
