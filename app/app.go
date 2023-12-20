package app

import (
	"context"
	"time"
)

type (
	Config struct {
		AdminCredentials  Credentials
		SessionExpiration time.Duration
		SessionRepo       SessionRepo
		GenerateToken     func() (Token, error)
		Now               func() time.Time
	}

	App struct {
		cfg Config
	}

	SessionRepo interface {
		Save(context.Context, Session) error
		//GetByToken(context.Context, Token) (Session, bool, error)
		//Delete(context.Context, Token) error
	}
)

func New(cfg Config) (*App, error) {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	return &App{cfg}, nil
}
