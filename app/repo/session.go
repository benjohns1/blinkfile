package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
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
		app.Token              `json:"-"`
		Username               string    `json:"username"`
		LoggedIn               time.Time `json:"logged_in"`
		Expires                time.Time `json:"expires"`
		app.SessionRequestData `json:"data"`
	}
)

func NewSession(cfg SessionConfig) (*Session, error) {
	dir := filepath.Clean(cfg.Dir)
	err := os.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return nil, fmt.Errorf("making directory %q: %w", dir, err)
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, fmt.Errorf("getting directory %q info: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", dir)
	}

	return &Session{dir}, nil
}

func (r *Session) Save(_ context.Context, session app.Session) error {
	if session.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	data, err := json.Marshal(sessionData(session))
	if err != nil {
		return err
	}
	err = os.WriteFile(r.filename(session.Token), data, os.ModePerm)
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
	data, err := os.ReadFile(r.filename(token))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		return app.Session{}, false, err
	}
	var sd sessionData
	err = json.Unmarshal(data, &sd)
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
	err := os.Remove(r.filename(token))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
