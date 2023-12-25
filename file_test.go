package domain_test

import (
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestUploadFile(t *testing.T) {
	tests := []struct {
		name    string
		args    domain.UploadFileArgs
		want    domain.File
		wantErr error
	}{
		{
			name: "should fail with empty file ID",
			args: domain.UploadFileArgs{
				ID: "",
			},
			wantErr: fmt.Errorf("file ID cannot be empty"),
		},
		{
			name: "should fail with empty file name",
			args: domain.UploadFileArgs{
				ID:   "file1",
				Name: "",
			},
			wantErr: fmt.Errorf("file name cannot be empty"),
		},
		{
			name: "should fail with empty owner",
			args: domain.UploadFileArgs{
				ID:    "file1",
				Name:  "file1",
				Owner: "",
			},
			wantErr: fmt.Errorf("file owner cannot be empty"),
		},
		{
			name: "should fail with a nil reader",
			args: domain.UploadFileArgs{
				ID:     "file1",
				Name:   "file1",
				Owner:  "user1",
				Reader: nil,
			},
			wantErr: fmt.Errorf("file reader cannot be empty"),
		},
		{
			name: "should fail with a nil now() service",
			args: domain.UploadFileArgs{
				ID:     "file1",
				Name:   "file1",
				Owner:  "user1",
				Reader: io.NopCloser(strings.NewReader("file-data")),
				Now:    nil,
			},
			wantErr: fmt.Errorf("now() service cannot be empty"),
		},
		{
			name: "should upload a new file",
			args: domain.UploadFileArgs{
				ID:     "file1",
				Name:   "file1",
				Owner:  "user1",
				Reader: io.NopCloser(strings.NewReader("file-data")),
				Now:    func() time.Time { return time.Unix(1, 0).UTC() },
			},
			want: domain.File{
				FileHeader: domain.FileHeader{
					ID:      "file1",
					Name:    "file1",
					Owner:   "user1",
					Created: time.Unix(1, 0).UTC(),
				},
				Data: io.NopCloser(strings.NewReader("file-data")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.UploadFile(tt.args)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("UploadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UploadFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
