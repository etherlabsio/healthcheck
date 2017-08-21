package healthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewHandlerFunc(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name       string
		args       []Option
		statusCode int
		response   response
	}{
		{
			name:       "returns 200 status if no errors",
			statusCode: http.StatusOK,
			response: response{
				Status: http.StatusText(http.StatusOK),
			},
		},
		{
			name:       "returns 503 status if errors",
			statusCode: http.StatusServiceUnavailable,
			args: []Option{
				WithChecker("database", CheckerFunc(func() error {
					return fmt.Errorf("connection to db timed out")
				})),
				WithChecker("testService", CheckerFunc(func() error {
					return fmt.Errorf("connection refused")
				})),
			},
			response: response{
				Status: http.StatusText(http.StatusServiceUnavailable),
				Errors: map[string]string{
					"database":    "connection to db timed out",
					"testService": "connection refused",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost/health", nil)
			if err != nil {
				t.Errorf("Failed to create request.")
			}
			HandlerFunc(tt.args...)(res, req)
			if res.Code != tt.statusCode {
				t.Errorf("expected code %d, got %d", tt.statusCode, res.Code)
			}
			var respBody response
			if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
				t.Fatal("failed to parse the body")
			}
			if !reflect.DeepEqual(respBody, tt.response) {
				t.Errorf("NewHandlerFunc() = %v, want %v", respBody, tt.response)
			}
		})
	}
}
