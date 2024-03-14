package repo_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/benjohns1/blinkfile/app"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app/repo"
)

var (
	userDirs   map[string]struct{}
	userDirNum int
	userNumMu  sync.Mutex
)

func newUserDir(t *testing.T, dirName string) string {
	t.Helper()
	const testDir = "./_test/repo_user"
	userNumMu.Lock()
	defer userNumMu.Unlock()
	if dirName == "" {
		userDirNum++
		dirName = fmt.Sprintf("_%d", userDirNum)
	}
	if _, ok := userDirs[dirName]; ok {
		t.Fatalf("user dir %q already used for another test", dirName)
	}
	return filepath.Clean(fmt.Sprintf("%s/%s", testDir, dirName))
}

func TestNewUserRepo(t *testing.T) {
	type args struct {
		ctx     context.Context
		makeCfg func(t *testing.T, dir string) repo.UserRepoConfig
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
				makeCfg: func(_ *testing.T, dir string) repo.UserRepoConfig {
					return repo.UserRepoConfig{Dir: dir}
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.MkdirAll
				repo.MkdirAll = func(string, os.FileMode) error {
					return fmt.Errorf("mkdir err")
				}
				return func() { repo.MkdirAll = prev }
			},
			wantErr: fmt.Errorf(`making directory %q: %w`, filepath.Clean("_test/repo_user/new_test"), fmt.Errorf("mkdir err")),
		},
		{
			name: "should halt building existing indexes if context is cancelled",
			args: args{
				ctx: func() context.Context {
					ctx, cancelFunc := context.WithCancel(context.Background())
					cancelFunc()
					return ctx
				}(),
				makeCfg: func(t *testing.T, dir string) repo.UserRepoConfig {
					cfg := repo.UserRepoConfig{Dir: dir}
					r, err := repo.NewUserRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					fatalOnErr(t,
						r.Create(context.Background(), blinkfile.User{
							ID:       "u1",
							Username: "u1",
						}),
					)
					return cfg
				},
			},
			wantErr: fmt.Errorf("context canceled"),
		},
		{
			name: "should still create a user repo even if reading an existing user fails, but log an error",
			args: args{
				makeCfg: func(_ *testing.T, dir string) repo.UserRepoConfig {
					cfg := repo.UserRepoConfig{Dir: dir}
					r, err := repo.NewUserRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					fatalOnErr(t, r.Create(context.Background(), blinkfile.User{
						ID:       "u1",
						Username: "u1",
					}))
					return cfg
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.ReadFile
				repo.ReadFile = func(name string) ([]byte, error) {
					return nil, fmt.Errorf("file read err")
				}
				return func() { repo.ReadFile = prev }
			},
			wantErrLogs: []string{
				fmt.Sprintf("Loading user data %q: file read err", filepath.Clean(`_test/repo_user/new_test/u1.json`)),
			},
		},
		{
			name: "should create a new user repo",
			args: args{
				makeCfg: func(_ *testing.T, dir string) repo.UserRepoConfig {
					return repo.UserRepoConfig{Dir: dir}
				},
			},
		},
		{
			name: "should create a new user repo with existing users",
			args: args{
				makeCfg: func(_ *testing.T, dir string) repo.UserRepoConfig {
					cfg := repo.UserRepoConfig{Dir: dir}
					r, err := repo.NewUserRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					fatalOnErr(t, r.Create(context.Background(), blinkfile.User{
						ID:       "u1",
						Username: "username",
						Created:  time.Unix(1, 1),
					}))
					return cfg
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := newUserDir(t, "new_test")
			cleanDir(t, dir)
			if tt.patch != nil {
				defer tt.patch(t)()
			}
			spy := &spyLog{}
			cfg := tt.args.makeCfg(t, dir)
			if cfg.Log == nil {
				cfg.Log = spy
			}
			defer cleanDir(t, cfg.Dir)
			if tt.args.ctx == nil {
				tt.args.ctx = context.Background()
			}
			_, err := repo.NewUserRepo(tt.args.ctx, cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("NewUserRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrLogs != nil && !reflect.DeepEqual(spy.errors, tt.wantErrLogs) {
				t.Errorf("NewUserRepo() error logs = %v, wantErrLogs %v", spy.errors, tt.wantErrLogs)
			}
			if !t.Failed() {
				cleanDir(t, dir)
			}
		})
	}
}

func TestUserRepo_Create(t *testing.T) {
	type args struct {
		user blinkfile.User
	}
	tests := []struct {
		name    string
		r       func(*testing.T, string) *repo.UserRepo
		patch   func(*testing.T) func()
		args    args
		wantErr error
		wantAll []blinkfile.User
	}{
		{
			name: "should fail if the user ID is empty",
			args: args{
				user: blinkfile.User{
					ID: "",
				},
			},
			wantErr: fmt.Errorf("user ID cannot be empty"),
		},
		{
			name: "should fail if the username is empty",
			args: args{
				user: blinkfile.User{
					ID:       "u1",
					Username: "",
				},
			},
			wantErr: fmt.Errorf("username cannot be empty"),
		},
		{
			name: "should fail if marshaling the user data fails",
			args: args{
				user: blinkfile.User{
					ID:       "u1",
					Username: "u1",
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.Marshal
				repo.Marshal = func(any) ([]byte, error) {
					return nil, fmt.Errorf("marshal err")
				}
				return func() { repo.Marshal = prev }
			},
			wantErr: fmt.Errorf("marshaling user data: %w", fmt.Errorf("marshal err")),
		},
		{
			name: "should fail if a duplicate user ID already exists",
			r: func(t *testing.T, dir string) *repo.UserRepo {
				cfg := repo.UserRepoConfig{Dir: dir}
				r, err := repo.NewUserRepo(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Create(context.Background(), blinkfile.User{
					ID:       "u1",
					Username: "username1",
				}))
				return r
			},
			args: args{
				user: blinkfile.User{
					ID:       "u1",
					Username: "username2",
				},
			},
			wantErr: fmt.Errorf(`duplicate user ID "u1" already exists`),
		},
		{
			name: "should fail if a duplicate username already exists",
			r: func(t *testing.T, dir string) *repo.UserRepo {
				cfg := repo.UserRepoConfig{Dir: dir}
				r, err := repo.NewUserRepo(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Create(context.Background(), blinkfile.User{
					ID:       "u1",
					Username: "username1",
				}))
				return r
			},
			args: args{
				user: blinkfile.User{
					ID:       "u2",
					Username: "username1",
				},
			},
			wantErr: fmt.Errorf(`%w: "username1"`, app.ErrDuplicateUsername),
		},
		{
			name: "should fail if writing the user fails",
			args: args{
				user: blinkfile.User{
					ID:       "u1",
					Username: "u1",
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.WriteFile
				repo.WriteFile = func(name string, data []byte, perm os.FileMode) error {
					return fmt.Errorf("file write err")
				}
				return func() { repo.WriteFile = prev }
			},
			wantErr: fmt.Errorf("writing user data: %w", fmt.Errorf("file write err")),
		},
		{
			name: "should create a user",
			args: args{
				user: blinkfile.User{
					ID:       "u1",
					Username: "username1",
					Created:  time.Unix(1, 1),
				},
			},
			wantAll: []blinkfile.User{{
				ID:       "u1",
				Username: "username1",
				Created:  time.Unix(1, 1),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			dir := newUserDir(t, "create_test")
			cleanDir(t, dir)
			defer func() {
				if !t.Failed() {
					cleanDir(t, dir)
				}
			}()
			if tt.patch != nil {
				defer tt.patch(t)()
			}
			var r *repo.UserRepo
			if tt.r == nil {
				r = newTestUserRepo(t, dir)
			} else {
				r = tt.r(t, dir)
			}
			err := r.Create(ctx, tt.args.user)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantAll == nil {
				return
			}
			got, err := r.ListAll(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.wantAll) {
				t.Errorf("Create() resulted in all users:\n\t%+v\nwantAll:\n\t%+v", got, tt.wantAll)
			}
		})
	}
}

func newTestUserRepo(t *testing.T, dir string) *repo.UserRepo {
	r, err := repo.NewUserRepo(context.Background(), repo.UserRepoConfig{Dir: dir, Log: &spyLog{}})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestUserRepo_ListAll(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		r    func(*testing.T, string) *repo.UserRepo
		want []blinkfile.User
	}{
		{
			name: "should return an empty list from an empty repo",
			want: []blinkfile.User{},
		},
		{
			name: "should return a user after creating one",
			r: func(t *testing.T, dir string) *repo.UserRepo {
				cfg := repo.UserRepoConfig{Dir: dir}
				r, err := repo.NewUserRepo(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Create(ctx, blinkfile.User{
					ID:       "u1",
					Username: "username1",
					Created:  time.Unix(1, 1),
				}))
				return r
			},
			want: []blinkfile.User{{
				ID:       "u1",
				Username: "username1",
				Created:  time.Unix(1, 1),
			}},
		},
		{
			name: "should return 2 users sorted by ascending username",
			r: func(t *testing.T, dir string) *repo.UserRepo {
				cfg := repo.UserRepoConfig{Dir: dir}
				r, err := repo.NewUserRepo(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t,
					r.Create(ctx, blinkfile.User{
						ID:       "u1",
						Username: "bbb",
						Created:  time.Unix(1, 1),
					}),
					r.Create(ctx, blinkfile.User{
						ID:       "u2",
						Username: "aaa",
						Created:  time.Unix(1, 1),
					}),
				)
				return r
			},
			want: []blinkfile.User{
				{
					ID:       "u2",
					Username: "aaa",
					Created:  time.Unix(1, 1),
				},
				{
					ID:       "u1",
					Username: "bbb",
					Created:  time.Unix(1, 1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := newUserDir(t, "list_all_test")
			cleanDir(t, dir)
			defer func() {
				if !t.Failed() {
					cleanDir(t, dir)
				}
			}()
			var r *repo.UserRepo
			if tt.r == nil {
				r = newTestUserRepo(t, dir)
			} else {
				r = tt.r(t, dir)
			}
			got, err := r.ListAll(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListAll() got:\n\t%+v\nwant:\n\t%+v", got, tt.want)
			}
		})
	}
}

func TestUserRepo_Delete(t *testing.T) {
	ctx := context.Background()
	type args struct {
		userID blinkfile.UserID
	}
	tests := []struct {
		name    string
		patch   func(*testing.T) func()
		r       func(t *testing.T, dir string) *repo.UserRepo
		args    args
		wantErr error
		wantAll []blinkfile.User
	}{
		{
			name: "should fail if user ID is empty",
			args: args{
				userID: "",
			},
			wantErr: fmt.Errorf("user ID cannot be empty"),
		},
		{
			name: "should fail if user to delete doesn't exist",
			patch: func(t *testing.T) func() {
				prev := repo.RemoveAll
				repo.RemoveFile = func(path string) error {
					return fmt.Errorf("remove err")
				}
				return func() { repo.RemoveFile = prev }
			},
			args: args{
				userID: "u1",
			},
			wantErr: fmt.Errorf("remove err"),
		},
		{
			name: "should delete a user",
			args: args{
				userID: "u1",
			},
			r: func(t *testing.T, dir string) *repo.UserRepo {
				cfg := repo.UserRepoConfig{Dir: dir}
				r, err := repo.NewUserRepo(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Create(ctx, blinkfile.User{
					ID:       "u1",
					Username: "user1",
				}))
				return r
			},
			wantAll: []blinkfile.User{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.patch != nil {
				defer tt.patch(t)()
			}
			dir := newUserDir(t, "delete_test")
			cleanDir(t, dir)
			defer func() {
				if !t.Failed() {
					cleanDir(t, dir)
				}
			}()
			var r *repo.UserRepo
			if tt.r == nil {
				r = newTestUserRepo(t, dir)
			} else {
				r = tt.r(t, dir)
			}
			err := r.Delete(ctx, tt.args.userID)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantAll == nil {
				return
			}
			got, err := r.ListAll(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.wantAll) {
				t.Errorf("Delete() resulted in all users:\n\t%+v\nwantAll:\n\t%+v", got, tt.wantAll)
			}
		})
	}
}

func TestUserRepo_Get(t *testing.T) {
	ctx := context.Background()
	type args struct {
		userID blinkfile.UserID
	}
	tests := []struct {
		name      string
		r         func(t *testing.T, dir string) *repo.UserRepo
		args      args
		wantErr   error
		wantFound bool
		want      blinkfile.User
	}{
		{
			name: "should fail if user ID is empty",
			args: args{
				userID: "",
			},
			wantErr: fmt.Errorf("user ID cannot be empty"),
		},
		{
			name: "should not find a user if user doesn't exist",
			args: args{
				userID: "u1",
			},
			wantFound: false,
		},
		{
			name: "should get a user",
			args: args{
				userID: "u1",
			},
			r: func(t *testing.T, dir string) *repo.UserRepo {
				cfg := repo.UserRepoConfig{Dir: dir}
				r, err := repo.NewUserRepo(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Create(ctx, blinkfile.User{
					ID:       "u1",
					Username: "user1",
					Created:  time.Unix(1, 1).UTC(),
				}))
				return r
			},
			wantFound: true,
			want: blinkfile.User{
				ID:       "u1",
				Username: "user1",
				Created:  time.Unix(1, 1).UTC(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := newUserDir(t, "get_test")
			cleanDir(t, dir)
			defer func() {
				if !t.Failed() {
					cleanDir(t, dir)
				}
			}()
			var r *repo.UserRepo
			if tt.r == nil {
				r = newTestUserRepo(t, dir)
			} else {
				r = tt.r(t, dir)
			}
			got, found, err := r.Get(ctx, tt.args.userID)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if found != tt.wantFound {
				t.Errorf("Get() found: %v, want: %v", found, tt.wantFound)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get():\n\t%+v\nwant:\n\t%+v", got, tt.want)
			}
		})
	}
}
