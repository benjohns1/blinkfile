package repo

import (
	"context"
	"errors"
	"fmt"
	"git.jfam.app/blinkfile"
	"git.jfam.app/blinkfile/app"
	"os"
	"path/filepath"
	"time"
)

type (
	SessionConfig struct {
		Dir string
	}

	Session struct {
		dir string
	}

	sessionData struct {
		app.Token `json:"-"`
		blinkfile.UserID
		LoggedIn time.Time
		Expires  time.Time
		app.SessionRequestData
	}
)

func NewSession(cfg SessionConfig) (*Session, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	return &Session{dir}, err
}

func (r *Session) Save(_ context.Context, session app.Session) error {
	if session.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	data, err := Marshal(sessionData(session))
	if err != nil {
		return err
	}
	err = WriteFile(r.filename(session.Token), data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (r *Session) filename(token app.Token) string {
	return fmt.Sprintf("%s/%s.json", r.dir, token)
}

func (r *Session) Get(_ context.Context, token app.Token) (app.Session, bool, error) {
	if token == "" {
		return app.Session{}, false, fmt.Errorf("token cannot be empty")
	}
	data, err := ReadFile(r.filename(token))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		return app.Session{}, false, err
	}
	var sd sessionData
	err = Unmarshal(data, &sd)
	if err != nil {
		return app.Session{}, false, err
	}
	session := app.Session(sd)
	session.Token = token
	return session, true, nil
}

func (r *Session) Delete(_ context.Context, token app.Token) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	err := RemoveFile(r.filename(token))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
