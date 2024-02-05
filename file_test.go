package blinkfile_test

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/benjohns1/blinkfile"
)

func TestUploadFile(t *testing.T) {
	tests := []struct {
		name    string
		args    blinkfile.UploadFileArgs
		want    blinkfile.File
		wantErr error
	}{
		{
			name: "should fail with empty file ID",
			args: blinkfile.UploadFileArgs{
				ID: "",
			},
			wantErr: fmt.Errorf("file ID cannot be empty"),
		},
		{
			name: "should fail with empty file name",
			args: blinkfile.UploadFileArgs{
				ID:   "file1",
				Name: "",
			},
			wantErr: fmt.Errorf("file name cannot be empty"),
		},
		{
			name: "should fail with empty owner",
			args: blinkfile.UploadFileArgs{
				ID:    "file1",
				Name:  "file1",
				Owner: "",
			},
			wantErr: fmt.Errorf("file owner cannot be empty"),
		},
		{
			name: "should fail with a nil reader",
			args: blinkfile.UploadFileArgs{
				ID:     "file1",
				Name:   "file1",
				Owner:  "user1",
				Reader: nil,
			},
			wantErr: fmt.Errorf("file reader cannot be empty"),
		},
		{
			name: "should fail with a nil now() service",
			args: blinkfile.UploadFileArgs{
				ID:     "file1",
				Name:   "file1",
				Owner:  "user1",
				Reader: io.NopCloser(strings.NewReader("file-data")),
				Now:    nil,
			},
			wantErr: fmt.Errorf("now() service cannot be empty"),
		},
		{
			name: "should upload a new file without a password",
			args: blinkfile.UploadFileArgs{
				ID:     "file1",
				Name:   "file1",
				Owner:  "user1",
				Reader: io.NopCloser(strings.NewReader("file-data")),
				Now:    func() time.Time { return time.Unix(1, 0).UTC() },
			},
			want: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					ID:      "file1",
					Name:    "file1",
					Owner:   "user1",
					Created: time.Unix(1, 0).UTC(),
				},
				Data: io.NopCloser(strings.NewReader("file-data")),
			},
		},
		{
			name: "should fail with a nil hashFunc() service if password is set",
			args: blinkfile.UploadFileArgs{
				ID:       "file1",
				Name:     "file1",
				Owner:    "user1",
				Reader:   io.NopCloser(strings.NewReader("file-data")),
				Now:      func() time.Time { return time.Unix(1, 0).UTC() },
				Password: "secret-password",
			},
			wantErr: fmt.Errorf("a password is set, so hashFunc() service cannot be empty"),
		},
		{
			name: "should upload a new file with a password",
			args: blinkfile.UploadFileArgs{
				ID:       "file1",
				Name:     "file1",
				Owner:    "user1",
				Reader:   io.NopCloser(strings.NewReader("file-data")),
				Now:      func() time.Time { return time.Unix(1, 0).UTC() },
				Password: "secret-password",
				HashFunc: func(string) string {
					return "hashed-password"
				},
			},
			want: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					ID:           "file1",
					Name:         "file1",
					Owner:        "user1",
					Created:      time.Unix(1, 0).UTC(),
					PasswordHash: "hashed-password",
				},
				Data: io.NopCloser(strings.NewReader("file-data")),
			},
		},
		{
			name: "should fail if expiration is in the past",
			args: blinkfile.UploadFileArgs{
				ID:      "file1",
				Name:    "file1",
				Owner:   "user1",
				Reader:  io.NopCloser(strings.NewReader("file-data")),
				Now:     func() time.Time { return time.Unix(1, 0).UTC() },
				Expires: time.Unix(0, 0).UTC(),
			},
			wantErr: fmt.Errorf("expiration cannot be set in the past"),
		},
		{
			name: "should fail if expiration is right now",
			args: blinkfile.UploadFileArgs{
				ID:      "file1",
				Name:    "file1",
				Owner:   "user1",
				Reader:  io.NopCloser(strings.NewReader("file-data")),
				Now:     func() time.Time { return time.Unix(0, 0).UTC() },
				Expires: time.Unix(0, 0).UTC(),
			},
			wantErr: fmt.Errorf("expiration cannot be set in the past"),
		},
		{
			name: "should fail with a negative download limit",
			args: blinkfile.UploadFileArgs{
				ID:            "file1",
				Name:          "file1",
				Owner:         "user1",
				Reader:        io.NopCloser(strings.NewReader("file-data")),
				Now:           func() time.Time { return time.Unix(0, 0).UTC() },
				DownloadLimit: -1,
			},
			wantErr: fmt.Errorf("download limit cannot be negative"),
		},
		{
			name: "should upload a new file with an expiration time",
			args: blinkfile.UploadFileArgs{
				ID:      "file1",
				Name:    "file1",
				Owner:   "user1",
				Reader:  io.NopCloser(strings.NewReader("file-data")),
				Now:     func() time.Time { return time.Unix(0, 0).UTC() },
				Expires: time.Unix(1, 0).UTC(),
			},
			want: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					ID:      "file1",
					Name:    "file1",
					Owner:   "user1",
					Created: time.Unix(0, 0).UTC(),
					Expires: time.Unix(1, 0).UTC(),
				},
				Data: io.NopCloser(strings.NewReader("file-data")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := blinkfile.UploadFile(tt.args)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("UploadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UploadFile() got = \n\t%#v, want \n\t%#v", got, tt.want)
			}
		})
	}
}

