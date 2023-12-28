package repo_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app/repo"
	"git.jfam.app/one-way-file-send/domain"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	fileDirNum int
	fileNumMu  sync.Mutex
)

func newFileDir(t *testing.T) string {
	t.Helper()
	const testDir = "./_test/repo_file"
	fileNumMu.Lock()
	defer fileNumMu.Unlock()
	fileDirNum++
	return filepath.Clean(fmt.Sprintf("%s/%d", testDir, fileDirNum))
}

func TestNewFileRepo(t *testing.T) {
	type args struct {
		cfg repo.FileRepoConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "should create a new file repo",
			args: args{
				cfg: repo.FileRepoConfig{Dir: newFileDir(t)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanDir(t, tt.args.cfg.Dir)
			_, err := repo.NewFileRepo(context.Background(), tt.args.cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestFileRepo_Save(t *testing.T) {
	type args struct {
		file domain.File
	}
	tests := []struct {
		name    string
		cfg     repo.FileRepoConfig
		args    args
		wantErr error
	}{
		{
			name: "should save a file",
			cfg:  repo.FileRepoConfig{Dir: newFileDir(t)},
			args: args{
				file: domain.File{
					FileHeader: domain.FileHeader{
						ID:      "file1",
						Name:    "file1.txt",
						Owner:   "user1",
						Created: time.Unix(1, 0),
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanDir(t, tt.cfg.Dir)
			r, err := repo.NewFileRepo(context.Background(), tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			err = r.Save(context.Background(), tt.args.file)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
