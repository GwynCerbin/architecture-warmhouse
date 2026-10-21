package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// Controller runs until cancellation or failure. Run must not return while its
// handlers or child goroutines still use shared infrastructure.
type Controller func(ctx context.Context) error

// Cleanup releases a resource and must respect the supplied deadline.
// Resources are registered immediately after creation, in creation order.
type Cleanup func(context.Context) error

var (
	ErrStarted        = errors.New("application is already running or closed")
	ErrRunning        = errors.New("cannot close resources while controllers are running")
	ErrUnexpectedExit = errors.New("controller exited before cancellation")
)

type controller struct {
	name string
	run  Controller
}

type resource struct {
	name  string
	close Cleanup
}

const (
	building uint32 = iota
	running
	stopped
	closing
	closed
)

// Application is single-use. Registrations are accepted only before Run/Close.
// Cancellation stops controllers; resources are closed only after all Run calls
// return. A controller that ignores cancellation can block Run indefinitely:
// closing its resources concurrently would be unsafe.
// AddController and AddCleanup must be called sequentially, before Run or Close.
// Atomic state transitions do not synchronize registration slices.
// An Application must not be copied after first use.
type Application struct {
	state          atomic.Uint32
	controllers    []controller
	resources      []resource
	cleanupTimeout time.Duration
	closed         chan struct{}
	closeErr       error
}

// New sets the total cooperative cleanup budget. Nonpositive values use 10s.
func New(cleanupTimeout time.Duration) *Application {
	if cleanupTimeout <= 0 {
		cleanupTimeout = 10 * time.Second
	}

	return &Application{cleanupTimeout: cleanupTimeout, closed: make(chan struct{})}
}

// Close releases resources once. It rejects calls while controllers are running.
// Concurrent callers wait for cleanup and receive the same result.
func (a *Application) Close() error {
	for {
		current := a.state.Load()
		switch current {
		case running:
			return ErrRunning
		case closing, closed:
			<-a.closed
			return a.closeErr
		case building, stopped:
			if !a.state.CompareAndSwap(current, closing) {
				continue
			}
		}
		break
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.cleanupTimeout)
	defer cancel()

	var failures []error

	for i := len(a.resources) - 1; i >= 0; i-- {
		r := a.resources[i]
		// Attempt every cleanup, even when another failed or exhausted the budget.
		if err := r.close(ctx); err != nil {
			failures = append(failures, fmt.Errorf("close %q: %w", r.name, err))
		}
	}

	a.closeErr = errors.Join(failures...)
	a.state.Store(closed)
	close(a.closed)

	return a.closeErr
}

func (a *Application) AddController(name string, c Controller) error {
	if a.state.Load() != building {
		return ErrStarted
	}

	if name == "" || c == nil {
		return errors.New("controller name and implementation are required")
	}

	for _, existing := range a.controllers {
		if existing.name == name {
			return fmt.Errorf("duplicate controller %q", name)
		}
	}

	a.controllers = append(a.controllers, controller{name: name, run: c})

	return nil
}

func (a *Application) AddCleanup(name string, cleanup Cleanup) error {
	if a.state.Load() != building {
		return ErrStarted
	}

	if name == "" || cleanup == nil {
		return errors.New("resource name and cleanup are required")
	}

	for _, existing := range a.resources {
		if existing.name == name {
			return fmt.Errorf("duplicate resource %q", name)
		}
	}

	a.resources = append(a.resources, resource{name: name, close: cleanup})

	return nil
}

// Run starts controllers and waits for them to finish. OS signals cancel their
// context. Call Close after Run returns to release registered resources.
func (a *Application) Run() error {
	if !a.state.CompareAndSwap(building, running) {
		return ErrStarted
	}

	var (
		notifyCtx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		group, runCtx     = errgroup.WithContext(notifyCtx)
	)

	defer cancel()

	for _, c := range a.controllers {
		group.Go(func() error {
			if runCtx.Err() != nil {
				return nil
			}

			err := c.run(runCtx)
			if err != nil {
				if runCtx.Err() != nil && errors.Is(err, runCtx.Err()) {
					return nil
				}

				return fmt.Errorf("controller %q: %w", c.name, err)
			}

			if runCtx.Err() == nil {
				return fmt.Errorf("controller %q: %w", c.name, ErrUnexpectedExit)
			}

			return nil
		})
	}

	runErr := group.Wait()

	a.state.CompareAndSwap(running, stopped)

	return runErr
}
