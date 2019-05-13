package checkers

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func Test_heartbeat_Check(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd unknown: %+v", err)
	}
	_, err = os.Create("heartbeat.txt")
	if err != nil {
		t.Fatalf("heartbeat file create failed: %+v", err)
	}
	fileName := "heartbeat.txt"
	filePath := fmt.Sprintf("%s/%s", cwd, fileName)
	defer os.Remove(filePath)
	type fields struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"if heartbeat file exists, no error should be returned",
			fields{
				absFilePath(filePath),
			},
			false,
		},
		{
			"if valid heartbeat filepath is set in env variable, no error should be returned",
			fields{
				func() string {
					os.Setenv("HBFILE_PATH", cwd)
					return absFilePath("$HBFILE_PATH" + "/" + fileName)
				}(),
			},
			false,
		},
		{
			"if heartbeat file does not exist, error should be returned",
			fields{"/etc/hosts1"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &heartbeat{
				path: tt.fields.path,
			}
			if err := h.Check(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("heartbeat.Check() returned error = %v but expected %v", err, tt.wantErr)
			}
		})
	}
}
