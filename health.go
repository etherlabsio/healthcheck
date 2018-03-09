package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

type response struct {
	Status string            `json:"status,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

type health struct {
	checkers  map[string]Checker
	observers map[string]Checker
	timeout   time.Duration
}

// Checker checks the status of the dependency and returns error.
// In case the dependency is working as expected, return nil.
type Checker interface {
	Check(ctx context.Context) error
}

// CheckerFunc is a convenience type to create functions that implement the Checker interface.
type CheckerFunc func(ctx context.Context) error

// Check Implements the Checker interface to allow for any func() error method
// to be passed as a Checker
func (c CheckerFunc) Check(ctx context.Context) error {
	return c(ctx)
}

// Handler returns an http.Handler
func Handler(opts ...Option) http.Handler {
	h := &health{
		checkers:  make(map[string]Checker),
		observers: make(map[string]Checker),
		timeout:   30 * time.Second,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// HandlerFunc returns an http.HandlerFunc to mount the API implementation at a specific route
func HandlerFunc(opts ...Option) http.HandlerFunc {
	return Handler(opts...).ServeHTTP
}

// Option adds optional parameter for the HealthcheckHandlerFunc
type Option func(*health)

// WithChecker adds a status checker that needs to be added as part of healthcheck. i.e database, cache or any external dependency
func WithChecker(name string, s Checker) Option {
	return func(h *health) {
		h.checkers[name] = &timeoutChecker{s}
	}
}

// WithObserver adds a status checker but it does not fail the entire status.
func WithObserver(name string, s Checker) Option {
	return func(h *health) {
		h.observers[name] = &timeoutChecker{s}
	}
}

// WithTimeout configures the global timeout for all individual checkers.
func WithTimeout(timeout time.Duration) Option {
	return func(h *health) {
		h.timeout = timeout
	}
}

func (h *health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nCheckers := len(h.checkers) + len(h.observers)

	code := http.StatusOK
	errorMsgs := make(map[string]string, nCheckers)

	ctx, cancel := context.Background(), func() {}
	if h.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, h.timeout)
	}
	defer cancel()

	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(nCheckers)

	for key, checker := range h.checkers {
		go func(key string, checker Checker) {
			if err := checker.Check(ctx); err != nil {
				mutex.Lock()
				errorMsgs[key] = err.Error()
				code = http.StatusServiceUnavailable
				mutex.Unlock()
			}
			wg.Done()
		}(key, checker)
	}
	for key, observer := range h.observers {
		go func(key string, observer Checker) {
			if err := observer.Check(ctx); err != nil {
				mutex.Lock()
				errorMsgs[key] = err.Error()
				mutex.Unlock()
			}
			wg.Done()
		}(key, observer)
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response{
		Status: http.StatusText(code),
		Errors: errorMsgs,
	})
}

type timeoutChecker struct {
	checker Checker
}

func (t *timeoutChecker) Check(ctx context.Context) error {
	checkerChan := make(chan error)
	go func() {
		checkerChan <- t.checker.Check(ctx)
	}()
	select {
	case err := <-checkerChan:
		return err
	case <-ctx.Done():
		return errors.New("max check time exceeded")
	}
}
