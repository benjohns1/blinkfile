package app_test

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestApp_ListFiles(t *testing.T) {
	ctx := context.Background()
	type args struct {
		owner blinkfile.UserID
	}
	tests := []struct {
		name    string
		cfg     app.Config
		args    args
		want    []blinkfile.FileHeader
		wantErr error
	}{
		{
			name: "should fail if owner is empty",
			args: args{
				"",
			},
			wantErr: &app.Error{
				Type: app.ErrBadRequest,
				Err:  fmt.Errorf("owner is required"),
			},
		},
		{
			name: "should fail with repo error",
			cfg: app.Config{
				FileRepo: &StubFileRepo{
					ListByUserFunc: func(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error) {
						return nil, fmt.Errorf("list file err")
					},
				},
			},
			args: args{
				"user1",
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("retrieving file list: %w", fmt.Errorf("list file err")),
			},
		},
		{
			name: "should sort files by created time descending, then name, then internal file ID",
			cfg: app.Config{
				FileRepo: &StubFileRepo{
					ListByUserFunc: func(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error) {
						return []blinkfile.FileHeader{
							{
								ID:      "1",
								Name:    "File1",
								Created: time.Unix(1, 0),
							},
							{
								ID:      "2",
								Name:    "File2",
								Created: time.Unix(2, 0),
							},
							{
								ID:      "3b",
								Name:    "File3b",
								Created: time.Unix(3, 0),
							},
							{
								ID:      "3a",
								Name:    "File3a",
								Created: time.Unix(3, 0),
							},
							{
								ID:      "4b",
								Name:    "File4",
								Created: time.Unix(4, 0),
							},
							{
								ID:      "4a",
								Name:    "File4",
								Created: time.Unix(4, 0),
							},
						}, nil
					},
				},
			},
			args: args{
				"user1",
			},
			want: []blinkfile.FileHeader{
				{
					ID:      "4a",
					Name:    "File4",
					Created: time.Unix(4, 0),
				},
				{
					ID:      "4b",
					Name:    "File4",
					Created: time.Unix(4, 0),
				},
				{
					ID:      "3a",
					Name:    "File3a",
					Created: time.Unix(3, 0),
				},
				{
					ID:      "3b",
					Name:    "File3b",
					Created: time.Unix(3, 0),
				},
				{
					ID:      "2",
					Name:    "File2",
					Created: time.Unix(2, 0),
				},
				{
					ID:      "1",
					Name:    "File1",
					Created: time.Unix(1, 0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			got, err := application.ListFiles(ctx, tt.args.owner)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListFiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApp_UploadFile(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		cfg     app.Config
		args    app.UploadFileArgs
		wantErr error
	}{
		{
			name: "should fail generating a file ID",
			cfg: app.Config{
				GenerateFileID: func() (blinkfile.FileID, error) {
					return "", fmt.Errorf("file ID err")
				},
			},
			wantErr: &app.Error{
				Type: app.ErrInternal,
				Err:  fmt.Errorf("generating file ID: %w", fmt.Errorf("file ID err")),
			},
		},
		{
			name: "should fail if expires-in and expires fields are both set",
			args: app.UploadFileArgs{
				ExpiresIn: "1w",
				Expires:   time.Unix(0, 0),
			},
			wantErr: &app.Error{
				Type:   app.ErrBadRequest,
				Title:  "Error validating file expiration",
				Detail: "Can only set one of the expiration fields at a time.",
			},
		},
		{
			name: "should fail if adding expires-in to current time fails",
			args: app.UploadFileArgs{
				ExpiresIn: "invalid-duration",
			},
			wantErr: &app.Error{
				Type:   app.ErrBadRequest,
				Title:  "Error calculating file expiration",
				Detail: "Expires In field is not in a valid format.",
				Err:    fmt.Errorf(`time: invalid duration "invalid-duration"`),
			},
		},
		{
			name: "should fail if a file argument is not valid, such as an empty filename",
			args: app.UploadFileArgs{
				Filename: "",
			},
			wantErr: &app.Error{
				Type: app.ErrBadRequest,
				Err:  fmt.Errorf("file name cannot be empty"),
			},
		},
		{
			name: "should fail if the repo fails to save the uploaded file",
			cfg: app.Config{
				FileRepo: &StubFileRepo{SaveFunc: func(context.Context, blinkfile.File) error {
					return fmt.Errorf("repo save err")
				}},
			},
			args: app.UploadFileArgs{
				Filename: "file1",
				Owner:    "user1",
				Reader:   io.NopCloser(strings.NewReader("file-data")),
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("repo save err"),
			},
		},
		{
			name: "should successfully upload a file",
			args: app.UploadFileArgs{
				Filename: "file1",
				Owner:    "user1",
				Reader:   io.NopCloser(strings.NewReader("file-data")),
			},
		},
		{
			name: "should successfully upload a file with a password",
			args: app.UploadFileArgs{
				Filename: "file1",
				Owner:    "user1",
				Reader:   io.NopCloser(strings.NewReader("file-data")),
				Password: "file-password",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			err := application.UploadFile(ctx, tt.args)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("UploadFile() error = \n\t%v\n, wantErr \n\t%v", err, tt.wantErr)
				return
			}
		})
	}
}

type StubPasswordHasher struct {
	HashFunc  func([]byte) string
	MatchFunc func(string, []byte) (bool, error)
}

var _ app.PasswordHasher = &StubPasswordHasher{}

func (ph *StubPasswordHasher) Hash(data []byte) (hash string) {
	if ph.HashFunc != nil {
		return ph.HashFunc(data)
	}
	return ""
}
func (ph *StubPasswordHasher) Match(hash string, data []byte) (matched bool, err error) {
	if ph.MatchFunc != nil {
		return ph.MatchFunc(hash, data)
	}
	return false, nil
}

func TestApp_DownloadFile(t *testing.T) {
	ctx := context.Background()
	type args struct {
		userID   blinkfile.UserID
		fileID   blinkfile.FileID
		password string
	}
	tests := []struct {
		name    string
		cfg     app.Config
		args    args
		want    blinkfile.FileHeader
		wantErr error
	}{
		{
			name: "should fail if file ID is empty",
			args: args{
				fileID: "",
			},
			wantErr: &app.Error{
				Type: app.ErrBadRequest,
				Err:  fmt.Errorf("file ID is required"),
			},
		},
		{
			name: "should fail if repo retrieval fails",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{}, fmt.Errorf("get err")
				}},
			},
			args: args{
				fileID: "file1",
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("get err"),
			},
		},
		{
			name: "should fail with an authorization error if file is not found",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{}, app.ErrFileNotFound
				}},
			},
			args: args{
				fileID: "file1",
			},
			wantErr: &app.Error{
				Type: app.ErrAuthzFailed,
				Err:  blinkfile.ErrFilePasswordRequired,
			},
		},
		{
			name: "should fail with an authorization error if file is not found but a password is supplied",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{}, app.ErrFileNotFound
				}},
			},
			args: args{
				fileID:   "file1",
				password: "file-password",
			},
			wantErr: &app.Error{
				Type: app.ErrAuthzFailed,
				Err:  blinkfile.ErrFilePasswordInvalid,
			},
		},
		{
			name: "should fail with an authorization error if file is expired",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{}, blinkfile.ErrFileExpired
				}},
			},
			args: args{
				fileID: "file1",
			},
			wantErr: &app.Error{
				Type: app.ErrAuthzFailed,
				Err:  blinkfile.ErrFilePasswordRequired,
			},
		},
		{
			name: "should fail with an authorization error if password is invalid",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{
						PasswordHash: "password-hash",
					}, nil
				}},
				PasswordHasher: &StubPasswordHasher{
					MatchFunc: func(string, []byte) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				fileID:   "file1",
				password: "invalid-password",
			},
			wantErr: &app.Error{
				Type: app.ErrAuthzFailed,
				Err:  blinkfile.ErrFilePasswordInvalid,
			},
		},
		{
			name: "should download a public file with minimal fields",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{
						ID: "file1",
					}, nil
				}},
			},
			args: args{
				fileID: "file1",
			},
			want: blinkfile.FileHeader{
				ID: "file1",
			},
		},
		{
			name: "should download a file with minimal fields that is password protected",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{
						ID:           "file1",
						PasswordHash: "password-hash",
					}, nil
				}},
				PasswordHasher: &StubPasswordHasher{
					MatchFunc: func(string, []byte) (bool, error) {
						return true, nil
					},
				},
			},
			args: args{
				fileID:   "file1",
				password: "correct-password",
			},
			want: blinkfile.FileHeader{
				ID:           "file1",
				PasswordHash: "password-hash",
			},
		},
		{
			name: "should download a file with all available fields",
			cfg: app.Config{
				FileRepo: &StubFileRepo{GetFunc: func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error) {
					return blinkfile.FileHeader{
						ID:           "file1",
						Name:         "filename",
						Location:     "location",
						Owner:        "user1",
						Created:      time.Unix(1, 0).UTC(),
						Expires:      time.Unix(2, 0).UTC(),
						Size:         3,
						PasswordHash: "password-hash",
					}, nil
				}},
				PasswordHasher: &StubPasswordHasher{
					MatchFunc: func(string, []byte) (bool, error) {
						return true, nil
					},
				},
				Now: func() time.Time { return time.Unix(1, 0).UTC() },
			},
			args: args{
				fileID:   "file1",
				password: "correct-password",
			},
			want: blinkfile.FileHeader{
				ID:           "file1",
				Name:         "filename",
				Location:     "location",
				Owner:        "user1",
				Created:      time.Unix(1, 0).UTC(),
				Expires:      time.Unix(2, 0).UTC(),
				Size:         3,
				PasswordHash: "password-hash",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			got, err := application.DownloadFile(ctx, tt.args.userID, tt.args.fileID, tt.args.password)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("DownloadFile() error = \n\t%v\n, wantErr \n\t%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DownloadFile() got:\n\t%+v\nwant:\n\t%+v", got, tt.want)
			}
		})
	}
}

func TestApp_DeleteFiles(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx         context.Context
		owner       blinkfile.UserID
		deleteFiles []blinkfile.FileID
	}
	tests := []struct {
		name    string
		cfg     app.Config
		args    args
		wantErr error
	}{
		{
			name: "should fail if owner ID is missing",
			args: args{
				owner: "",
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("owner is required"),
			},
		},
		{
			name: "should fail if repo returns an error",
			cfg: app.Config{
				FileRepo: &StubFileRepo{
					DeleteFunc: func(context.Context, blinkfile.UserID, []blinkfile.FileID) error {
						return fmt.Errorf("delete err")
					},
				},
			},
			args: args{
				owner: "user1",
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("delete err"),
			},
		},
		{
			name: "should successfully delete files",
			args: args{
				owner: "user1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			err := application.DeleteFiles(tt.args.ctx, tt.args.owner, tt.args.deleteFiles)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("DeleteFiles() error = \n\t%v\n, wantErr \n\t%v", err, tt.wantErr)
				return
			}
		})
	}
}
