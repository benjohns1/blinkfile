package app_test

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"reflect"
	"testing"
	"time"
)

func TestApp_Login(t *testing.T) {
	ctx := context.Background()
	type args struct {
		username    blinkfile.Username
		password    string
		requestData app.SessionRequestData
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
			wantErr: &app.Error{
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
			wantErr: &app.Error{
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
			wantErr: &app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf(`invalid credentials: no username "unknown-username" found`),
			},
		},
		{
			name: "should fail to authenticate the admin user due to mismatched password",
			cfg: app.Config{
				AdminUsername: "admin-username",
				AdminPassword: "super-secret-password",
			},
			args: args{
				username: "admin-username",
				password: "bad-password",
			},
			wantErr: &app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf(`invalid credentials: passwords do not match`),
			},
		},
		{
			name: "should fail with an internal error if token generation fails",
			cfg: app.Config{
				AdminUsername: "admin-username",
				AdminPassword: "super-secret-password",
				GenerateToken: func() (app.Token, error) {
					return "", fmt.Errorf("token generation error")
				},
			},
			args: args{
				username: "admin-username",
				password: "super-secret-password",
			},
			wantErr: &app.Error{
				Type: app.ErrInternal,
				Err:  fmt.Errorf("token generation error"),
			},
		},
		{
			name: "should fail with a repo error if storing the session state fails",
			cfg: app.Config{
				AdminUsername: "admin-username",
				AdminPassword: "super-secret-password",
				GenerateToken: func() (app.Token, error) {
					return "token1", nil
				},
				SessionRepo: &StubSessionRepo{
					SaveFunc: func(context.Context, app.Session) error {
						return fmt.Errorf("session save error")
					},
				},
			},
			args: args{
				username: "admin-username",
				password: "super-secret-password",
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("session save error"),
			},
		},
		{
			name: "should successfully login the admin user and return a valid session",
			cfg: app.Config{
				AdminUsername: "admin-username",
				AdminPassword: "super-secret-password",
				GenerateToken: func() (app.Token, error) {
					return "token1", nil
				},
				Now:               func() time.Time { return time.Unix(1, 1) },
				SessionExpiration: 2*time.Second + 2*time.Nanosecond,
			},
			args: args{
				username: "admin-username",
				password: "super-secret-password",
				requestData: app.SessionRequestData{
					UserAgent: "ua-data",
					IP:        "ip-addr",
				},
			},
			want: app.Session{
				Token:    "token1",
				UserID:   app.AdminUserID,
				LoggedIn: time.Unix(1, 1),
				Expires:  time.Unix(3, 3),
				SessionRequestData: app.SessionRequestData{
					UserAgent: "ua-data",
					IP:        "ip-addr",
				},
			},
			wantSessionSave: []app.Session{{
				Token:    "token1",
				UserID:   app.AdminUserID,
				LoggedIn: time.Unix(1, 1),
				Expires:  time.Unix(3, 3),
				SessionRequestData: app.SessionRequestData{
					UserAgent: "ua-data",
					IP:        "ip-addr",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			spy := &SpySessionRepo{repo: cfg.SessionRepo}
			cfg.SessionRepo = spy
			application := NewTestApp(ctx, t, cfg)
			got, err := application.Login(ctx, tt.args.username, tt.args.password, tt.args.requestData)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Login() got = \n\t%+v, want \n\t%+v", got, tt.want)
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
				SessionRepo: &StubSessionRepo{
					DeleteFunc: func(context.Context, app.Token) error {
						return fmt.Errorf("session delete error")
					},
				},
			},
			args: args{
				token: "token1",
			},
			wantErr: &app.Error{
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
			cfg := AppConfigDefaults(tt.cfg)
			spy := &SpySessionRepo{repo: cfg.SessionRepo}
			cfg.SessionRepo = spy
			application := NewTestApp(ctx, t, cfg)
			err := application.Logout(ctx, tt.args.token)
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
		wantUserID     blinkfile.UserID
		want           bool
		wantSessionGet []app.Token
	}{
		{
			name: "should fail with an empty token",
			args: args{
				token: "",
			},
			wantErr: app.Err(app.ErrBadRequest, fmt.Errorf("session token cannot be empty")),
		},
		{
			name: "should fail with a repo error if getting the session state fails",
			cfg: app.Config{
				SessionRepo: &StubSessionRepo{
					GetFunc: func(context.Context, app.Token) (app.Session, bool, error) {
						return app.Session{}, false, fmt.Errorf("session get error")
					},
				},
			},
			args: args{
				token: "token1",
			},
			wantErr: app.Err(app.ErrRepo, fmt.Errorf("session get error")),
		},
		{
			name: "should return false if no session found for the given token",
			cfg: app.Config{
				SessionRepo: &StubSessionRepo{
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
				SessionRepo: &StubSessionRepo{
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
			name: "should fail if user doesn't exist even though they have a valid session",
			cfg: app.Config{
				SessionRepo: &StubSessionRepo{
					GetFunc: func(context.Context, app.Token) (app.Session, bool, error) {
						return app.Session{
							UserID:  "user1",
							Expires: time.Unix(1, 1),
						}, true, nil
					},
				},
				Now: func() time.Time { return time.Unix(1, 0) },
			},
			args: args{
				token: "token1",
			},
			wantErr: app.Err(app.ErrAuthnFailed, fmt.Errorf(`session is valid but user ID "user1" isn't valid`)),
		},
		{
			name: "should return userID if session is valid",
			cfg: app.Config{
				SessionRepo: &StubSessionRepo{
					GetFunc: func(context.Context, app.Token) (app.Session, bool, error) {
						return app.Session{
							UserID:  app.AdminUserID,
							Expires: time.Unix(1, 1),
						}, true, nil
					},
				},
				Now:           func() time.Time { return time.Unix(1, 0) },
				AdminUsername: "my-admin-user",
				AdminPassword: "super-secure-admin-password",
			},
			args: args{
				token: "token1",
			},
			wantUserID: app.AdminUserID,
			want:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			spy := &SpySessionRepo{repo: cfg.SessionRepo}
			cfg.SessionRepo = spy
			application := NewTestApp(ctx, t, cfg)
			gotUserID, got, err := application.IsAuthenticated(ctx, tt.args.token)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsAuthenticated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotUserID, tt.wantUserID) {
				t.Errorf("IsAuthenticated() userID = %v, want %v", gotUserID, tt.wantUserID)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tt.want)
			}
			if tt.wantSessionGet != nil && !reflect.DeepEqual(spy.GetCalls, tt.wantSessionGet) {
				t.Errorf("IsAuthenticated() got session get calls = %+v, want %+v", spy.GetCalls, tt.wantSessionGet)
			}
		})
	}
}
