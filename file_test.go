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
	type args struct {
		id     domain.FileID
		name   string
		owner  domain.UserID
		reader io.ReadCloser
		size   int64
		now    func() time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    domain.File
		wantErr error
	}{
		{
			name: "should fail with empty file ID",
			args: args{
				id: "",
			},
			wantErr: fmt.Errorf("file ID cannot be empty"),
		},
		{
			name: "should fail with empty file name",
			args: args{
				id:   "file1",
				name: "",
			},
			wantErr: fmt.Errorf("file name cannot be empty"),
		},
		{
			name: "should fail with empty owner",
			args: args{
				id:    "file1",
				name:  "file1",
				owner: "",
			},
			wantErr: fmt.Errorf("file owner cannot be empty"),
		},
		{
			name: "should fail with a nil reader",
			args: args{
				id:     "file1",
				name:   "file1",
				owner:  "user1",
				reader: nil,
			},
			wantErr: fmt.Errorf("file reader cannot be empty"),
		},
		{
			name: "should fail with a nil now() service",
			args: args{
				id:     "file1",
				name:   "file1",
				owner:  "user1",
				reader: io.NopCloser(strings.NewReader("file-data")),
				now:    nil,
			},
			wantErr: fmt.Errorf("now() service cannot be empty"),
		},
		{
			name: "should upload a new file",
			args: args{
				id:     "file1",
				name:   "file1",
				owner:  "user1",
				reader: io.NopCloser(strings.NewReader("file-data")),
				now:    func() time.Time { return time.Unix(1, 0).UTC() },
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
			got, err := domain.UploadFile(tt.args.id, tt.args.name, tt.args.owner, tt.args.reader, tt.args.size, tt.args.now)
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
