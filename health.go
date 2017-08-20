package health

import (
	"encoding/json"
	"net/http"
)

type (
	// Option adds optional parameter for the HealthcheckHandlerFunc
	Option func(*health)

	// Checker checks the status of the dependency and returns error.
	// In case the dependency is working as expected, return nil.
	Checker interface {
		Check() error
	}

	// CheckFunc is a convenience type to create functions that implement
	// the Checker interface
	CheckerFunc func() error

	// Check Implements the Checker interface to allow for any func() error method
	// to be passed as a Checker
	func (c CheckerFunc) Check() error {
		return c()
	}

	health struct {
		checkers map[string]Checker
	}
)

func (h *health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := http.StatusOK
	var errorMsgs map[string]string
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	for key, checker := range h.checkers {
		if err := checker.Check(); err != nil {
			errorMsgs[key] = err.Error()
			code = http.StatusServiceUnavailable
		}
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(struct {
		Status string            `json:"status,omitempty"`
		Errors map[string]string `json:"errors,omitempty"`
	}{
		Status: http.StatusText(code),
		Errors: errorMsgs,
	})
}

func new(opts ...Option) *health {
	h := &health{
		checkers: make(map[string]Checker),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// NewHandlerFunc returns a http.Handler
func NewHandlerFunc(opts ...Option) http.Handler {
	return new(opts...)
}

// NewHandler returns a http.Handler
func NewHandler(opts ...Option) http.HandlerFunc {
	return new(opts...).ServeHTTP
}

// WithChecker adds a status checker that needs to be added as part of healthcheck. i.e database, cache or any external dependency
func WithChecker(name string, s Checker) Option {
	return func(h *health) {
		h.checkers[name] = s
	}
}
