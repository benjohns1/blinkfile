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

var ErrDuplicateUsername = fmt.Errorf("username already exists")

func (a *App) CreateUser(ctx context.Context, args CreateUserArgs) error {
	uID, err := a.cfg.GenerateUserID()
	if err != nil {
		return Err(ErrInternal, fmt.Errorf("generating user ID: %w", err))
	}
	user, err := blinkfile.CreateUser(uID, args.Username, a.cfg.Now)
	if err != nil {
		if errors.Is(err, blinkfile.ErrEmptyUsername) {
			return ErrUser("Error creating user", "Username cannot be empty.", err)
		}
		return Err(ErrInternal, err)
	}
	err = a.cfg.UserRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, ErrDuplicateUsername) {
			return ErrUser("Error creating user", fmt.Sprintf("Username %q already exists.", user.Username), err)
		}
		return Err(ErrRepo, err)
	}
	return nil
}

func (a *App) ListUsers(ctx context.Context) ([]blinkfile.User, error) {
	users, err := a.cfg.UserRepo.ListAll(ctx)
	if err != nil {
		return nil, Err(ErrRepo, err)
	}
	return users, nil

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
