package manager_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/flowshot-io/x/pkg/logger"
	"github.com/flowshot-io/x/pkg/manager"
)

type SimpleService struct {
	name    manager.ServiceName
	started bool
	stopped bool
}

func (s *SimpleService) Start() error {
	s.started = true
	return nil
}

func (s *SimpleService) Stop() error {
	s.stopped = true
	return nil
}

func TestServiceManager(t *testing.T) {
	t.Run("Add and Start Services", func(t *testing.T) {
		serviceManager := manager.New(&manager.Options{Logger: logger.NoOp()})

		s1 := &SimpleService{name: "Service1"}
		s2 := &SimpleService{name: "Service2"}

		if err := serviceManager.Add("Service1", s1); err != nil {
			t.Fatalf("Failed to add service: %v", err)
		}

		if err := serviceManager.Add("Service2", s2); err != nil {
			t.Fatalf("Failed to add service: %v", err)
		}

		if err := serviceManager.Start(); err != nil {
			t.Fatalf("Failed to start services: %v", err)
		}

		if !s1.started {
			t.Errorf("Service1 was not started")
		}

		if !s2.started {
			t.Errorf("Service2 was not started")
		}
	})

	t.Run("Stop Services", func(t *testing.T) {
		serviceManager := manager.New(&manager.Options{Logger: logger.NoOp()})

		s1 := &SimpleService{name: "Service1"}
		s2 := &SimpleService{name: "Service2"}

		if err := serviceManager.Add("Service1", s1); err != nil {
			t.Fatalf("Failed to add service: %v", err)
		}

		if err := serviceManager.Add("Service2", s2); err != nil {
			t.Fatalf("Failed to add service: %v", err)
		}

		if err := serviceManager.Start(); err != nil {
			t.Fatalf("Failed to start services: %v", err)
		}

		if err := serviceManager.Stop(); err != nil {
			t.Fatalf("Failed to stop services: %v", err)
		}

		if !s1.stopped {
			t.Errorf("Service1 was not stopped")
		}

		if !s2.stopped {
			t.Errorf("Service2 was not stopped")
		}
	})
}

func TestRaceCondition(t *testing.T) {
	serviceManager := manager.New(&manager.Options{Logger: logger.NoOp()})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			service := &SimpleService{name: manager.ServiceName(fmt.Sprintf("Service%d", i))}
			err := serviceManager.Add(manager.ServiceName(fmt.Sprintf("Service%d", i)), service)
			if err != nil {
				t.Errorf("Failed to add service: %v", err)
			}
		}(i)
	}

	wg.Wait()

	err := serviceManager.Start()
	if err != nil {
		t.Fatalf("Failed to start services: %v", err)
	}

	err = serviceManager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop services: %v", err)
	}
}
