package repo_test

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/app/repo"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	sessionDirNum int
	sessionNumMu  sync.Mutex
)

func newSessionDir(t *testing.T) string {
	t.Helper()
	const testDir = "./_test/repo_session"
	sessionNumMu.Lock()
	defer sessionNumMu.Unlock()
	sessionDirNum++
	return filepath.Clean(fmt.Sprintf("%s/%d", testDir, sessionDirNum))
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
				cfg: repo.SessionConfig{Dir: newSessionDir(t)},
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
		name      string
		cfg       repo.SessionConfig
		marshal   func(v any) ([]byte, error)
		writeFile func(name string, data []byte, perm os.FileMode) error
		args      args
		wantErr   error
	}{
		{
			name: "should fail with an empty token",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			args: args{
				session: app.Session{
					Token: "",
				},
			},
			wantErr: fmt.Errorf("token cannot be empty"),
		},
		{
			name: "should fail if marshaling data fails",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			marshal: func(any) ([]byte, error) {
				return nil, fmt.Errorf("marshal err")
			},
			args: args{
				session: app.Session{
					Token: "token",
				},
			},
			wantErr: fmt.Errorf("marshal err"),
		},
		{
			name: "should fail if writing data fails",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			writeFile: func(string, []byte, os.FileMode) error {
				return fmt.Errorf("write err")
			},
			args: args{
				session: app.Session{
					Token: "token",
				},
			},
			wantErr: fmt.Errorf("write err"),
		},
		{
			name: "should save a session",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
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
			if tt.marshal != nil {
				prev := repo.Marshal
				repo.Marshal = tt.marshal
				defer func() { repo.Marshal = prev }()
			}
			if tt.writeFile != nil {
				prev := repo.WriteFile
				repo.WriteFile = tt.writeFile
				defer func() { repo.WriteFile = prev }()
			}
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
		name      string
		cfg       repo.SessionConfig
		unmarshal func(data []byte, v any) error
		arrange   func(*testing.T, *repo.Session)
		args      args
		want      app.Session
		wantOK    bool
		wantErr   error
	}{
		{
			name: "should fail with an empty token",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			args: args{
				token: "",
			},
			wantErr: fmt.Errorf("token cannot be empty"),
		},
		{
			name: "should not get a non-existent token",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			args: args{
				token: "token1",
			},
			wantOK: false,
		},
		{
			name: "should fail if unmarshalling data fails",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			unmarshal: func([]byte, any) error {
				return fmt.Errorf("unmarshal err")
			},
			arrange: func(t *testing.T, r *repo.Session) {
				if err := r.Save(ctx, app.Session{Token: "token1"}); err != nil {
					t.Fatal(err)
				}
			},
			args: args{
				token: "token1",
			},
			wantErr: fmt.Errorf("unmarshal err"),
		},
		{
			name: "should get a session",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
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
		{
			name: "should get a fully populated session",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			arrange: func(t *testing.T, r *repo.Session) {
				if err := r.Save(ctx, app.Session{
					Token:    "token1",
					UserID:   "user1",
					LoggedIn: time.Unix(1, 0).UTC(),
					Expires:  time.Unix(2, 0).UTC(),
					SessionRequestData: app.SessionRequestData{
						UserAgent: "ua-agent",
						IP:        "ip-addr",
					},
				}); err != nil {
					t.Fatal(err)
				}
			},
			args: args{
				token: "token1",
			},
			want: app.Session{
				Token:    "token1",
				UserID:   "user1",
				LoggedIn: time.Unix(1, 0).UTC(),
				Expires:  time.Unix(2, 0).UTC(),
				SessionRequestData: app.SessionRequestData{
					UserAgent: "ua-agent",
					IP:        "ip-addr",
				},
			},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unmarshal != nil {
				prev := repo.Unmarshal
				repo.Unmarshal = tt.unmarshal
				defer func() { repo.Unmarshal = prev }()
			}
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
				t.Errorf("Get() got = \n\t%#v, want \n\t%#v", got, tt.want)
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
		name       string
		cfg        repo.SessionConfig
		removeFile func(name string) error
		arrange    func(*testing.T, *repo.Session)
		args       args
		wantErr    error
		assert     func(*testing.T, *repo.Session)
	}{
		{
			name: "should fail with an empty token",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			args: args{
				token: "",
			},
			wantErr: fmt.Errorf("token cannot be empty"),
		},
		{
			name: "should no-op a non-existent token",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			args: args{
				token: "token1",
			},
			wantErr: nil,
		},
		{
			name: "should fail if removing the file fails",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
			removeFile: func(name string) error {
				return fmt.Errorf("err remove")
			},
			args: args{
				token: "token1",
			},
			wantErr: fmt.Errorf("err remove"),
		},
		{
			name: "should delete a token",
			cfg:  repo.SessionConfig{Dir: newSessionDir(t)},
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
			if tt.removeFile != nil {
				prev := repo.RemoveFile
				repo.RemoveFile = tt.removeFile
				defer func() { repo.RemoveFile = prev }()
			}
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
