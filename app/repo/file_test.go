package repo_test

import (
	"context"
	"fmt"
	"git.jfam.app/blinkfile/app/repo"
	"git.jfam.app/blinkfile/domain"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	fileDirs   map[string]struct{}
	fileDirNum int
	fileNumMu  sync.Mutex
)

func newFileDir(t *testing.T, dirName string) string {
	t.Helper()
	const testDir = "./_test/repo_file"
	fileNumMu.Lock()
	defer fileNumMu.Unlock()
	if dirName == "" {
		fileDirNum++
		dirName = fmt.Sprintf("_%d", fileDirNum)
	}
	if _, ok := fileDirs[dirName]; ok {
		t.Fatalf("file dir %q already used for another test", dirName)
	}
	return filepath.Clean(fmt.Sprintf("%s/%s", testDir, dirName))
}

func TestNewFileRepo(t *testing.T) {
	type args struct {
		cfg repo.FileRepoConfig
	}
	tests := []struct {
		name     string
		mkdirAll func(path string, perm os.FileMode) error
		args     args
		wantErr  error
	}{
		{
			name: "should fail if making directory fails",
			mkdirAll: func(string, os.FileMode) error {
				return fmt.Errorf("mkdir err")
			},
			args: args{
				cfg: repo.FileRepoConfig{Dir: newFileDir(t, "new1")},
			},
			wantErr: fmt.Errorf(`making directory %q: %w`, filepath.Clean("_test/repo_file/new1"), fmt.Errorf("mkdir err")),
		},
		{
			name: "should create a new file repo",
			args: args{
				cfg: repo.FileRepoConfig{Dir: newFileDir(t, "new1")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mkdirAll != nil {
				prev := repo.MkdirAll
				repo.MkdirAll = tt.mkdirAll
				defer func() { repo.MkdirAll = prev }()
			}
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
		name     string
		cfg      repo.FileRepoConfig
		mkdirAll func(path string, perm os.FileMode) error
		args     args
		wantErr  error
	}{
		{
			name: "should fail if the file data is nil",
			cfg:  repo.FileRepoConfig{Dir: newFileDir(t, "")},
			args: args{
				file: domain.File{
					FileHeader: domain.FileHeader{
						ID: "file1",
					},
					Data: nil,
				},
			},
			wantErr: fmt.Errorf("file data cannot be nil"),
		},
		{
			name: "should fail if making the file directory fails",
			cfg:  repo.FileRepoConfig{Dir: newFileDir(t, "mkdirfail1")},
			args: args{
				file: domain.File{
					FileHeader: domain.FileHeader{
						ID: "file1",
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			mkdirAll: func(path string, perm os.FileMode) error {
				if path == filepath.Clean("_test/repo_file/mkdirfail1/file1") {
					return fmt.Errorf("mkdir err")
				}
				return os.MkdirAll(path, perm)
			},
			wantErr: fmt.Errorf(`making directory %q: %w`, filepath.Clean("_test/repo_file/mkdirfail1/file1"), fmt.Errorf("mkdir err")),
		},
		{
			name: "should save a file",
			cfg:  repo.FileRepoConfig{Dir: newFileDir(t, "")},
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
			if tt.mkdirAll != nil {
				prev := repo.MkdirAll
				repo.MkdirAll = tt.mkdirAll
				defer func() { repo.MkdirAll = prev }()
			}
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