func TestFile_Download(t *testing.T) {
	type args struct {
		user      blinkfile.UserID
		password  string
		matchFunc blinkfile.PasswordMatchFunc
		nowFunc   blinkfile.NowFunc
	}
	tests := []struct {
		name    string
		f       blinkfile.File
		args    args
		wantErr error
		want    *blinkfile.File
	}{
		{
			name: "should fail if a matchFunc() service was not passed in",
			f:    blinkfile.File{},
			args: args{
				matchFunc: nil,
			},
			wantErr: fmt.Errorf("matchFunc() service cannot be empty"),
		},
		{
			name: "should fail if a now() service was not passed in",
			f:    blinkfile.File{},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   nil,
			},
			wantErr: fmt.Errorf("now() service cannot be empty"),
		},
		{
			name: "should fail if file is password-protected but no password is supplied",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					PasswordHash: "password-hash",
				},
			},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			wantErr: blinkfile.ErrFilePasswordRequired,
		},
		{
			name: "should fail if file is password-protected but the matchFunc() service returns an error",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					PasswordHash: "password-hash",
				},
			},
			args: args{
				password: "provided-password",
				matchFunc: func(string, string) (bool, error) {
					return false, fmt.Errorf("hash match err")
				},
				nowFunc: func() time.Time { return time.Unix(0, 0).UTC() },
			},
			wantErr: fmt.Errorf("hash match err"),
		},
		{
			name: "should fail if file is password-protected but passwords do not match",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					PasswordHash: "password-hash",
				},
			},
			args: args{
				password:  "incorrect-password",
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			wantErr: blinkfile.ErrFilePasswordInvalid,
		},
		{
			name: "should succeed if file is password-protected but the user is the owner of the file and no password is sent",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Owner:        "user1",
					PasswordHash: "password-hash",
				},
			},
			args: args{
				password:  "",
				user:      "user1",
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			want: &blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Owner:        "user1",
					PasswordHash: "password-hash",
					Downloads:    1,
				},
			},
		},
		{
			name: "should fail if file has expired",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Expires: time.Unix(1, 0).UTC(),
				},
			},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc: func() time.Time {
					return time.Unix(2, 0).UTC()
				},
			},
			wantErr: blinkfile.ErrFileExpired,
		},
		{
			name: "should succeed if the file has no password or expiration time",
			f:    blinkfile.File{},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			want: &blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Downloads: 1,
				},
			},
		},
		{
			name: "should succeed if the password matches and the file is not yet expired",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Owner:        "file-owner",
					Expires:      time.Unix(2, 0).UTC(),
					PasswordHash: "password-hash",
				},
			},
			args: args{
				user:     "other-user",
				password: "correct-password",
				matchFunc: func(string, string) (bool, error) {
					return true, nil
				},
				nowFunc: func() time.Time {
					return time.Unix(1, 0).UTC()
				},
			},
			want: &blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Owner:        "file-owner",
					Expires:      time.Unix(2, 0).UTC(),
					PasswordHash: "password-hash",
					Downloads:    1,
				},
			},
		},
		{
			name: "should fail if the file downloads have reached the limit",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Downloads:     1,
					DownloadLimit: 1,
				},
			},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			wantErr: blinkfile.ErrDownloadLimitReached,
			want: &blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Downloads:     1,
					DownloadLimit: 1,
				},
			},
		},
		{
			name: "should succeed if the file downloads are within the limit",
			f: blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Downloads:     1,
					DownloadLimit: 2,
				},
			},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			want: &blinkfile.File{
				FileHeader: blinkfile.FileHeader{
					Downloads:     2,
					DownloadLimit: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f.Download(tt.args.user, tt.args.password, tt.args.matchFunc, tt.args.nowFunc)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && !reflect.DeepEqual(tt.f, *tt.want) {
				t.Errorf("Download() changed file to:\n\t%+v\nwant:\n\t%+v", tt.f, *tt.want)
			}
		})
	}
}
