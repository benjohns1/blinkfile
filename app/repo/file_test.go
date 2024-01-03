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
		ctx context.Context
		cfg repo.FileRepoConfig
	}
	tests := []struct {
		name        string
		mkdirAll    func(path string, perm os.FileMode) error
		readFile    func(name string) ([]byte, error)
		args        args
		wantErr     error
		wantErrLogs []string
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
			name: "should halt building existing indexes if context is cancelled",
			args: args{
				ctx: func() context.Context {
					ctx, cancelFunc := context.WithCancel(context.Background())
					cancelFunc()
					return ctx
				}(),
				cfg: func() repo.FileRepoConfig {
					cfg := repo.FileRepoConfig{Dir: newFileDir(t, "haltIndexing1")}
					r, err := repo.NewFileRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					err = r.Save(context.Background(), domain.File{
						FileHeader: domain.FileHeader{
							ID:    "file1",
							Name:  "filename",
							Owner: "user1",
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					})
					if err != nil {
						t.Fatal(err)
					}
					return cfg
				}(),
			},
			wantErr: fmt.Errorf("context canceled"),
		},
		{
			name: "should still create a file repo even if reading an existing file header fails, but log an error",
			args: args{
				cfg: func() repo.FileRepoConfig {
					cfg := repo.FileRepoConfig{Dir: newFileDir(t, "readFileHeader1")}
					r, err := repo.NewFileRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					err = r.Save(context.Background(), domain.File{
						FileHeader: domain.FileHeader{
							ID:    "file1",
							Name:  "filename",
							Owner: "user1",
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					})
					if err != nil {
						t.Fatal(err)
					}
					return cfg
				}(),
			},
			readFile: func(name string) ([]byte, error) {
				return nil, fmt.Errorf("file read err")
			},
			wantErrLogs: []string{
				fmt.Sprintf("Loading file header %q: file read err", `_test\repo_file\readFileHeader1\file1\header.json`),
			},
		},
		{
			name: "should create a new file repo",
			args: args{
				cfg: repo.FileRepoConfig{Dir: newFileDir(t, "new1")},
			},
		},
		{
			name: "should create a new file repo with existing files",
			args: args{
				cfg: func() repo.FileRepoConfig {
					cfg := repo.FileRepoConfig{Dir: newFileDir(t, "existingFiles1")}
					r, err := repo.NewFileRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					err = r.Save(context.Background(), domain.File{
						FileHeader: domain.FileHeader{
							ID:    "file1",
							Name:  "filename",
							Owner: "user1",
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					})
					if err != nil {
						t.Fatal(err)
					}
					return cfg
				}(),
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
			if tt.readFile != nil {
				prev := repo.ReadFile
				repo.ReadFile = tt.readFile
				defer func() { repo.ReadFile = prev }()
			}
			spy := &spyLog{}
			if tt.args.cfg.Log == nil {
				tt.args.cfg.Log = spy
			}
			defer cleanDir(t, tt.args.cfg.Dir)
			if tt.args.ctx == nil {
				tt.args.ctx = context.Background()
			}
			_, err := repo.NewFileRepo(tt.args.ctx, tt.args.cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrLogs != nil && !reflect.DeepEqual(spy.errors, tt.wantErrLogs) {
				t.Errorf("NewSession() error logs = %v, wantErrLogs %v", spy.errors, tt.wantErrLogs)
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

type spyLog struct {
	errors []string
}

func (l *spyLog) Errorf(_ context.Context, format string, v ...any) {
	l.errors = append(l.errors, fmt.Sprintf(format, v...))
}
