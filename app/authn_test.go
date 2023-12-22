package app_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"reflect"
	"testing"
)

func TestApp_Authenticate(t *testing.T) {
	ctx := context.Background()
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		cfg     app.Config
		args    args
		wantErr error
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
				AdminCredentials: func() app.Credentials {
					creds, err := app.NewCredentials("admin-username", "super-secret-password")
					if err != nil {
						t.Fatal(err)
					}
					return creds
				}(),
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
			name: "should successfully authenticate the admin user",
			cfg: app.Config{
				AdminCredentials: func() app.Credentials {
					creds, err := app.NewCredentials("admin-username", "super-secret-password")
					if err != nil {
						t.Fatal(err)
					}
					return creds
				}(),
			},
			args: args{
				username: "admin-username",
				password: "super-secret-password",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application, err := app.New(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			err = application.Authenticate(ctx, tt.args.username, tt.args.password)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
