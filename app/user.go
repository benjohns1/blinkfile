package app

import (
	"context"
	"fmt"
	"git.jfam.app/blinkfile"
)

const AdminUserID = "_admin"

func (a *App) registerAdminUser(ctx context.Context, username blinkfile.Username, password string) error {
	if username == "" {
		return nil
	}
	creds, err := a.NewCredentials(AdminUserID, username, password)
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
