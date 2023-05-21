package manager

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

type (
	// ServiceName defines the type for a service name.
	ServiceName string

	// ServiceController defines the interface for a service manager.
	ServiceController interface {
		Add(name ServiceName, s Service) error
		Start() error
		Stop() error
	}

	// Service defines the interface for a service with Start and Stop methods.
	Service interface {
		Start() error
		Stop() error
	}
)

// ServiceManager is responsible for managing multiple services.
type ServiceManager struct {
	mu        sync.RWMutex
	services  map[ServiceName]Service
	opMutex   sync.Mutex
	operating bool
}

// New creates and returns a new ServiceManager.
func New() *ServiceManager {
	return &ServiceManager{
		services: make(map[ServiceName]Service),
	}
}

// Add adds a service to the ServiceManager.
// It returns an error if a Start or Stop operation is currently in progress.
func (sm *ServiceManager) Add(name ServiceName, s Service) error {
	sm.opMutex.Lock()
	if sm.operating {
		sm.opMutex.Unlock()
		return errors.New("cannot add service during start or stop operation")
	}
	sm.opMutex.Unlock()

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.services[name] = s

	return nil
}

// Start starts all services concurrently.
// If any service fails to start, it stops the already started services and returns the error.
func (sm *ServiceManager) Start() error {
	sm.opMutex.Lock()
	sm.operating = true
	sm.opMutex.Unlock()

	defer func() {
		sm.opMutex.Lock()
		sm.operating = false
		sm.opMutex.Unlock()
	}()

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var startedServices []Service

	var g errgroup.Group
	for name, s := range sm.services {
		// Copy variables to prevent data race
		s := s
		name := name

		g.Go(func() error {
			err := s.Start()
			if err != nil {
				return fmt.Errorf("error starting service %s: %w", name, err)
			}

			startedServices = append(startedServices, s)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		// Stop the already started services
		for _, s := range startedServices {
			s.Stop()
		}

		return err
	}

	return nil
}

// Stop stops all services concurrently.
// If any service fails to stop, it continues to stop other services and returns the error.
func (sm *ServiceManager) Stop() error {
	sm.opMutex.Lock()
	sm.operating = true
	sm.opMutex.Unlock()

	defer func() {
		sm.opMutex.Lock()
		sm.operating = false
		sm.opMutex.Unlock()
	}()

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var g errgroup.Group
	for name, s := range sm.services {
		// Copy variables to prevent data race
		s := s
		name := name

		g.Go(func() error {
			if err := s.Stop(); err != nil {
				return fmt.Errorf("error stopping service %s: %w", name, err)
			}

			return nil
		})
	}

	return g.Wait()
}
