package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/benjohns1/blinkfile"
)

type (
	CreateUserArgs struct {
		Username blinkfile.Username
		Password string
	}
)

func (a *App) CreateUser(_ context.Context, args CreateUserArgs) error {
	uID, err := a.cfg.GenerateUserID()
	if err != nil {
		return Err(ErrInternal, fmt.Errorf("generating user ID: %w", err))
	}
	_, err = blinkfile.CreateUser(uID, args.Username, a.cfg.Now)
	if err != nil {
		if errors.Is(err, blinkfile.ErrEmptyUsername) {
			return ErrUser("Error creating user", "Username cannot be empty.", err)
		}
		return Err(ErrInternal, err)
	}
	return nil
}

func (a *App) ListUsers(context.Context) ([]blinkfile.User, error) {
	return nil, nil
}

const AdminUserID = "_admin"

func (a *App) registerAdminUser(ctx context.Context, username blinkfile.Username, password string) error {
	if username == "" {
		return nil
	}
	creds, err := newPasswordCredentials(AdminUserID, username, password, a.cfg.PasswordHasher.Hash)
	if err != nil {
		return err
	}
	err = a.registerUserCredentials(creds)
	if err != nil {
		return err
	}
	a.Printf(ctx, "Registered admin credentials for username %q", username)
	return nil
}

var ErrUsernameTaken = fmt.Errorf("username already taken")

func (a *App) registerUserCredentials(creds Credentials) error {
	if _, exists := a.credentials[creds.username]; exists {
		return ErrUsernameTaken
	}
	a.credentials[creds.username] = creds
	return nil
}
