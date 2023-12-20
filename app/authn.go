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
		Username   string
		Token      Token
		Expiration time.Time
	}
)

func (a App) Login(ctx context.Context, creds Credentials) (Token, error) {
	if creds.Username == "" {
		return "", Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: username cannot be empty"),
		}
	}
	if stringsAreEqual(creds.Username, a.cfg.AdminCredentials.Username) {
		token, err := a.cfg.GenerateToken()
		if err != nil {
			return "", Error{ErrInternal, err}
		}
		session := Session{
			Username:   creds.Username,
			Token:      token,
			Expiration: a.cfg.Now().Add(a.cfg.SessionExpiration),
		}
		if saveErr := a.cfg.SessionRepo.Save(ctx, session); saveErr != nil {
			return "", Error{ErrRepo, saveErr}
		}
		return token, nil
	}
	return "", Error{
		Type: ErrAuthnFailed,
		Err:  fmt.Errorf("invalid credentials"),
	}
}

func stringsAreEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
