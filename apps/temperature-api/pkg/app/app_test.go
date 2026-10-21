package app

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentCloseRunsCleanupOnce(t *testing.T) {
	a := New(0)
	var calls atomic.Int32
	want := errors.New("cleanup failed")
	if err := a.AddCleanup("resource", func(context.Context) error {
		calls.Add(1)
		return want
	}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.Close(); !errors.Is(err, want) {
				t.Errorf("Close() = %v", err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("cleanup called %d times", calls.Load())
	}
	if err := a.Run(); !errors.Is(err, ErrStarted) {
		t.Fatalf("Run after Close = %v", err)
	}
}

func TestRunRejectsConcurrentRunAndClose(t *testing.T) {
	a := New(0)
	entered := make(chan struct{})
	release := make(chan struct{})
	want := errors.New("controller stopped")
	if err := a.AddController("controller", func(context.Context) error {
		close(entered)
		<-release
		return want
	}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- a.Run() }()
	<-entered
	runErr := a.Run()
	closeErr := a.Close()
	close(release)
	result := <-done
	if !errors.Is(runErr, ErrStarted) {
		t.Errorf("second Run = %v", runErr)
	}
	if !errors.Is(closeErr, ErrRunning) {
		t.Errorf("Close while running = %v", closeErr)
	}
	if !errors.Is(result, want) {
		t.Errorf("Run = %v", result)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
