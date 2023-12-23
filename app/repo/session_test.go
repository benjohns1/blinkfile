package repo_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/repo"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

var (
	dirNum int
	mu     sync.Mutex
)

func newDir(t *testing.T) string {
	t.Helper()
	const testDir = "./_test/repo_session"
	mu.Lock()
	defer mu.Unlock()
	dirNum++
	return filepath.Clean(fmt.Sprintf("%s/%d", testDir, dirNum))
}

func cleanDir(t *testing.T, dir string) {
	t.Helper()
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewSession(t *testing.T) {
	type args struct {
		cfg repo.SessionConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "should create a new session repo",
			args: args{
				cfg: repo.SessionConfig{Dir: newDir(t)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanDir(t, tt.args.cfg.Dir)
			_, err := repo.NewSession(tt.args.cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSession_Save(t *testing.T) {
	type args struct {
		session app.Session
	}
	tests := []struct {
		name    string
		cfg     repo.SessionConfig
		args    args
		wantErr error
	}{
		{
			name: "should fail with an empty token",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			args: args{
				session: app.Session{
					Token: "",
				},
			},
			wantErr: fmt.Errorf("token cannot be empty"),
		},
		{
			name: "should save a session",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			args: args{
				session: app.Session{
					Token: "token1",
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanDir(t, tt.cfg.Dir)
			r, err := repo.NewSession(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			err = r.Save(context.Background(), tt.args.session)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_Get(t *testing.T) {
	ctx := context.Background()
	type args struct {
		token app.Token
	}
	tests := []struct {
		name    string
		cfg     repo.SessionConfig
		arrange func(*testing.T, *repo.Session)
		args    args
		want    app.Session
		wantOK  bool
		wantErr error
	}{
		{
			name: "should fail with an empty token",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			args: args{
				token: "",
			},
			wantErr: fmt.Errorf("token cannot be empty"),
		},
		{
			name: "should not get a non-existent token",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			args: args{
				token: "token1",
			},
			wantOK: false,
		},
		{
			name: "should get a token",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			arrange: func(t *testing.T, r *repo.Session) {
				if err := r.Save(ctx, app.Session{Token: "token1"}); err != nil {
					t.Fatal(err)
				}
			},
			args: args{
				token: "token1",
			},
			want:   app.Session{Token: "token1"},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanDir(t, tt.cfg.Dir)
			r, err := repo.NewSession(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			if tt.arrange != nil {
				tt.arrange(t, r)
			}
			got, gotOK, err := r.Get(ctx, tt.args.token)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("Get() gotOK = %v, wantOK %v", gotOK, tt.wantOK)
			}
		})
	}
}

func TestSession_Delete(t *testing.T) {
	ctx := context.Background()
	type args struct {
		token app.Token
	}
	tests := []struct {
		name    string
		cfg     repo.SessionConfig
		arrange func(*testing.T, *repo.Session)
		args    args
		wantErr error
		assert  func(*testing.T, *repo.Session)
	}{
		{
			name: "should fail with an empty token",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			args: args{
				token: "",
			},
			wantErr: fmt.Errorf("token cannot be empty"),
		},
		{
			name: "should no-op a non-existent token",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			args: args{
				token: "token1",
			},
			wantErr: nil,
		},
		{
			name: "should delete a token",
			cfg:  repo.SessionConfig{Dir: newDir(t)},
			arrange: func(t *testing.T, r *repo.Session) {
				if err := r.Save(ctx, app.Session{Token: "token1"}); err != nil {
					t.Fatal(err)
				}
			},
			args: args{
				token: "token1",
			},
			assert: func(t *testing.T, r *repo.Session) {
				_, ok, _ := r.Get(ctx, "token1")
				if ok {
					t.Errorf("Delete() did not delete token")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanDir(t, tt.cfg.Dir)
			r, err := repo.NewSession(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			if tt.arrange != nil {
				tt.arrange(t, r)
			}
			err = r.Delete(ctx, tt.args.token)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.assert != nil {
				tt.assert(t, r)
			}
		})
	}
}
