package app

import (
	"context"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
)

const AdminUserID = "_admin"

func (a *App) registerAdminUser(ctx context.Context, username domain.Username, password string) error {
	if username == "" {
		return nil
	}
	creds, err := NewCredentials(AdminUserID, username, password)
	if err != nil {
		return err
	}
	err = a.registerUserCredentials(creds)
	if err != nil {
		return err
	}
	Log.Printf(ctx, "Registered admin credentials for %q", username)
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

func (a *App) getUserID(username domain.Username) (domain.UserID, error) {
	creds, found := a.credentials[username]
	if !found {
		return "", fmt.Errorf("user not found with username %q", username)
	}
	return creds.UserID, nil
}
