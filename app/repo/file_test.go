package repo_test

import (
	"context"
	"fmt"
	"git.jfam.app/blinkfile"
	"git.jfam.app/blinkfile/app"
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
			name: "should fail if the file ID is empty",
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID: "",
					},
				},
			},
			wantErr: fmt.Errorf("file ID cannot be empty"),
		},
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
			name: "should fail if the file owner is empty",
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID:    "file1",
						Owner: "",
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				},
			},
			wantErr: fmt.Errorf("file owner cannot be empty"),
		},
		{
			name: "should fail if making the file directory fails",
			r:    newTestFileRepo(t, "mkdirfail1"),
			args: args{
				file: blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID:    "file1",
						Owner: "owner1",
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

func fatalOnErr(t *testing.T, errs ...error) {
	t.Helper()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestFileRepo_DeleteExpiredBefore(t *testing.T) {
	ctx := context.Background()
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		patch   func(*testing.T) func()
		r       *repo.FileRepo
		args    args
		want    int
		wantErr error
		assert  func(*testing.T, *repo.FileRepo)
	}{
		{
			name: "should do nothing if repo is empty",
			want: 0,
		},
		{
			name: "should do nothing if all files have expirations in the future or no expiration",
			r: func() *repo.FileRepo {
				r := newTestFileRepo(t, "deleteExpiredBefore_noExpiration")
				fatalOnErr(t,
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:      "file1",
							Owner:   "user1",
							Expires: time.Time{},
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:      "file2",
							Owner:   "user1",
							Expires: time.Unix(1, 1),
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
				)
				return r
			}(),
			args: args{
				t: time.Unix(1, 0),
			},
			want: 0,
			assert: func(t *testing.T, r *repo.FileRepo) {
				want := []blinkfile.FileHeader{
					{
						ID:       "file1",
						Location: filepath.Clean("_test/repo_file/deleteExpiredBefore_noExpiration/file1/file"),
						Owner:    "user1",
						Expires:  time.Time{},
					},
					{
						ID:       "file2",
						Location: filepath.Clean("_test/repo_file/deleteExpiredBefore_noExpiration/file2/file"),
						Owner:    "user1",
						Expires:  time.Unix(1, 1),
					},
				}

				got, err := r.ListByUser(ctx, "user1")
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("After DeleteExpiredBefore(), ListByUser() for user1 got: \n\t%+v\nwant: \n\t%+v", got, want)
				}
			},
		},
		{
			name: "should fail if deleting an expired file fails, but still return the number of deleted files up to that point",
			patch: func(t *testing.T) func() {
				prev := repo.RemoveAll
				var count int
				repo.RemoveAll = func(path string) error {
					if count == 0 {
						count++
						return nil
					}
					return fmt.Errorf("remove file failure")
				}
				return func() { repo.RemoveAll = prev }
			},
			r: func() *repo.FileRepo {
				r := newTestFileRepo(t, "deleteExpiredBefore_deleteFailure")
				fatalOnErr(t,
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:      "file1",
							Owner:   "user1",
							Expires: time.Unix(0, 0),
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:      "file2",
							Owner:   "user1",
							Expires: time.Unix(1, 0),
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
				)
				return r
			}(),
			args: args{
				t: time.Unix(2, 0),
			},
			want:    1,
			wantErr: fmt.Errorf("remove file failure"),
			assert: func(t *testing.T, r *repo.FileRepo) {
				want := []blinkfile.FileHeader{
					{
						ID:       "file2",
						Location: filepath.Clean("_test/repo_file/deleteExpiredBefore_deleteFailure/file2/file"),
						Owner:    "user1",
						Expires:  time.Unix(1, 0),
					},
				}

				got, err := r.ListByUser(ctx, "user1")
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("After DeleteExpiredBefore(), ListByUser() for user1 got: \n\t%+v\nwant: \n\t%+v", got, want)
				}
			},
		},
		{
			name: "should delete all files older than or equal to the given time",
			r: func() *repo.FileRepo {
				r := newTestFileRepo(t, "deleteExpiredBefore_success")
				fatalOnErr(t,
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:      "file1",
							Owner:   "user1",
							Expires: time.Unix(0, 0),
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:      "file2",
							Owner:   "user1",
							Expires: time.Unix(1, 0),
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
				)
				return r
			}(),
			args: args{
				t: time.Unix(1, 0),
			},
			want: 2,
			assert: func(t *testing.T, r *repo.FileRepo) {
				want := []blinkfile.FileHeader{}

				got, err := r.ListByUser(ctx, "user1")
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("After DeleteExpiredBefore(), ListByUser() for user1 got: \n\t%+v\nwant: \n\t%+v", got, want)
				}
			},
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
			got, err := tt.r.DeleteExpiredBefore(ctx, tt.args.t)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("DeleteExpiredBefore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeleteExpiredBefore() got = %v, want %v", got, tt.want)
			}
			if tt.assert != nil {
				tt.assert(t, tt.r)
			}
		})
	}
}

