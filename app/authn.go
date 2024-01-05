package app

import (
	"context"
	"crypto/subtle"
	"fmt"
	"git.jfam.app/blinkfile"
)

type Credentials struct {
	blinkfile.UserID
	username            blinkfile.Username
	encodedPasswordHash string
}

const passwordMinLength = 16

func (a *App) NewCredentials(userID blinkfile.UserID, user blinkfile.Username, pass string) (Credentials, error) {
	if userID == "" {
		return Credentials{}, fmt.Errorf("user ID cannot be empty")
	}
	if user == "" {
		return Credentials{}, fmt.Errorf("username cannot be empty")
	}
	if len(pass) < passwordMinLength {
		return Credentials{}, fmt.Errorf("password must be at least %d characters long", passwordMinLength)
	}
	encodedHash := a.cfg.PasswordHasher.Hash([]byte(pass))
	return Credentials{
		UserID:              userID,
		username:            user,
		encodedPasswordHash: encodedHash,
	}, nil
}

func (a *App) CredentialsMatch(c Credentials, username blinkfile.Username, password string) (bool, error) {
	if !stringsAreEqual(string(c.username), string(username)) {
		return false, nil
	}
	return a.cfg.PasswordHasher.Match(c.encodedPasswordHash, []byte(password))
}

func (a *App) IsAuthenticated(ctx context.Context, token Token) (blinkfile.UserID, bool, error) {
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
	if !a.userIsValid(session.UserID) {
		return "", false, Err(ErrAuthnFailed, fmt.Errorf("session is valid but user ID %q isn't valid", session.UserID))
	}

	return session.UserID, true, nil
}

func (a *App) Login(ctx context.Context, username blinkfile.Username, password string, requestData SessionRequestData) (Session, error) {
	userID, err := a.authenticate(username, password)
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

func (a *App) authenticate(username blinkfile.Username, password string) (blinkfile.UserID, error) {
	if username == "" {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: username cannot be empty"))
	}
	if password == "" {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: password cannot be empty"))
	}
	credentials, found, err := a.getCredentials(username)
	if err != nil {
		return "", Err(ErrInternal, fmt.Errorf("error retrieving credentials for %q: %w", username, err))
	}
	if !found {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: no username %q found", username))
	}
	match, err := a.CredentialsMatch(credentials, username, password)
	if err != nil {
		return "", Err(ErrInternal, fmt.Errorf("error matching credentials: %w", err))
	}
	if !match {
		return "", Err(ErrAuthnFailed, fmt.Errorf("invalid credentials: passwords do not match"))
	}
	return credentials.UserID, nil
}

func (a *App) getCredentials(username blinkfile.Username) (Credentials, bool, error) {
	creds, found := a.credentials[username]
	if !found {
		return Credentials{}, false, nil
	}
	return creds, true, nil
}

func (a *App) userIsValid(userID blinkfile.UserID) bool {
	for _, creds := range a.credentials {
		if creds.UserID == userID {
			return true
		}
	}
	return false
}

func stringsAreEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
