package app

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"
)

type (
	Credentials struct {
		Username string
	}

	Token string

	Session struct {
		Username string
		Token    Token
		Expires  time.Time
	}
)

func (a App) Login(ctx context.Context, creds Credentials) (Session, error) {
	if creds.Username == "" {
		return Session{}, Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: username cannot be empty"),
		}
	}
	if stringsAreEqual(creds.Username, a.cfg.AdminCredentials.Username) {
		token, err := a.cfg.GenerateToken()
		if err != nil {
			return Session{}, Error{ErrInternal, err}
		}
		session := Session{
			Username: creds.Username,
			Token:    token,
			Expires:  a.cfg.Now().Add(a.cfg.SessionExpiration),
		}
		if saveErr := a.cfg.SessionRepo.Save(ctx, session); saveErr != nil {
			return Session{}, Error{ErrRepo, saveErr}
		}
		return session, nil
	}
	return Session{}, Error{
		Type: ErrAuthnFailed,
		Err:  fmt.Errorf("invalid credentials"),
	}
}

func stringsAreEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