func TestFileRepo_Get(t *testing.T) {
	ctx := context.Background()
	type args struct {
		fileID blinkfile.FileID
	}
	tests := []struct {
		name    string
		r       *repo.FileRepo
		args    args
		want    blinkfile.FileHeader
		wantErr error
	}{
		{
			name: "should fail if file ID is empty",
			args: args{
				fileID: "",
			},
			wantErr: fmt.Errorf("file ID cannot be empty"),
		},
		{
			name: "should fail if file not found",
			args: args{
				fileID: "file1",
			},
			wantErr: app.ErrFileNotFound,
		},
		{
			name: "should return file header with its location",
			r: func() *repo.FileRepo {
				r := newTestFileRepo(t, "get_withLocation")
				fatalOnErr(t, r.Save(ctx, blinkfile.File{
					FileHeader: blinkfile.FileHeader{
						ID:    "file1",
						Owner: "user1",
					},
					Data: io.NopCloser(strings.NewReader("file-data")),
				}))
				return r
			}(),
			args: args{
				fileID: "file1",
			},
			want: blinkfile.FileHeader{
				ID:       "file1",
				Owner:    "user1",
				Location: filepath.Clean(`_test/repo_file/get_withLocation/file1/file`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.r == nil {
				tt.r = newTestFileRepo(t, "")
			}
			defer cleanDir(t, tt.r.Dir())
			got, err := tt.r.Get(ctx, tt.args.fileID)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileRepo_Delete(t *testing.T) {
	ctx := context.Background()
	type args struct {
		owner       blinkfile.UserID
		deleteFiles []blinkfile.FileID
	}
	tests := []struct {
		name    string
		patch   func(*testing.T) func()
		r       *repo.FileRepo
		args    args
		wantErr error
		assert  func(*testing.T, *repo.FileRepo)
	}{
		{
			name: "should fail if file owner is empty",
			args: args{
				owner: "",
			},
			wantErr: fmt.Errorf("file owner ID cannot be empty"),
		},
		{
			name: "should do nothing if list of files to delete is empty",
			args: args{
				owner:       "user1",
				deleteFiles: nil,
			},
		},
		{
			name: "should fail if any files are not found without deleting any files",
			r: func() *repo.FileRepo {
				r := newTestFileRepo(t, "delete_failWithoutDelete")
				fatalOnErr(t,
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:    "file1",
							Owner: "user1",
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
				)
				return r
			}(),
			args: args{
				owner:       "user1",
				deleteFiles: []blinkfile.FileID{"file1", "this-file-doesn't-exist"},
			},
			wantErr: fmt.Errorf(`file "this-file-doesn't-exist" not found to delete by user "user1"`),
			assert: func(t *testing.T, r *repo.FileRepo) {
				want := blinkfile.FileHeader{
					ID:       "file1",
					Owner:    "user1",
					Location: filepath.Clean(`_test/repo_file/delete_failWithoutDelete/file1/file`),
				}
				got, _ := r.Get(ctx, "file1")
				if !reflect.DeepEqual(got, want) {
					t.Errorf("After Delete(), Get() got: \n\t%+v\nwant: \n\t%+v", got, want)
				}
			},
		},
		{
			name: "should partially fail if deletion fails part way through",
			patch: func(t *testing.T) func() {
				prev := repo.RemoveAll
				var count int
				repo.RemoveAll = func(path string) error {
					if count == 0 {
						count++
						return nil
					}
					return fmt.Errorf("remove err")
				}
				return func() { repo.RemoveAll = prev }
			},
			r: func() *repo.FileRepo {
				r := newTestFileRepo(t, "delete_partialFailure")
				fatalOnErr(t,
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:    "file1",
							Owner: "user1",
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
					r.Save(ctx, blinkfile.File{
						FileHeader: blinkfile.FileHeader{
							ID:    "file2",
							Owner: "user1",
						},
						Data: io.NopCloser(strings.NewReader("file-data")),
					}),
				)
				return r
			}(),
			args: args{
				owner:       "user1",
				deleteFiles: []blinkfile.FileID{"file1", "file2"},
			},
			wantErr: fmt.Errorf(`successfully deleted the first %d file(s) but failed deleting file "file2": %w`, 1, fmt.Errorf("remove err")),
			assert: func(t *testing.T, r *repo.FileRepo) {
				want := []blinkfile.FileHeader{{
					ID:       "file2",
					Owner:    "user1",
					Location: filepath.Clean(`_test/repo_file/delete_partialFailure/file2/file`),
				}}
				got, _ := r.ListByUser(ctx, "user1")
				if !reflect.DeepEqual(got, want) {
					t.Errorf("After Delete(), ListByUser() for user1 got: \n\t%+v\nwant: \n\t%+v", got, want)
				}
			},
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
			err := tt.r.Delete(ctx, tt.args.owner, tt.args.deleteFiles)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.assert != nil {
				tt.assert(t, tt.r)
			}
		})
	}
}
