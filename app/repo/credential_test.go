package repo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/benjohns1/blinkfile"

	"github.com/benjohns1/blinkfile/app"

	"github.com/benjohns1/blinkfile/app/repo"
)

var (
	credentialDirs   map[string]struct{}
	credentialDirNum int
	credentialNumMu  sync.Mutex
)

func newCredentialDir(t *testing.T, dirName string) string {
	t.Helper()
	const testDir = "./_test/repo_credential"
	credentialNumMu.Lock()
	defer credentialNumMu.Unlock()
	if dirName == "" {
		credentialDirNum++
		dirName = fmt.Sprintf("_%d", credentialDirNum)
	}
	if _, ok := credentialDirs[dirName]; ok {
		t.Fatalf("credential dir %q already used for another test", dirName)
	}
	return filepath.Clean(fmt.Sprintf("%s/%s", testDir, dirName))
}

func TestNewCredentialRepo(t *testing.T) {
	type args struct {
		ctx     context.Context
		makeCfg func(t *testing.T, dir string) repo.CredentialRepoConfig
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
				makeCfg: func(_ *testing.T, dir string) repo.CredentialRepoConfig {
					return repo.CredentialRepoConfig{Dir: dir}
				},
			},
			patch: func(t *testing.T) func() {
				prev := repo.MkdirAll
				repo.MkdirAll = func(string, os.FileMode) error {
					return fmt.Errorf("mkdir err")
				}
				return func() { repo.MkdirAll = prev }
			},
			wantErr: fmt.Errorf(`making directory %q: %w`, filepath.Clean("_test/repo_credential/new_test"), fmt.Errorf("mkdir err")),
		},
		{
			name: "should halt building existing indexes if context is cancelled",
			args: args{
				ctx: func() context.Context {
					ctx, cancelFunc := context.WithCancel(context.Background())
					cancelFunc()
					return ctx
				}(),
				makeCfg: func(t *testing.T, dir string) repo.CredentialRepoConfig {
					cfg := repo.CredentialRepoConfig{Dir: dir}
					r, err := repo.NewCredentialRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					fatalOnErr(t,
						r.Set(context.Background(), app.Credentials{
							UserID:       "u1",
							Username:     "u1",
							PasswordHash: "1234",
						}),
					)
					return cfg
				},
			},
			wantErr: fmt.Errorf("context canceled"),
		},
		{
			name: "should still create a credential repo even if reading an existing credential fails, but log an error",
			args: args{
				makeCfg: func(_ *testing.T, dir string) repo.CredentialRepoConfig {
					cfg := repo.CredentialRepoConfig{Dir: dir}
					r, err := repo.NewCredentialRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					fatalOnErr(t, r.Set(context.Background(), app.Credentials{
						UserID:       "u1",
						Username:     "u1",
						PasswordHash: "1234",
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
				fmt.Sprintf("Loading credential data %q: file read err", filepath.Clean(`_test/repo_credential/new_test/u1.json`)),
			},
		},
		{
			name: "should create a new credential repo",
			args: args{
				makeCfg: func(_ *testing.T, dir string) repo.CredentialRepoConfig {
					return repo.CredentialRepoConfig{Dir: dir}
				},
			},
		},
		{
			name: "should create a new credential repo with existing credentials",
			args: args{
				makeCfg: func(_ *testing.T, dir string) repo.CredentialRepoConfig {
					cfg := repo.CredentialRepoConfig{Dir: dir}
					r, err := repo.NewCredentialRepo(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					fatalOnErr(t, r.Set(context.Background(), app.Credentials{
						UserID:       "u1",
						Username:     "u1",
						PasswordHash: "1234",
					}))
					return cfg
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := newCredentialDir(t, "new_test")
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
			_, err := repo.NewCredentialRepo(tt.args.ctx, cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("NewCredentialRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrLogs != nil && !reflect.DeepEqual(spy.errors, tt.wantErrLogs) {
				t.Errorf("NewCredentialRepo() error logs:\n\t%v\nwantErrLogs:\n\t%v", spy.errors, tt.wantErrLogs)
			}
			if !t.Failed() {
				cleanDir(t, dir)
			}
		})
	}
}

func TestUserRepo_Set(t *testing.T) {
	type args struct {
		cred app.Credentials
	}
	tests := []struct {
		name    string
		r       func(*testing.T, string) *repo.CredentialRepo
		patch   func(*testing.T) func()
		args    args
		wantErr error
		want    *app.Credentials
	}{
		{
			name: "should fail if the user ID is empty",
			args: args{
				cred: app.Credentials{
					UserID: "",
				},
			},
			wantErr: fmt.Errorf("user ID cannot be empty"),
		},
		{
			name: "should fail if the username is empty",
			args: args{
				cred: app.Credentials{
					UserID:   "u1",
					Username: "",
				},
			},
			wantErr: fmt.Errorf("username cannot be empty"),
		},
		{
			name: "should fail if marshaling the credential data fails",
			args: args{
				cred: app.Credentials{
					UserID:   "u1",
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
			wantErr: fmt.Errorf("marshaling credential data: %w", fmt.Errorf("marshal err")),
		},
		{
			name: "should succeed even the same user ID already has another credential by another username",
			r: func(t *testing.T, dir string) *repo.CredentialRepo {
				cfg := repo.CredentialRepoConfig{Dir: dir}
				r, err := repo.NewCredentialRepo(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Set(context.Background(), app.Credentials{
					UserID:   "u1",
					Username: "username1",
				}))
				return r
			},
			args: args{
				cred: app.Credentials{
					UserID:   "u1",
					Username: "username2",
				},
			},
			want: &app.Credentials{
				UserID:   "u1",
				Username: "username2",
			},
		},
		{
			name: "should overwrite an existing username whose credentials have the same user ID",
			r: func(t *testing.T, dir string) *repo.CredentialRepo {
				cfg := repo.CredentialRepoConfig{Dir: dir}
				r, err := repo.NewCredentialRepo(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Set(context.Background(), app.Credentials{
					UserID:       "u1",
					Username:     "username1",
					PasswordHash: "111",
				}))
				return r
			},
			args: args{
				cred: app.Credentials{
					UserID:       "u1",
					Username:     "username1",
					PasswordHash: "222",
				},
			},
			want: &app.Credentials{
				UserID:       "u1",
				Username:     "username1",
				PasswordHash: "222",
			},
		},
		{
			name: "should fail to overwrite if credentials user IDs do not match",
			r: func(t *testing.T, dir string) *repo.CredentialRepo {
				cfg := repo.CredentialRepoConfig{Dir: dir}
				r, err := repo.NewCredentialRepo(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Set(context.Background(), app.Credentials{
					UserID:       "u1",
					Username:     "username1",
					PasswordHash: "111",
				}))
				return r
			},
			args: args{
				cred: app.Credentials{
					UserID:       "u2",
					Username:     "username1",
					PasswordHash: "222",
				},
			},
			wantErr: fmt.Errorf(`%w: "username1"`, app.ErrDuplicateUsername),
			want: &app.Credentials{
				UserID:       "u1",
				Username:     "username1",
				PasswordHash: "111",
			},
		},
		{
			name: "should fail if writing the credential fails",
			args: args{
				cred: app.Credentials{
					UserID:   "u1",
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
			wantErr: fmt.Errorf("writing credential data: %w", fmt.Errorf("file write err")),
		},
		{
			name: "should set a username/password credential",
			args: args{
				cred: app.Credentials{
					UserID:       "u1",
					Username:     "username1",
					PasswordHash: "1234",
				},
			},
			want: &app.Credentials{
				UserID:       "u1",
				Username:     "username1",
				PasswordHash: "1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			dir := newCredentialDir(t, "set_test")
			cleanDir(t, dir)
			defer func() {
				if !t.Failed() {
					cleanDir(t, dir)
				}
			}()
			if tt.patch != nil {
				defer tt.patch(t)()
			}
			var r *repo.CredentialRepo
			if tt.r == nil {
				r = newTestCredentialRepo(t, dir)
			} else {
				r = tt.r(t, dir)
			}
			err := r.Set(ctx, tt.args.cred)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.cred.Username == "" {
				return
			}
			got, err := r.GetByUsername(ctx, tt.args.cred.Username)
			if err != nil {
				if tt.want != nil {
					t.Fatal(err)
				}
				if !errors.Is(err, app.ErrCredentialNotFound) {
					t.Fatal(err)
				}
				return
			}
			if !reflect.DeepEqual(&got, tt.want) {
				t.Errorf("after Set() got credential:\n\t%+v\nwant:\n\t%+v", got, tt.want)
			}
		})
	}
}

func newTestCredentialRepo(t *testing.T, dir string) *repo.CredentialRepo {
	r, err := repo.NewCredentialRepo(context.Background(), repo.CredentialRepoConfig{Dir: dir, Log: &spyLog{}})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestUserRepo_Remove(t *testing.T) {
	ctx := context.Background()
	type args struct {
		userID blinkfile.UserID
	}
	tests := []struct {
		name       string
		patch      func(*testing.T) func()
		r          func(t *testing.T, dir string) *repo.CredentialRepo
		args       args
		wantErr    error
		lookupUser blinkfile.Username
		want       *app.Credentials
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
			args: args{
				userID: "u1",
			},
			wantErr: app.ErrCredentialNotFound,
		},
		{
			name: "should delete a user",
			args: args{
				userID: "u1",
			},
			r: func(t *testing.T, dir string) *repo.CredentialRepo {
				cfg := repo.CredentialRepoConfig{Dir: dir}
				r, err := repo.NewCredentialRepo(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				fatalOnErr(t, r.Set(ctx, app.Credentials{
					UserID:   "u1",
					Username: "user1",
				}))
				return r
			},
			lookupUser: "user1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.patch != nil {
				defer tt.patch(t)()
			}
			dir := newCredentialDir(t, "delete_test")
			cleanDir(t, dir)
			defer func() {
				if !t.Failed() {
					cleanDir(t, dir)
				}
			}()
			var r *repo.CredentialRepo
			if tt.r == nil {
				r = newTestCredentialRepo(t, dir)
			} else {
				r = tt.r(t, dir)
			}
			err := r.Remove(ctx, tt.args.userID)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.lookupUser == "" {
				return
			}
			got, err := r.GetByUsername(ctx, tt.lookupUser)
			if err != nil {
				if tt.want != nil {
					t.Fatal(err)
				}
				if !errors.Is(err, app.ErrCredentialNotFound) {
					t.Fatal(err)
				}
				return
			}
			if !reflect.DeepEqual(&got, tt.want) {
				t.Errorf("after Remove() got credential:\n\t%+v\nwant:\n\t%+v", got, tt.want)
			}
		})
	}
}
