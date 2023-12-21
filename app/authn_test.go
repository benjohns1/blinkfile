package app_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"reflect"
	"testing"
	"time"
)

func TestApp_Login(t *testing.T) {
	ctx := context.Background()
	type args struct {
		creds app.Credentials
	}
	tests := []struct {
		name            string
		cfg             app.Config
		args            args
		want            app.Session
		wantErr         error
		wantSessionSave []app.Session
	}{
		{
			name: "should fail authentication if admin username is empty",
			args: args{
				creds: app.Credentials{
					Username: "",
				},
			},
			wantErr: app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf("invalid credentials: username cannot be empty"),
			},
		},
		{
			name: "should fail authentication if admin username does not match",
			args: args{
				creds: app.Credentials{
					Username: "incorrect-admin-username",
				},
			},
			wantErr: app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf("invalid credentials"),
			},
		},
		{
			name: "should fail with an internal error if token generation fails",
			cfg: app.Config{
				AdminCredentials: app.Credentials{
					Username: "admin-username",
				},
				GenerateToken: func() (app.Token, error) {
					return "", fmt.Errorf("token generation error")
				},
			},
			args: args{
				creds: app.Credentials{
					Username: "admin-username",
				},
			},
			wantErr: app.Error{
				Type: app.ErrInternal,
				Err:  fmt.Errorf("token generation error"),
			},
		},
		{
			name: "should fail with a repo error if storing the session state fails",
			cfg: app.Config{
				AdminCredentials: app.Credentials{
					Username: "admin-username",
				},
				GenerateToken: func() (app.Token, error) {
					return "token1", nil
				},
				SessionRepo: stubSessionRepo{
					SaveFunc: func(context.Context, app.Session) error {
						return fmt.Errorf("session save error")
					},
				},
			},
			args: args{
				creds: app.Credentials{
					Username: "admin-username",
				},
			},
			wantErr: app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("session save error"),
			},
		},
		{
			name: "should successfully save a new session",
			cfg: app.Config{
				AdminCredentials: app.Credentials{
					Username: "admin-username",
				},
				GenerateToken: func() (app.Token, error) {
					return "token1", nil
				},
				Now:               func() time.Time { return time.Unix(1, 1) },
				SessionExpiration: 2*time.Second + 2*time.Nanosecond,
			},
			args: args{
				creds: app.Credentials{
					Username: "admin-username",
				},
			},
			want: app.Session{
				Username: "admin-username",
				Token:    "token1",
				Expires:  time.Unix(3, 3),
			},
			wantSessionSave: []app.Session{{
				Username: "admin-username",
				Token:    "token1",
				Expires:  time.Unix(3, 3),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := tt.cfg.SessionRepo
			if sessionRepo == nil {
				sessionRepo = stubSessionRepo{}
			}
			spy := &spySessionRepo{repo: sessionRepo}
			tt.cfg.SessionRepo = spy
			application, err := app.New(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			got, err := application.Login(ctx, tt.args.creds)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Login() got = %v, want %v", got, tt.want)
			}
			if tt.wantSessionSave != nil && !reflect.DeepEqual(spy.SaveCalls, tt.wantSessionSave) {
				t.Errorf("Login() got session save calls = %+v, want %+v", spy.SaveCalls, tt.wantSessionSave)
			}
		})
	}
}

type spySessionRepo struct {
	repo      app.SessionRepo
	SaveCalls []app.Session
}

func (r *spySessionRepo) Save(ctx context.Context, session app.Session) error {
	r.SaveCalls = append(r.SaveCalls, session)
	return r.repo.Save(ctx, session)
}

type stubSessionRepo struct {
	SaveFunc func(context.Context, app.Session) error
}

func (r stubSessionRepo) Save(ctx context.Context, session app.Session) error {
	if r.SaveFunc == nil {
		return nil
	}
	return r.SaveFunc(ctx, session)
}
