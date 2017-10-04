package healthcheck

import (
	"encoding/json"
	"net/http"
)

type (
	health struct {
		checkers  map[string]Checker
		observers map[string]Checker
	}

	response struct {
		Status string            `json:"status,omitempty"`
		Errors map[string]string `json:"errors,omitempty"`
	}

	// Option adds optional parameter for the HealthcheckHandlerFunc
	Option func(*health)

	// Checker checks the status of the dependency and returns error.
	// In case the dependency is working as expected, return nil.
	Checker interface {
		Check() error
	}

	// CheckerFunc is a convenience type to create functions that implement the Checker interface. Shoutout to https://github.com/docker/go-healthcheck for this tip :)
	CheckerFunc func() error
)

// Handler returns an http.Handler
func Handler(opts ...Option) http.Handler {
	h := &health{
		checkers:  make(map[string]Checker),
		observers: make(map[string]Checker),
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

// Check Implements the Checker interface to allow for any func() error method
// to be passed as a Checker
func (c CheckerFunc) Check() error {
	return c()
}

// WithChecker adds a status checker that needs to be added as part of healthcheck. i.e database, cache or any external dependency
func WithChecker(name string, s Checker) Option {
	return func(h *health) {
		h.checkers[name] = s
	}
}

// WithObserver adds a status checker but it does not fail the whole status
func WithObserver(name string, s Checker) Option {
	return func(h *health) {
		h.observers[name] = s
	}
}

func (h *health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := http.StatusOK
	errorMsgs := make(map[string]string, len(h.checkers))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	for key, checker := range h.checkers {
		if err := checker.Check(); err != nil {
			errorMsgs[key] = err.Error()
			code = http.StatusServiceUnavailable
		}
	}
	for key, observer := range h.observers {
		if err := observer.Check(); err != nil {
			errorMsgs[key] = err.Error()
		}
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response{
		Status: http.StatusText(code),
		Errors: errorMsgs,
	})
}
