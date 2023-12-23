package app_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"reflect"
	"testing"
	"time"
)

func newCredentials(t *testing.T, user, pass string) app.Credentials {
	t.Helper()
	creds, err := app.NewCredentials(user, pass)
	if err != nil {
		t.Fatal(err)
	}
	return creds
}

func TestApp_Login(t *testing.T) {
	ctx := context.Background()
	type args struct {
		username string
		password string
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
			name: "should fail authentication if username is empty",
			args: args{
				username: "",
			},
			wantErr: app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf("invalid credentials: username cannot be empty"),
			},
		},
		{
			name: "should fail authentication if password is empty",
			args: args{
				username: "admin",
				password: "",
			},
			wantErr: app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf("invalid credentials: password cannot be empty"),
			},
		},
		{
			name: "should fail authentication if username cannot be found",
			args: args{
				username: "unknown-username",
				password: "super-secret-password",
			},
			wantErr: app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf(`invalid credentials: no username "unknown-username" found`),
			},
		},
		{
			name: "should fail to authenticate the admin user due to mismatched password",
			cfg: app.Config{
				AdminCredentials: newCredentials(t, "admin-username", "super-secret-password"),
			},
			args: args{
				username: "admin-username",
				password: "bad-password",
			},
			wantErr: app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf(`invalid credentials: passwords do not match`),
			},
		},
		{
			name: "should fail with an internal error if token generation fails",
			cfg: app.Config{
				AdminCredentials: newCredentials(t, "admin-username", "super-secret-password"),
				GenerateToken: func() (app.Token, error) {
					return "", fmt.Errorf("token generation error")
				},
			},
			args: args{
				username: "admin-username",
				password: "super-secret-password",
			},
			wantErr: app.Error{
				Type: app.ErrInternal,
				Err:  fmt.Errorf("token generation error"),
			},
		},
		{
			name: "should fail with a repo error if storing the session state fails",
			cfg: app.Config{
				AdminCredentials: newCredentials(t, "admin-username", "super-secret-password"),
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
				username: "admin-username",
				password: "super-secret-password",
			},
			wantErr: app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("session save error"),
			},
		},
		{
			name: "should successfully login the admin user and return a valid session",
			cfg: app.Config{
				AdminCredentials: newCredentials(t, "admin-username", "super-secret-password"),
				GenerateToken: func() (app.Token, error) {
					return "token1", nil
				},
				Now:               func() time.Time { return time.Unix(1, 1) },
				SessionExpiration: 2*time.Second + 2*time.Nanosecond,
			},
			args: args{
				username: "admin-username",
				password: "super-secret-password",
			},
			want: app.Session{
				Token:    "token1",
				Username: "admin-username",
				Expires:  time.Unix(3, 3),
			},
			wantSessionSave: []app.Session{{
				Token:    "token1",
				Username: "admin-username",
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
			got, err := application.Login(ctx, tt.args.username, tt.args.password)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Login() got = %+v, want %+v", got, tt.want)
			}
			if tt.wantSessionSave != nil && !reflect.DeepEqual(spy.SaveCalls, tt.wantSessionSave) {
				t.Errorf("Login() got session save calls = %+v, want %+v", spy.SaveCalls, tt.wantSessionSave)
			}
		})
	}
}

func TestApp_Logout(t *testing.T) {
	ctx := context.Background()
	type args struct {
		token app.Token
	}
	tests := []struct {
		name              string
		cfg               app.Config
		args              args
		wantErr           error
		wantSessionDelete []app.Token
	}{
		{
			name: "should fail with a repo error if deleting the session state fails",
			cfg: app.Config{
				SessionRepo: stubSessionRepo{
					DeleteFunc: func(context.Context, app.Token) error {
						return fmt.Errorf("session delete error")
					},
				},
			},
			args: args{
				token: "token1",
			},
			wantErr: app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("session delete error"),
			},
		},
		{
			name: "should successfully logout and delete the session state",
			args: args{
				token: "token1",
			},
			wantSessionDelete: []app.Token{"token1"},
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
			err = application.Logout(ctx, tt.args.token)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantSessionDelete != nil && !reflect.DeepEqual(spy.DeleteCalls, tt.wantSessionDelete) {
				t.Errorf("Logout() got session delete calls = %+v, want %+v", spy.DeleteCalls, tt.wantSessionDelete)
			}
		})
	}
}

