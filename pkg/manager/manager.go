package manager

import (
	"fmt"
	"sync"
)

type ServiceName string

// Service interface with Start() and Stop() methods
type Service interface {
	Start() error
	Stop() error
}

// ServiceManager manages multiple services
type ServiceManager struct {
	services map[ServiceName]Service
}

// New creates a new ServiceManager
func New() *ServiceManager {
	return &ServiceManager{
		services: make(map[ServiceName]Service),
	}
}

func (sm *ServiceManager) Add(name ServiceName, s Service) {
	sm.services[name] = s
}

// Start starts all services concurrently
func (sm *ServiceManager) Start() {
	var wg sync.WaitGroup
	for name, s := range sm.services {
		wg.Add(1)
		go func(name string, s Service) {
			defer wg.Done()
			if err := s.Start(); err != nil {
				fmt.Printf("Error starting service %s: %v\n", name, err)
			}
		}(string(name), s)
	}
	wg.Wait()
}

// Stop stops all services concurrently
func (sm *ServiceManager) Stop() {
	var wg sync.WaitGroup
	for name, s := range sm.services {
		wg.Add(1)
		go func(name string, s Service) {
			defer wg.Done()
			if err := s.Stop(); err != nil {
				fmt.Printf("Error stopping service %s: %v\n", name, err)
			}
		}(string(name), s)
	}
	wg.Wait()
}
