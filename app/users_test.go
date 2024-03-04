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
		{
			name: "should fail if username is a duplicate",
			cfg: app.Config{
				UserRepo: &StubUserRepo{CreateFunc: func(context.Context, blinkfile.User) error {
					return fmt.Errorf("errWrap: %w", app.ErrDuplicateUsername)
				}},
			},
			args: app.CreateUserArgs{
				Username: "user1",
			},
			wantErr: &app.Error{
				Type:   app.ErrBadRequest,
				Title:  "Error creating user",
				Detail: `Username "user1" already exists.`,
				Err:    fmt.Errorf("errWrap: %w", app.ErrDuplicateUsername),
			},
		},
		{
			name: "should fail if user cannot be created in repo",
			cfg: app.Config{
				UserRepo: &StubUserRepo{CreateFunc: func(context.Context, blinkfile.User) error {
					return fmt.Errorf("user repo create err")
				}},
			},
			args: app.CreateUserArgs{
				Username: "user1",
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("user repo create err"),
			},
		},
		{
			name: "should create a new user",
			args: app.CreateUserArgs{
				Username: "user1",
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

func TestApp_ListUsers(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		cfg     app.Config
		want    []blinkfile.User
		wantErr error
	}{
		{
			name: "should fail if user repo returns an error",
			cfg: app.Config{
				UserRepo: &StubUserRepo{ListAllFunc: func(context.Context) ([]blinkfile.User, error) {
					return nil, fmt.Errorf("user repo list err")
				}},
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("user repo list err"),
			},
		},
		{
			name: "should return a list of users",
			cfg: app.Config{
				UserRepo: &StubUserRepo{ListAllFunc: func(context.Context) ([]blinkfile.User, error) {
					return []blinkfile.User{{ID: "u1"}}, nil
				}},
			},
			want: []blinkfile.User{{ID: "u1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			got, err := application.ListUsers(ctx)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListUsers() got\n\t%+v\nwant:\n\t%+v", got, tt.want)
			}
		})
	}
}

func TestApp_DeleteUsers(t *testing.T) {
	ctx := context.Background()
	type args struct {
		userIDs []blinkfile.UserID
	}
	tests := []struct {
		name    string
		args    args
		cfg     app.Config
		wantErr error
	}{
		{
			name: "should fail if user repo returns an error",
			args: args{
				userIDs: []blinkfile.UserID{"u1"},
			},
			cfg: app.Config{
				UserRepo: &StubUserRepo{DeleteFunc: func(context.Context, blinkfile.UserID) error {
					return fmt.Errorf("user repo delete err")
				}},
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("user repo delete err"),
			},
		},
		{
			name: "should delete a user",
			args: args{
				userIDs: []blinkfile.UserID{"u1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			err := application.DeleteUsers(ctx, tt.args.userIDs)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("DeleteUsers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
