package manager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/flowshot-io/x/pkg/logger"
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

	Options struct {
		Logger logger.Logger
	}

	// ServiceManager is responsible for managing multiple services.
	ServiceManager struct {
		mu        sync.RWMutex
		services  map[ServiceName]Service
		opMutex   sync.Mutex
		operating bool
		logger    logger.Logger
	}
)

// New creates and returns a new ServiceManager.
func New(opts *Options) *ServiceManager {
	if opts == nil {
		opts = &Options{}
	}

	if opts.Logger == nil {
		opts.Logger = logger.New(nil)
	}

	return &ServiceManager{
		services: make(map[ServiceName]Service),
		logger:   opts.Logger,
	}
}

// Add adds a service to the ServiceManager.
// It returns an error if a Start or Stop operation is currently in progress.
func (sm *ServiceManager) Add(name ServiceName, s Service) error {
	sm.opMutex.Lock()
	if sm.operating {
		sm.opMutex.Unlock()
		sm.logger.Warn("Cannot add service during start or stop operation", map[string]interface{}{
			"serviceName": name,
		})
		return errors.New("cannot add service during start or stop operation")
	}
	sm.opMutex.Unlock()

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.services[name] = s
	sm.logger.Info("Service added to service manager", map[string]interface{}{
		"serviceName": name,
	})

	return nil
}

// Start starts all services concurrently.
// If any service fails to start, it stops the already started services and returns the error.
func (sm *ServiceManager) Start() error {
	sm.logger.Info("Starting services...")
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
	var startedServicesMutex sync.Mutex

	var g errgroup.Group
	for name, s := range sm.services {
		// Copy variables to prevent data race
		s := s
		name := name

		g.Go(func() error {
			err := s.Start()
			if err != nil {
				sm.logger.Error(fmt.Sprintf("Error starting service %s", name), map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("error starting service %s: %w", name, err)
			}

			startedServicesMutex.Lock()
			startedServices = append(startedServices, s)
			startedServicesMutex.Unlock()

			sm.logger.Info(fmt.Sprintf("Service %s started successfully", name))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		// Stop the already started services
		for _, s := range startedServices {
			s.Stop()
		}

		sm.logger.Error("Error during starting services", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	sm.logger.Info("All services started successfully")
	return nil
}

// Stop stops all services concurrently.
// If any service fails to stop, it continues to stop other services and returns the error.
func (sm *ServiceManager) Stop() error {
	sm.logger.Info("Stopping services...")
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
				sm.logger.Error(fmt.Sprintf("Error stopping service %s", name), map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("error stopping service %s: %w", name, err)
			}

			sm.logger.Info(fmt.Sprintf("Service %s stopped successfully", name))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		sm.logger.Error("Error during stopping services", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	sm.logger.Info("All services stopped successfully")
	return nil
}
