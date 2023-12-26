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
			name: "should upload a new file without a password",
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
		{
			name: "should fail with a nil hashFunc() service if password is set",
			args: domain.UploadFileArgs{
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
			name: "should fail if hashing the password fails",
			args: domain.UploadFileArgs{
				ID:       "file1",
				Name:     "file1",
				Owner:    "user1",
				Reader:   io.NopCloser(strings.NewReader("file-data")),
				Now:      func() time.Time { return time.Unix(1, 0).UTC() },
				Password: "secret-password",
				HashFunc: func(string) (string, error) {
					return "", fmt.Errorf("hash err")
				},
			},
			wantErr: fmt.Errorf("hash err"),
		},
		{
			name: "should upload a new file with a password",
			args: domain.UploadFileArgs{
				ID:       "file1",
				Name:     "file1",
				Owner:    "user1",
				Reader:   io.NopCloser(strings.NewReader("file-data")),
				Now:      func() time.Time { return time.Unix(1, 0).UTC() },
				Password: "secret-password",
				HashFunc: func(string) (string, error) {
					return "hashed-password", nil
				},
			},
			want: domain.File{
				FileHeader: domain.FileHeader{
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
			name: "should fail if expires-in is an invalid format",
			args: domain.UploadFileArgs{
				ID:        "file1",
				Name:      "file1",
				Owner:     "user1",
				Reader:    io.NopCloser(strings.NewReader("file-data")),
				Now:       func() time.Time { return time.Unix(1, 0).UTC() },
				ExpiresIn: "not-a-valid-long-duration",
			},
			wantErr: fmt.Errorf(`time: invalid duration "not-a-valid-long-duration"`),
		},
		{
			name: "should fail if expires-in is negative",
			args: domain.UploadFileArgs{
				ID:        "file1",
				Name:      "file1",
				Owner:     "user1",
				Reader:    io.NopCloser(strings.NewReader("file-data")),
				Now:       func() time.Time { return time.Unix(1, 0).UTC() },
				ExpiresIn: "-1h",
			},
			wantErr: fmt.Errorf("expiration cannot be set in the past"),
		},
		{
			name: "should upload a new file that expires 3.5 days from now",
			args: domain.UploadFileArgs{
				ID:        "file1",
				Name:      "file1",
				Owner:     "user1",
				Reader:    io.NopCloser(strings.NewReader("file-data")),
				Now:       func() time.Time { return time.Unix(0, 123456789).UTC() },
				ExpiresIn: "3.5d",
			},
			want: domain.File{
				FileHeader: domain.FileHeader{
					ID:      "file1",
					Name:    "file1",
					Owner:   "user1",
					Created: time.Unix(0, 123456789).UTC(),
					Expires: time.Unix((3*60*60*24)+(30*60*24), 123456789).UTC(),
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
				t.Errorf("UploadFile() got = \n\t%#v, want \n\t%#v", got, tt.want)
			}
		})
	}
}

func TestFile_Download(t *testing.T) {
	type args struct {
		user      domain.UserID
		password  string
		matchFunc domain.PasswordMatchFunc
		nowFunc   domain.NowFunc
	}
	tests := []struct {
		name    string
		f       domain.File
		args    args
		wantErr error
	}{
		{
			name: "should fail if a matchFunc() service was not passed in",
			f:    domain.File{},
			args: args{
				matchFunc: nil,
			},
			wantErr: fmt.Errorf("matchFunc() service cannot be empty"),
		},
		{
			name: "should fail if a now() service was not passed in",
			f:    domain.File{},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   nil,
			},
			wantErr: fmt.Errorf("now() service cannot be empty"),
		},
		{
			name: "should fail if file is password-protected but no password is supplied",
			f: domain.File{
				FileHeader: domain.FileHeader{
					PasswordHash: "password-hash",
				},
			},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			wantErr: domain.ErrFilePasswordRequired,
		},
		{
			name: "should fail if file is password-protected but the matchFunc() service returns an error",
			f: domain.File{
				FileHeader: domain.FileHeader{
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
			f: domain.File{
				FileHeader: domain.FileHeader{
					PasswordHash: "password-hash",
				},
			},
			args: args{
				password:  "incorrect-password",
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
			wantErr: domain.ErrFilePasswordInvalid,
		},
		{
			name: "should succeed if file is password-protected but the user is the owner of the file and no password is sent",
			f: domain.File{
				FileHeader: domain.FileHeader{
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
		},
		{
			name: "should fail if file has expired",
			f: domain.File{
				FileHeader: domain.FileHeader{
					Expires: time.Unix(1, 0).UTC(),
				},
			},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc: func() time.Time {
					return time.Unix(2, 0).UTC()
				},
			},
			wantErr: domain.ErrFileExpired,
		},
		{
			name: "should succeed if the file has no password or expiration time",
			f:    domain.File{},
			args: args{
				matchFunc: func(string, string) (bool, error) { return false, nil },
				nowFunc:   func() time.Time { return time.Unix(0, 0).UTC() },
			},
		},
		{
			name: "should succeed if the password matches and the file is not yet expired",
			f: domain.File{
				FileHeader: domain.FileHeader{
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f.Download(tt.args.user, tt.args.password, tt.args.matchFunc, tt.args.nowFunc)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
