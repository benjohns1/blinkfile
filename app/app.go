package app

import (
	"context"
	"crypto/rand"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"time"
)

type (
	Config struct {
		AdminUsername     string
		AdminPassword     string
		SessionExpiration time.Duration
		SessionRepo
		FileRepo
		GenerateToken func() (Token, error)
		Now           func() time.Time
	}

	SessionRepo interface {
		Save(context.Context, Session) error
		Get(context.Context, Token) (Session, bool, error)
		Delete(context.Context, Token) error
	}

	FileRepo interface {
		Save(context.Context, domain.File) error
		ListByUser(context.Context, domain.UserID) ([]domain.FileHeader, error)
		Get(context.Context, domain.FileID) (domain.File, error)
		Delete(context.Context, domain.UserID, []domain.FileID) error
	}

	App struct {
		cfg         Config
		credentials map[domain.Username]Credentials
	}
)

func New(ctx context.Context, cfg Config) (*App, error) {
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

	a := &App{cfg, make(map[domain.Username]Credentials, 1)}

	err := a.registerAdminUser(ctx, domain.Username(cfg.AdminUsername), cfg.AdminPassword)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
