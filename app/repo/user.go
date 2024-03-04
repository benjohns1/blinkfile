package repo

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/benjohns1/blinkfile/app"

	"github.com/benjohns1/blinkfile"
)

type (
	UserRepoConfig struct {
		Log
		Dir string
	}

	UserRepo struct {
		mu            sync.RWMutex
		dir           string
		usernameIndex map[blinkfile.Username]userData
		idIndex       map[blinkfile.UserID]userData
		Log
	}

	userData struct {
		ID blinkfile.UserID
		blinkfile.Username
		Created time.Time
	}
)

func NewUserRepo(ctx context.Context, cfg UserRepoConfig) (*UserRepo, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	if err != nil {
		return nil, err
	}
	r := &UserRepo{
		sync.RWMutex{},
		dir,
		make(map[blinkfile.Username]userData),
		make(map[blinkfile.UserID]userData),
		cfg.Log,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	err = r.buildIndices(ctx, dir)
	if err != nil {
		return nil, err
	}
	return r, err
}

func (r *UserRepo) buildIndices(ctx context.Context, dir string) error {
	return filepath.WalkDir(dir, func(path string, f fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			r.Errorf(ctx, "Loading file from %q: %v", path, err)
			return nil
		}
		if path == dir {
			return nil
		}
		if f.IsDir() {
			return nil
		}
		user, err := loadUser(path)
		if err != nil {
			r.Errorf(ctx, "Loading user data %q: %v", path, err)
			return nil
		}
		r.addToIndices(user)
		return nil
	})
}

func loadUser(path string) (user userData, err error) {
	data, err := ReadFile(path)
	if err != nil {
		return user, err
	}
	return user, Unmarshal(data, &user)
}

func (r *UserRepo) addToIndices(user userData) {
	r.idIndex[user.ID] = user
	r.usernameIndex[user.Username] = user
}

func (r *UserRepo) Create(_ context.Context, user blinkfile.User) error {
	if user.ID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	u := userData(user)
	data, err := Marshal(u)
	if err != nil {
		return fmt.Errorf("marshaling user data: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.idIndex[u.ID]; exists {
		return fmt.Errorf("duplicate user ID %q already exists", u.ID)
	}
	if _, exists := r.usernameIndex[u.Username]; exists {
		return fmt.Errorf(`%w: %q`, app.ErrDuplicateUsername, u.Username)
	}
	err = WriteFile(r.filename(u.ID), data, 0644)
	if err != nil {
		return fmt.Errorf("writing user data: %w", err)
	}
	r.addToIndices(u)
	return nil
}

func (r *UserRepo) filename(userID blinkfile.UserID) string {
	return fmt.Sprintf("%s/%s.json", r.dir, userID)
}

func (r *UserRepo) ListAll(_ context.Context) ([]blinkfile.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]blinkfile.User, 0, len(r.idIndex))
	for _, user := range r.idIndex {
		out = append(out, blinkfile.User(user))
	}
	return sortUsers(out), nil
}

func sortUsers(users []blinkfile.User) []blinkfile.User {
	sort.Slice(users, func(i, j int) bool {
		return users[i].Username < users[j].Username
	})
	return users
}

func (r *UserRepo) Delete(_ context.Context, userID blinkfile.UserID) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	filename := r.filename(userID)
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := RemoveFile(filename); err != nil {
		return err
	}
	r.removeFromIndices(userID)
	return nil
}

func (r *UserRepo) removeFromIndices(userID blinkfile.UserID) {
	for _, username := range r.getUsernames(userID) {
		delete(r.usernameIndex, username)
	}
	delete(r.idIndex, userID)
}

func (r *UserRepo) getUsernames(userID blinkfile.UserID) []blinkfile.Username {
	var usernames []blinkfile.Username
	for username, user := range r.usernameIndex {
		if user.ID == userID {
			usernames = append(usernames, username)
		}
	}
	return usernames
}
