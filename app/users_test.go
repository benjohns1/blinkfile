package app_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/benjohns1/blinkfile"

	"github.com/benjohns1/blinkfile/app"
)

func TestApp_CreateUser(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		cfg     app.Config
		args    app.CreateUserArgs
		wantErr error
	}{
		{
			name: "should fail generating an user ID",
			cfg: app.Config{
				GenerateUserID: func() (blinkfile.UserID, error) {
					return "", fmt.Errorf("user ID err")
				},
			},
			wantErr: &app.Error{
				Type: app.ErrInternal,
				Err:  fmt.Errorf("generating user ID: %w", fmt.Errorf("user ID err")),
			},
		},
		{
			name: "should fail if username is empty",
			args: app.CreateUserArgs{
				Username: "",
			},
			wantErr: &app.Error{
				Type:   app.ErrBadRequest,
				Title:  "Error creating user",
				Detail: "Username cannot be empty.",
				Err:    blinkfile.ErrEmptyUsername,
			},
		},
		{
			name: "should fail if user details are invalid due to an internal logic error like an empty generated ID",
			cfg: app.Config{
				GenerateUserID: func() (blinkfile.UserID, error) {
					return "", nil
				},
			},
			args: app.CreateUserArgs{
				Username: "user1",
			},
			wantErr: &app.Error{
				Type: app.ErrInternal,
				Err:  blinkfile.ErrEmptyUserID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			err := application.CreateUser(ctx, tt.args)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
