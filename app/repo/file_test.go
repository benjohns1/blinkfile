package repo_test

import (
	"context"
	"fmt"
	"git.jfam.app/blinkfile"
	"git.jfam.app/blinkfile/app/repo"
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
		args        args
		patch       func(*testing.T) func()
		wantErr     error
		wantErrLogs []string
	}{
		{
			name: "should fail if making directory fails",
			args: args{
				cfg: repo.FileRepoConfig{Dir: newFileDir(t, "new1")},
			},
			patch: func(t *testing.T) func() {
				prev := repo.MkdirAll
				repo.MkdirAll = func(string, os.FileMode) error {
					return fmt.Errorf("mkdir err")
				}
				return func() { repo.MkdirAll = prev }
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
					cfg := repo.FileRepoConfig{Dir: newFileDir(t, "")}
					r, err := repo.NewFileRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					err = r.Save(context.Background(), blinkfile.File{
						FileHeader: blinkfile.FileHeader{
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
					err = r.Save(context.Background(), blinkfile.File{
						FileHeader: blinkfile.FileHeader{
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
			patch: func(t *testing.T) func() {
				prev := repo.ReadFile
				repo.ReadFile = func(name string) ([]byte, error) {
					return nil, fmt.Errorf("file read err")
				}
				return func() { repo.ReadFile = prev }
			},
			wantErrLogs: []string{
				fmt.Sprintf("Loading file header %q: file read err", filepath.Clean(`_test/repo_file/readFileHeader1/file1/header.json`)),
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
					cfg := repo.FileRepoConfig{Dir: newFileDir(t, "")}
					r, err := repo.NewFileRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					err = r.Save(context.Background(), blinkfile.File{
						FileHeader: blinkfile.FileHeader{
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
			if tt.patch != nil {
				defer tt.patch(t)()
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
		file blinkfile.File
	}
	tests := []struct {
		name    string
		r       *repo.FileRepo
		patch   func(*testing.T) func()
		args    args
		wantErr error
	}{
		{
			name: "should fail if the file data is nil",
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID: "file1",
					},
					Data: nil,
				},
			},
			wantErr: fmt.Errorf("file data cannot be nil"),
		},
		{
			name: "should fail if making the file directory fails",
			r:    newTestFileRepo(t, "mkdirfail1"),
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID: "file1",
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.MkdirAll
				repo.MkdirAll = func(path string, perm os.FileMode) error {
					if path == filepath.Clean("_test/repo_file/mkdirfail1/file1") {
						return fmt.Errorf("mkdir err")
					}
					return os.MkdirAll(path, perm)
				}
				return func() { repo.MkdirAll = prev }
			},
			wantErr: fmt.Errorf(`making directory %q: %w`, filepath.Clean("_test/repo_file/mkdirfail1/file1"), fmt.Errorf("mkdir err")),
		},
		{
			name: "should fail if marshaling the file header fails",
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID:      "file1",
						Name:    "file1.txt",
						Owner:   "user1",
						Created: time.Unix(1, 0),
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.Marshal
				repo.Marshal = func(any) ([]byte, error) {
					return nil, fmt.Errorf("marshal err")
				}
				return func() { repo.Marshal = prev }
			},
			wantErr: fmt.Errorf("marshaling file header: %w", fmt.Errorf("marshal err")),
		},
		{
			name: "should fail if writing the file header fails",
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID:      "file1",
						Name:    "file1.txt",
						Owner:   "user1",
						Created: time.Unix(1, 0),
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.WriteFile
				repo.WriteFile = func(name string, data []byte, perm os.FileMode) error {
					return fmt.Errorf("file write err")
				}
				return func() { repo.WriteFile = prev }
			},
			wantErr: fmt.Errorf("writing file header: %w", fmt.Errorf("file write err")),
		},
		{
			name: "should fail if creating the data file fails",
			r:    newTestFileRepo(t, "createFail"),
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID:      "id_of_file1",
						Name:    "file1.txt",
						Owner:   "user1",
						Created: time.Unix(1, 0),
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.CreateFile
				repo.CreateFile = func(string) (*os.File, error) {
					return nil, fmt.Errorf("create err")
				}
				return func() { repo.CreateFile = prev }
			},
			wantErr: fmt.Errorf(`creating file %q: %w`, filepath.Clean("_test/repo_file/createFail/id_of_file1/file"), fmt.Errorf("create err")),
		},
		{
			name: "should fail if copying the buffer for the data file fails",
			r:    newTestFileRepo(t, "bufferCopyFail"),
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID:      "id_of_file1",
						Name:    "file1.txt",
						Owner:   "user1",
						Created: time.Unix(1, 0),
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.Copy
				repo.Copy = func(io.Writer, io.Reader) (int64, error) {
					return 0, fmt.Errorf("copy err")
				}
				return func() { repo.Copy = prev }
			},
			wantErr: fmt.Errorf(`writing file %q: %w`, filepath.Clean("_test/repo_file/bufferCopyFail/id_of_file1/file"), fmt.Errorf("copy err")),
		},
		{
			name: "should save a file",
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
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
			if tt.patch != nil {
				defer tt.patch(t)()
			}
			if tt.r == nil {
				tt.r = newTestFileRepo(t, "")
			}
			defer cleanDir(t, tt.r.Dir())
			err := tt.r.Save(context.Background(), tt.args.file)
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

func newTestFileRepo(t *testing.T, dir string) *repo.FileRepo {
	r, err := repo.NewFileRepo(context.Background(), repo.FileRepoConfig{Dir: newFileDir(t, dir)})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFileRepo_DeleteExpiredBefore(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		r       *repo.FileRepo
		args    args
		want    int
		wantErr error
	}{
		{
			name:    "should do nothing if repo is empty",
			want:    0,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.r == nil {
				tt.r = newTestFileRepo(t, "")
			}
			defer cleanDir(t, tt.r.Dir())
			got, err := tt.r.DeleteExpiredBefore(ctx, tt.args.t)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("DeleteExpiredBefore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeleteExpiredBefore() got = %v, want %v", got, tt.want)
			}
		})
	}
}
