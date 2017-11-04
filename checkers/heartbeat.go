package checkers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/etherlabsio/healthcheck"
)

type heartbeat struct {
	path string
}

func (h *heartbeat) Check(ctx context.Context) error {
	if _, err := os.Stat(h.path); err != nil {
		return errors.New("heartbeat not found. application should be out of rotation")
	}
	return nil
}

// Heartbeat returns a heartbeat health checker. Heartbeat files are generally used to take hosts out of rotation from the loadbalancers.
// Removing the heartbeat file allows you to debug the application host in case of failures.
func Heartbeat(filepath string) healthcheck.Checker {
	return &heartbeat{absFilePath(filepath)}
}

func absFilePath(inPath string) string {
	if strings.HasPrefix(inPath, "$") {
		end := strings.Index(inPath, string(os.PathSeparator))
		inPath = os.Getenv(inPath[1:end]) + inPath[end:]
	}
	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}
	if p, err := filepath.Abs(inPath); err == nil {
		return filepath.Clean(p)
	}
	return ""
}
