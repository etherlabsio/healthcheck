package checkers

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/etherlabsio/healthcheck"
)

type diskspace struct {
	dir       string
	threshold uint64
	statfs    func(string, *syscall.Statfs_t) error
}

//Check test if the filesystem disk usage is above threshold
func (ds *diskspace) Check(ctx context.Context) error {
	if _, err := os.Stat(ds.dir); err != nil {
		return fmt.Errorf("filesystem not found: %v", err)
	}
	fs := syscall.Statfs_t{}
	err := ds.statfs(ds.dir, &fs)
	if err != nil {
		return fmt.Errorf("error looking for %s filesystem stats: %v", ds.dir, err)
	}

	total := fs.Blocks * uint64(fs.Bsize)
	free := fs.Bfree * uint64(fs.Bsize)
	used := total - free

	if 100*used/total > ds.threshold {
		return fmt.Errorf("space used on %s greater than threshold %d%%(%d%%)", ds.dir, ds.threshold, 100*used/total)
	}
	return nil
}

// DiskSpace returns a diskspace health checker, which checks if filesystem usage is above threshold
func DiskSpace(dir string, threshold uint64) healthcheck.Checker {
	return &diskspace{
		dir:       dir,
		threshold: threshold,
		statfs:    syscall.Statfs,
	}
}
