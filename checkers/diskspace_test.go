package checkers

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/sys/unix"
)

func Test_diskspace_Check(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		threshold   uint64
		totalBlocks uint64
		freeBlocks  uint64
		err         error
	}{
		{
			"Filesystem Empty",
			"/",
			80,
			100,
			100,
			nil,
		},
		{
			"Filesystem full",
			"/",
			80,
			100,
			0,
			fmt.Errorf("used: 100%% threshold: 80%% location: /"),
		},
		{
			"Filesystem at 50%. Threshold 60%",
			"/",
			60,
			100,
			50,
			nil,
		},
		{
			"Filesystem at 50%. Threshold 40%",
			"/",
			40,
			100,
			50,
			fmt.Errorf("used: 50%% threshold: 40%% location: /"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &diskspace{
				dir:       tt.dir,
				threshold: tt.threshold,
				statfs: func(fs string, stat *unix.Statfs_t) error {
					stat.Bsize = 1
					stat.Bfree = tt.freeBlocks
					stat.Blocks = tt.totalBlocks
					return nil
				},
			}
			if err := ds.Check(context.Background()); err != tt.err {
				if err == nil || tt.err == nil || err.Error() != tt.err.Error() {
					t.Errorf("diskspace.Check() returned error = \"%v\" but expected \"%v\"", err, tt.err)
				}
			}
		})
	}
}
