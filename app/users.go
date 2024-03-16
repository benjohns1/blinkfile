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

var (
	ErrDuplicateUsername  = fmt.Errorf("username already exists")
	ErrUsernameTaken      = fmt.Errorf("username already taken")
	ErrCredentialNotFound = fmt.Errorf("credential not found")
)

func (a *App) CreateUser(ctx context.Context, args CreateUserArgs) error {
	_, found, err := a.getAdminCredentials(args.Username)
	if err != nil {
		return Err(ErrInternal, err)
	}
	if found {
		return ErrUser("Error creating user", fmt.Sprintf("Username %q is reserved and cannot be used.", args.Username), fmt.Errorf("attempted to create a user with same username as the system admin %q", args.Username))
	}
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
	err = a.registerCredentials(ctx, user.ID, user.Username, args.Password)
	if err != nil {
		if deleteErr := a.cfg.UserRepo.Delete(ctx, user.ID); deleteErr != nil {
			a.Errorf(ctx, "deleting user after failure to register credentials: %v", deleteErr)
		}
		return err
	}

	return nil
}

func (a *App) registerCredentials(ctx context.Context, userID blinkfile.UserID, username blinkfile.Username, password string) error {
	cred, err := newPasswordCredentials(userID, username, password, a.cfg.PasswordHasher.Hash)
	if err != nil {
		return ErrUser("Error creating user credentials", fmt.Sprintf("Credential error: %s", err), err)
	}
	err = a.cfg.CredentialRepo.Set(ctx, cred)
	if err != nil {
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

func (a *App) DeleteUsers(ctx context.Context, userIDs []blinkfile.UserID) error {
	for _, userID := range userIDs {
		count, err := a.cfg.SessionRepo.DeleteAllUserSessions(ctx, userID)
		if err != nil {
			return Err(ErrRepo, err)
		}
		a.Printf(ctx, "deleted %d sessions for user ID %s", count, userID)
		files, err := a.cfg.FileRepo.ListByUser(ctx, userID)
		if err != nil {
			if err != nil {
				return Err(ErrRepo, err)
			}
		}
		filesToDelete := make([]blinkfile.FileID, 0, len(files))
		for _, file := range files {
			filesToDelete = append(filesToDelete, file.ID)
		}
		appErr := a.DeleteFiles(ctx, userID, filesToDelete)
		if appErr != nil {
			return appErr
		}
		a.Printf(ctx, "deleted %d files for user ID %s", len(filesToDelete), userID)
		err = a.cfg.CredentialRepo.Remove(ctx, userID)
		if err != nil {
			return Err(ErrRepo, err)
		}
		err = a.cfg.UserRepo.Delete(ctx, userID)
		if err != nil {
			return Err(ErrRepo, err)
		}
		a.Printf(ctx, "deleted user ID %s and their credentials", userID)
	}
	return nil
}

const AdminUserID = "_admin"

func (a *App) registerAdminUser(ctx context.Context, username blinkfile.Username, password string) error {
	if username == "" {
		return nil
	}
	cred, err := newPasswordCredentials(AdminUserID, username, password, a.cfg.PasswordHasher.Hash)
	if err != nil {
		return err
	}
	err = a.registerAdminCredentials(cred)
	if err != nil {
		return err
	}
	a.Printf(ctx, "Registered admin credentials for username %q", username)
	return nil
}

func (a *App) registerAdminCredentials(cred Credentials) error {
	if _, exists := a.adminCredentials[cred.Username]; exists {
		return ErrUsernameTaken
	}
	a.adminCredentials[cred.Username] = cred
	return nil
}
