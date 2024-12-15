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
		UserRepo
		CredentialRepo
		GenerateToken func() (Token, error)
		Clock
		PasswordHasher
		GenerateFileID func() (blinkfile.FileID, error)
		GenerateUserID func() (blinkfile.UserID, error)
	}

	SessionRepo interface {
		Save(context.Context, Session) error
		Get(context.Context, Token) (Session, bool, error)
		Delete(context.Context, Token) error
		DeleteAllUserSessions(context.Context, blinkfile.UserID) (int, error)
	}

	FileRepo interface {
		Save(context.Context, blinkfile.File) error
		ListByUser(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error)
		DeleteExpiredBefore(context.Context, time.Time) (int, error)
		Get(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error)
		Delete(context.Context, blinkfile.UserID, []blinkfile.FileID) error
		PutHeader(context.Context, blinkfile.FileHeader) error
	}

	UserRepo interface {
		Create(context.Context, blinkfile.User) error
		Update(context.Context, blinkfile.User) error
		Get(context.Context, blinkfile.UserID) (blinkfile.User, bool, error)
		ListAll(context.Context) ([]blinkfile.User, error)
		Delete(context.Context, blinkfile.UserID) error
	}

	CredentialRepo interface {
		Set(context.Context, Credentials) error
		UpdatePassword(context.Context, Credentials) error
		UpdateUsername(context.Context, blinkfile.UserID, blinkfile.Username, blinkfile.Username) error
		GetByUsername(context.Context, blinkfile.Username) (Credentials, error)
		Remove(context.Context, blinkfile.UserID) error
	}

	PasswordHasher interface {
		Hash(data []byte) (hash string)
		Match(hash string, data []byte) (matched bool, err error)
	}

	App struct {
		cfg              Config
		adminCredentials map[blinkfile.Username]Credentials
		Log
	}

	Log interface {
		Printf(ctx context.Context, format string, v ...any)
		Errorf(ctx context.Context, format string, v ...any)
	}

	Clock interface {
		Now() time.Time
	}
)

type DefaultClock struct{}

func (c *DefaultClock) Now() time.Time {
	return time.Now().UTC()
}

func New(ctx context.Context, cfg Config) (*App, error) {
	if cfg.Log == nil {
		return nil, fmt.Errorf("log instance is required")
	}
	if cfg.Clock == nil {
		cfg.Clock = &DefaultClock{}
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
	if cfg.GenerateUserID == nil {
		cfg.GenerateUserID = generateUserID
	}

	a := &App{cfg, make(map[blinkfile.Username]Credentials, 1), cfg.Log}

	err := a.registerAdminUser(ctx, blinkfile.Username(cfg.AdminUsername), cfg.AdminPassword)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func generateFileID() (blinkfile.FileID, error) {
	const fileIDLength = 64
	id, err := generateRandomBase64(fileIDLength)
	return blinkfile.FileID(id), err
}

func generateUserID() (blinkfile.UserID, error) {
	const userIDLength = 32
	id, err := generateRandomBase64(userIDLength)
	return blinkfile.UserID(id), err
}

func generateRandomBase64(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(b)
	return id, nil
}
