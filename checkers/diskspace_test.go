package checkers

import (
	"fmt"
	"syscall"
	"testing"
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
			fmt.Errorf("space used on / greater than threshold 80%%(100%%)"),
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
			fmt.Errorf("space used on / greater than threshold 40%%(50%%)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &diskspace{
				dir:       tt.dir,
				threshold: tt.threshold,
				statfs: func(fs string, stat *syscall.Statfs_t) error {
					stat.Bsize = 1
					stat.Bfree = tt.freeBlocks
					stat.Blocks = tt.totalBlocks
					return nil
				},
			}

			if err := ds.Check(); err != tt.err {
				if err == nil || tt.err == nil || err.Error() != tt.err.Error() {
					t.Errorf("diskspace.Check() error = %v, wantErr %v", err, tt.err)
				}
			}
		})
	}
}
