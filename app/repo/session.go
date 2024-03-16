package repo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
)

type (
	SessionConfig struct {
		Log
		Dir string
	}

	SessionRepo struct {
		mu          sync.Mutex
		dir         string
		userIDIndex map[blinkfile.UserID][]app.Token
		Log
	}

	sessionData struct {
		app.Token `json:"-"`
		blinkfile.UserID
		LoggedIn time.Time
		Expires  time.Time
		app.SessionRequestData
	}
)

func NewSessionRepo(ctx context.Context, cfg SessionConfig) (*SessionRepo, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	if err != nil {
		return nil, err
	}
	r := &SessionRepo{
		sync.Mutex{},
		dir,
		make(map[blinkfile.UserID][]app.Token),
		cfg.Log,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	err = r.buildIndices(ctx, dir)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *SessionRepo) buildIndices(ctx context.Context, dir string) error {
	return filepath.WalkDir(dir, func(path string, f fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			r.Errorf(ctx, "Loading session from %q: %v", path, err)
			return nil
		}
		if path == dir {
			return nil
		}
		if f.IsDir() {
			return nil
		}
		sess, err := loadSession(path)
		if err != nil {
			r.Errorf(ctx, "Loading session data %q: %v", path, err)
			return nil
		}
		r.addToIndices(sess)
		return nil
	})
}

func loadSession(path string) (sess sessionData, err error) {
	data, err := ReadFile(path)
	if err != nil {
		return sess, err
	}
	return sess, Unmarshal(data, &sess)
}

func (r *SessionRepo) addToIndices(sess sessionData) {
	r.userIDIndex[sess.UserID] = append(r.userIDIndex[sess.UserID], sess.Token)
}

func (r *SessionRepo) removeFromIndices(ctx context.Context, token app.Token) {
	sess, err := loadSession(r.filename(token))
	if err != nil {
		r.Errorf(ctx, "getting session to remove for token %q: %v", token, err)
	} else {
		delete(r.userIDIndex, sess.UserID)
	}
}

func (r *SessionRepo) Save(_ context.Context, session app.Session) error {
	if session.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	sd := sessionData(session)
	data, err := Marshal(sd)
	if err != nil {
		return err
	}
	err = WriteFile(r.filename(session.Token), data, 0644)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.addToIndices(sd)
	return nil
}

func (r *SessionRepo) filename(token app.Token) string {
	return fmt.Sprintf("%s/%s.json", r.dir, token)
}

func (r *SessionRepo) Get(_ context.Context, token app.Token) (app.Session, bool, error) {
	if token == "" {
		return app.Session{}, false, fmt.Errorf("token cannot be empty")
	}
	sd, err := loadSession(r.filename(token))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		return app.Session{}, false, err
	}
	session := app.Session(sd)
	session.Token = token
	return session, true, nil
}

func (r *SessionRepo) DeleteAllUserSessions(ctx context.Context, userID blinkfile.UserID) (int, error) {
	if userID == "" {
		return 0, fmt.Errorf("user ID cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int
	for _, token := range r.userIDIndex[userID] {
		err := r.delete(ctx, token)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (r *SessionRepo) Delete(ctx context.Context, token app.Token) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.delete(ctx, token)
}

func (r *SessionRepo) delete(ctx context.Context, token app.Token) error {
	err := RemoveFile(r.filename(token))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	r.removeFromIndices(ctx, token)
	return nil
}
