package app

import (
	"context"
	"time"
)

type (
	Config struct {
		AdminCredentials  Credentials
		SessionExpiration time.Duration
		SessionRepo
		GenerateToken func() (Token, error)
		Now           func() time.Time
	}

	SessionRepo interface {
		Save(context.Context, Session) error
		Get(context.Context, Token) (Session, bool, error)
		Delete(context.Context, Token) error
	}

	App struct {
		cfg Config
	}
)

func New(cfg Config) (*App, error) {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.GenerateToken == nil {
		cfg.GenerateToken = generateDefaultToken
	}

	return &App{cfg}, nil
}
