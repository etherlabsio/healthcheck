package checkers

import (
	"context"
	"fmt"
	"os"

	"github.com/etherlabsio/healthcheck"
	"golang.org/x/sys/unix"
)

type diskspace struct {
	dir       string
	threshold uint64
	statfs    func(string, *unix.Statfs_t) error
}

//Check test if the filesystem disk usage is above threshold
func (ds *diskspace) Check(ctx context.Context) error {
	if _, err := os.Stat(ds.dir); err != nil {
		return fmt.Errorf("filesystem not found: %v", err)
	}

	fs := unix.Statfs_t{}
	err := ds.statfs(ds.dir, &fs)
	if err != nil {
		return fmt.Errorf("error looking for %s filesystem stats: %v", ds.dir, err)
	}

	total := fs.Blocks * uint64(fs.Bsize)
	free := fs.Bfree * uint64(fs.Bsize)
	used := total - free
	usedPercentage := 100 * used / total
	if usedPercentage > ds.threshold {
		return fmt.Errorf("used: %d%% threshold: %d%% location: %s", usedPercentage, ds.threshold, ds.dir)
	}
	return nil
}

// DiskSpace returns a diskspace health checker, which checks if filesystem usage is above the threshold which is defined in percentage.
func DiskSpace(dir string, threshold uint64) healthcheck.Checker {
	return &diskspace{
		dir:       dir,
		threshold: threshold,
		statfs:    unix.Statfs,
	}
}