func TestApp_IsAuthenticated(t *testing.T) {
	ctx := context.Background()
	type args struct {
		token app.Token
	}
	tests := []struct {
		name           string
		cfg            app.Config
		args           args
		wantErr        error
		want           bool
		wantSessionGet []app.Token
	}{
		{
			name: "should fail with an empty token",
			args: args{
				token: "",
			},
			wantErr: app.Error{
				Type: app.ErrBadRequest,
				Err:  fmt.Errorf("session token cannot be empty"),
			},
		},
		{
			name: "should fail with a repo error if getting the session state fails",
			cfg: app.Config{
				SessionRepo: stubSessionRepo{
					GetFunc: func(context.Context, app.Token) (app.Session, bool, error) {
						return app.Session{}, false, fmt.Errorf("session get error")
					},
				},
			},
			args: args{
				token: "token1",
			},
			wantErr: app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("session get error"),
			},
		},
		{
			name: "should return false if no session found for the given token",
			cfg: app.Config{
				SessionRepo: stubSessionRepo{
					GetFunc: func(context.Context, app.Token) (app.Session, bool, error) {
						return app.Session{}, false, nil
					},
				},
			},
			args: args{
				token: "token1",
			},
			want: false,
		},
		{
			name: "should return false if session is expired",
			cfg: app.Config{
				SessionRepo: stubSessionRepo{
					GetFunc: func(context.Context, app.Token) (app.Session, bool, error) {
						return app.Session{
							Expires: time.Unix(1, 0),
						}, true, nil
					},
				},
				Now: func() time.Time { return time.Unix(1, 0) },
			},
			args: args{
				token: "token1",
			},
			want: false,
		},
		{
			name: "should return true if session is valid",
			cfg: app.Config{
				SessionRepo: stubSessionRepo{
					GetFunc: func(context.Context, app.Token) (app.Session, bool, error) {
						return app.Session{
							Expires: time.Unix(1, 1),
						}, true, nil
					},
				},
				Now: func() time.Time { return time.Unix(1, 0) },
			},
			args: args{
				token: "token1",
			},
			want: true,
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
			got, err := application.IsAuthenticated(ctx, tt.args.token)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsAuthenticated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tt.want)
				return
			}
			if tt.wantSessionGet != nil && !reflect.DeepEqual(spy.GetCalls, tt.wantSessionGet) {
				t.Errorf("IsAuthenticated() got session get calls = %+v, want %+v", spy.GetCalls, tt.wantSessionGet)
			}
		})
	}
}

type spySessionRepo struct {
	repo        app.SessionRepo
	SaveCalls   []app.Session
	DeleteCalls []app.Token
	GetCalls    []app.Token
}

func (r *spySessionRepo) Save(ctx context.Context, session app.Session) error {
	r.SaveCalls = append(r.SaveCalls, session)
	return r.repo.Save(ctx, session)
}

func (r *spySessionRepo) Delete(ctx context.Context, token app.Token) error {
	r.DeleteCalls = append(r.DeleteCalls, token)
	return r.repo.Delete(ctx, token)
}

func (r *spySessionRepo) Get(ctx context.Context, token app.Token) (app.Session, bool, error) {
	r.GetCalls = append(r.GetCalls, token)
	return r.repo.Get(ctx, token)
}

type stubSessionRepo struct {
	SaveFunc   func(context.Context, app.Session) error
	DeleteFunc func(context.Context, app.Token) error
	GetFunc    func(context.Context, app.Token) (app.Session, bool, error)
}

func (r stubSessionRepo) Save(ctx context.Context, session app.Session) error {
	if r.SaveFunc == nil {
		return nil
	}
	return r.SaveFunc(ctx, session)
}

func (r stubSessionRepo) Delete(ctx context.Context, token app.Token) error {
	if r.DeleteFunc == nil {
		return nil
	}
	return r.DeleteFunc(ctx, token)
}

func (r stubSessionRepo) Get(ctx context.Context, token app.Token) (app.Session, bool, error) {
	if r.GetFunc == nil {
		return app.Session{}, false, nil
	}
	return r.GetFunc(ctx, token)
}
