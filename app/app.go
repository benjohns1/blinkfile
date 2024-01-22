package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/benjohns1/blinkfile"
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
		GenerateFileID func() (blinkfile.FileID, error)
	}

	SessionRepo interface {
		Save(context.Context, Session) error
		Get(context.Context, Token) (Session, bool, error)
		Delete(context.Context, Token) error
	}

	FileRepo interface {
		Save(context.Context, blinkfile.File) error
		ListByUser(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error)
		DeleteExpiredBefore(context.Context, time.Time) (int, error)
		Get(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error)
		Delete(context.Context, blinkfile.UserID, []blinkfile.FileID) error
	}

	PasswordHasher interface {
		Hash(data []byte) (hash string)
		Match(hash string, data []byte) (matched bool, err error)
	}

	App struct {
		cfg         Config
		credentials map[blinkfile.Username]Credentials
		Log
	}

	Log interface {
		Printf(ctx context.Context, format string, v ...any)
		Errorf(ctx context.Context, format string, v ...any)
	}
)

var FeatureFlagIsOn = func(context.Context, string) bool { return false }

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
	if cfg.GenerateFileID == nil {
		cfg.GenerateFileID = generateFileID
	}

	a := &App{cfg, make(map[blinkfile.Username]Credentials, 1), cfg.Log}

	err := a.registerAdminUser(ctx, blinkfile.Username(cfg.AdminUsername), cfg.AdminPassword)
	if err != nil {
		return nil, err
	}

	return a, nil
}

const fileIDLength = 64

func generateFileID() (blinkfile.FileID, error) {
	b := make([]byte, fileIDLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(b)
	return blinkfile.FileID(id), nil
}
