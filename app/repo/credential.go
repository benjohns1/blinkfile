package repo

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
)

type (
	CredentialRepoConfig struct {
		Log
		Dir string
	}

	CredentialRepo struct {
		mu            sync.RWMutex
		dir           string
		usernameIndex map[blinkfile.Username]credentialData
		Log
	}

	credentialData struct {
		blinkfile.UserID
		blinkfile.Username
		PasswordHash string
	}
)

func NewCredentialRepo(ctx context.Context, cfg CredentialRepoConfig) (*CredentialRepo, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	if err != nil {
		return nil, err
	}
	r := &CredentialRepo{
		sync.RWMutex{},
		dir,
		make(map[blinkfile.Username]credentialData),
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

func (r *CredentialRepo) buildIndices(ctx context.Context, dir string) error {
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
		cred, err := loadCredentials(path)
		if err != nil {
			r.Errorf(ctx, "Loading credential data %q: %v", path, err)
			return nil
		}
		r.addToIndices(cred)
		return nil
	})
}

func loadCredentials(path string) (cred credentialData, err error) {
	data, err := ReadFile(path)
	if err != nil {
		return cred, err
	}
	return cred, Unmarshal(data, &cred)
}

func (r *CredentialRepo) addToIndices(cred credentialData) {
	r.usernameIndex[cred.Username] = cred
}

func (r *CredentialRepo) Set(_ context.Context, cred app.Credentials) error {
	if cred.UserID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if cred.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	cd := credentialData(cred)
	data, err := Marshal(cred)
	if err != nil {
		return fmt.Errorf("marshaling credential data: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if found, exists := r.usernameIndex[cd.Username]; exists && found.UserID != cd.UserID {
		return fmt.Errorf(`%w: %q`, app.ErrDuplicateUsername, cd.Username)
	}
	err = WriteFile(r.filename(cd.Username), data, 0644)
	if err != nil {
		return fmt.Errorf("writing credential data: %w", err)
	}
	r.addToIndices(cd)
	return nil
}

func (r *CredentialRepo) Get(_ context.Context, user blinkfile.Username) (out app.Credentials, err error) {
	if user == "" {
		return out, fmt.Errorf("username cannot be empty")
	}
	found, exists := r.usernameIndex[user]
	if !exists {
		return out, app.ErrCredentialNotFound
	}
	return app.Credentials(found), nil
}

func (r *CredentialRepo) filename(userID blinkfile.Username) string {
	return fmt.Sprintf("%s/%s.json", r.dir, userID)
}

func (r *CredentialRepo) Remove(_ context.Context, userID blinkfile.UserID) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	var usernames []blinkfile.Username
	var filenames []string
	for _, username := range r.getUsernames(userID) {
		usernames = append(usernames, username)
		filenames = append(filenames, r.filename(username))
	}
	if len(usernames) == 0 {
		return app.ErrCredentialNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for i, filename := range filenames {
		if err := RemoveFile(filename); err != nil {
			return err
		}
		r.removeFromIndices(usernames[i])
	}
	return nil
}

func (r *CredentialRepo) removeFromIndices(user blinkfile.Username) {
	delete(r.usernameIndex, user)
}

func (r *CredentialRepo) getUsernames(userID blinkfile.UserID) []blinkfile.Username {
	var usernames []blinkfile.Username
	for username, user := range r.usernameIndex {
		if user.UserID == userID {
			usernames = append(usernames, username)
		}
	}
	return usernames
}
