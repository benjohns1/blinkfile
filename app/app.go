package app

import (
	"context"
	"fmt"
	"git.jfam.app/blinkfile/domain"
	"time"
)

type (
	Config struct {
		Log
		AdminUsername     string
		AdminPassword     string
		SessionExpiration time.Duration
		SessionRepo
		FileRepo
		GenerateToken func() (Token, error)
		Now           func() time.Time
		PasswordHasher
	}

	SessionRepo interface {
		Save(context.Context, Session) error
		Get(context.Context, Token) (Session, bool, error)
		Delete(context.Context, Token) error
	}

	FileRepo interface {
		Save(context.Context, domain.File) error
		ListByUser(context.Context, domain.UserID) ([]domain.FileHeader, error)
		DeleteExpiredBefore(context.Context, time.Time) (int, error)
		Get(context.Context, domain.FileID) (domain.FileHeader, error)
		Delete(context.Context, domain.UserID, []domain.FileID) error
	}

	PasswordHasher interface {
		Hash(data []byte) (hash string)
		Match(hash string, data []byte) (matched bool, err error)
	}

	App struct {
		cfg         Config
		credentials map[domain.Username]Credentials
		Log
	}

	Log interface {
		Printf(ctx context.Context, format string, v ...any)
		Errorf(ctx context.Context, format string, v ...any)
	}
)

func New(ctx context.Context, cfg Config) (*App, error) {
	if cfg.Log == nil {
		return nil, fmt.Errorf("log instance is required")
	}
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.GenerateToken == nil {
		cfg.GenerateToken = generateDefaultToken
	}
	if cfg.SessionRepo == nil {
		return nil, fmt.Errorf("session repo is required")
	}
	if cfg.FileRepo == nil {
		return nil, fmt.Errorf("file repo is required")
	}
	if cfg.PasswordHasher == nil {
		return nil, fmt.Errorf("password hasher is required")
	}

	a := &App{cfg, make(map[domain.Username]Credentials, 1), cfg.Log}

	err := a.registerAdminUser(ctx, domain.Username(cfg.AdminUsername), cfg.AdminPassword)
	if err != nil {
		return nil, err
	}

	return a, nil
}
