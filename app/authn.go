package app

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/benjohns1/blinkfile"
)

type Credentials struct {
	blinkfile.UserID
	blinkfile.Username
	PasswordHash string
}

const PasswordMinLength = 16

var ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters long", PasswordMinLength)

func newPasswordCredentials(userID blinkfile.UserID, user blinkfile.Username, pass string, hash func([]byte) string) (Credentials, error) {
	if userID == "" {
		return Credentials{}, fmt.Errorf("user ID cannot be empty")
	}
	if user == "" {
		return Credentials{}, fmt.Errorf("username cannot be empty")
	}
	if len(pass) < PasswordMinLength {
		return Credentials{}, ErrPasswordTooShort
	}
	encodedHash := hash([]byte(pass))
	return Credentials{
		UserID:       userID,
		Username:     user,
		PasswordHash: encodedHash,
	}, nil
}

func (a *App) IsAuthenticated(ctx context.Context, token Token) (blinkfile.UserID, bool, error) {
	if FeatureFlagIsOn(ctx, "LogAllAuthnCalls") {
		a.Printf(ctx, "LogAllAuthnCalls flag is ON and the IsAuthenticated() method was called")
	}
	if token == "" {
		return "", false, Err(ErrBadRequest, fmt.Errorf("session token cannot be empty"))
	}
	session, found, err := a.cfg.SessionRepo.Get(ctx, token)
	if err != nil {
		return "", false, Err(ErrRepo, err)
	}
	if !found {
		return "", false, nil
	}
	if !session.isValid(a.cfg.Now) {
		return "", false, nil
	}
	if !a.userIsValid(ctx, session.UserID) {
		return "", false, Err(ErrAuthnFailed, fmt.Errorf("session is valid but user ID %q isn't valid", session.UserID))
	}

	return session.UserID, true, nil
}

func (a *App) Login(ctx context.Context, username blinkfile.Username, password string, requestData SessionRequestData) (Session, error) {
	userID, err := a.authenticate(ctx, username, password)
	if err != nil {
		return Session{}, err
	}
	session, err := a.newSession(ctx, userID, requestData)
	if err != nil {
		return Session{}, err
	}
	a.Log.Printf(ctx, "User ID %q logged in", userID)
	return session, nil
}

func (a *App) Logout(ctx context.Context, token Token) error {
	session, _, _ := a.cfg.SessionRepo.Get(ctx, token)
	userID := session.UserID
	err := a.cfg.SessionRepo.Delete(ctx, token)
	if err != nil {
		return Err(ErrRepo, err)
	}
	a.Log.Printf(ctx, "User ID %q logged out", userID)
	return nil
}

func (a *App) newSession(ctx context.Context, userID blinkfile.UserID, data SessionRequestData) (Session, error) {
	token, err := a.cfg.GenerateToken()
	if err != nil {
		return Session{}, Err(ErrInternal, err)
	}
	session := Session{
		Token:              token,
		UserID:             userID,
		LoggedIn:           a.cfg.Now(),
		Expires:            a.cfg.Now().Add(a.cfg.SessionExpiration),
		SessionRequestData: data,
	}
	err = a.cfg.SessionRepo.Save(ctx, session)
	if err != nil {
		return Session{}, Err(ErrRepo, err)
	}
	return session, nil
}

func (a *App) authenticate(ctx context.Context, username blinkfile.Username, password string) (blinkfile.UserID, error) {
	if username == "" {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: username cannot be empty"))
	}
	if password == "" {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: password cannot be empty"))
	}
	cred, found, err := a.getCredentials(ctx, username)
	if err != nil {
		return "", Err(ErrInternal, fmt.Errorf("error retrieving credentials for %q: %w", username, err))
	}
	if !found {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: no username %q found", username))
	}
	match, err := credentialsMatch(cred, username, password, a.cfg.PasswordHasher.Match)
	if err != nil {
		return "", Err(ErrInternal, fmt.Errorf("error matching credentials: %w", err))
	}
	if !match {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: passwords do not match"))
	}
	return cred.UserID, nil
}

func credentialsMatch(c Credentials, username blinkfile.Username, password string, passwordMatcher func(hash string, data []byte) (matched bool, err error)) (bool, error) {
	if !stringsAreEqual(string(c.Username), string(username)) {
		return false, nil
	}
	return passwordMatcher(c.PasswordHash, []byte(password))
}

func (a *App) getCredentials(ctx context.Context, username blinkfile.Username) (Credentials, bool, error) {
	cred, found, err := a.getAdminCredentials(username)
	if err != nil {
		return cred, false, err
	}
	if found {
		return cred, true, nil
	}
	cred, err = a.cfg.CredentialRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			err = nil
		}
		return cred, false, err
	}
	return cred, true, nil
}

func (a *App) getAdminCredentials(username blinkfile.Username) (Credentials, bool, error) {
	cred, found := a.adminCredentials[username]
	if !found {
		return Credentials{}, false, nil
	}
	return cred, true, nil
}

func (a *App) userIsValid(ctx context.Context, userID blinkfile.UserID) bool {
	for _, cred := range a.adminCredentials {
		if cred.UserID == userID {
			return true
		}
	}
	user, found, err := a.cfg.UserRepo.Get(ctx, userID)
	if err != nil {
		a.cfg.Log.Errorf(ctx, "getting user ID %q from repo: %+v", userID, err)
		return false
	}
	if !found {
		return false
	}
	if user.ID != userID {
		return false
	}

	return true
}

func stringsAreEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
