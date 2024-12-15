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
		r.setIndices(cred)
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

func (r *CredentialRepo) setIndices(cred credentialData) {
	r.usernameIndex[cred.Username] = cred
}

func (r *CredentialRepo) Set(_ context.Context, cred app.Credentials) error {
	cd, data, err := r.parseCredentials(cred)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if found, exists := r.usernameIndex[cd.Username]; exists && found.UserID != cd.UserID {
		return fmt.Errorf("%w: %q", app.ErrDuplicateUsername, cd.Username)
	}
	return r.writeCredentials(cd, data)
}

func (r *CredentialRepo) UpdatePassword(_ context.Context, cred app.Credentials) error {
	cd, data, err := r.parseCredentials(cred)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	found, exists := r.usernameIndex[cd.Username]
	if !exists {
		return fmt.Errorf("%w: %q", app.ErrCredentialNotFound, cd.Username)
	}
	if found.UserID != cd.UserID {
		return fmt.Errorf("user IDs don't match for username: %w: %q", app.ErrCredentialNotFound, cd.Username)
	}
	return r.writeCredentials(cd, data)
}

func (r *CredentialRepo) writeCredentials(cd credentialData, data []byte) error {
	err := WriteFile(r.filename(cd.Username), data, 0644)
	if err != nil {
		return fmt.Errorf("writing credential data: %w", err)
	}
	r.setIndices(cd)
	return nil
}

func (r *CredentialRepo) parseCredentials(cred app.Credentials) (credentialData, []byte, error) {
	if cred.UserID == "" {
		return credentialData{}, nil, fmt.Errorf("user ID cannot be empty")
	}
	if cred.Username == "" {
		return credentialData{}, nil, fmt.Errorf("username cannot be empty")
	}
	cd := credentialData(cred)
	data, err := Marshal(cd)
	if err != nil {
		return credentialData{}, nil, fmt.Errorf("marshaling credential data: %w", err)
	}
	return cd, data, nil
}

func (r *CredentialRepo) UpdateUsername(ctx context.Context, userID blinkfile.UserID, previousUsername, newUsername blinkfile.Username) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if previousUsername == "" {
		return fmt.Errorf("previous username cannot be empty")
	}
	if newUsername == "" {
		return fmt.Errorf("new username cannot be empty")
	}
	if previousUsername == newUsername {
		return fmt.Errorf("previous and new usernames cannot be the same")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cd, previousExists := r.usernameIndex[previousUsername]
	if !previousExists {
		return app.ErrCredentialNotFound
	}
	if found, newExists := r.usernameIndex[newUsername]; newExists && found.UserID != cd.UserID {
		return fmt.Errorf("%w: %q", app.ErrDuplicateUsername, cd.Username)
	}

	cd.Username = newUsername
	data, err := Marshal(cd)
	if err != nil {
		return fmt.Errorf("marshaling credential data: %w", err)
	}

	err = r.writeCredentials(cd, data)
	if err != nil {
		return err
	}
	if rmPrevErr := r.removeUsername(ctx, previousUsername); rmPrevErr != nil {
		if rmNewErr := r.removeUsername(ctx, newUsername); rmNewErr != nil {
			return fmt.Errorf("removing new username after failing to remove previous username: %w", rmPrevErr)
		}
		return fmt.Errorf("removing previous username: %w", rmPrevErr)
	}
	return nil
}

func (r *CredentialRepo) GetByUsername(_ context.Context, user blinkfile.Username) (out app.Credentials, err error) {
	if user == "" {
		return out, fmt.Errorf("username cannot be empty")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	found, exists := r.usernameIndex[user]
	if !exists {
		return out, app.ErrCredentialNotFound
	}
	return app.Credentials(found), nil
}

func (r *CredentialRepo) filename(userID blinkfile.Username) string {
	return fmt.Sprintf("%s/%s.json", r.dir, userID)
}

func (r *CredentialRepo) Remove(ctx context.Context, userID blinkfile.UserID) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	usernames := r.getUsernames(userID)
	if len(usernames) == 0 {
		return app.ErrCredentialNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, username := range usernames {
		if err := r.removeUsername(ctx, username); err != nil {
			return err
		}
	}
	return nil
}

func (r *CredentialRepo) removeUsername(_ context.Context, username blinkfile.Username) error {
	filename := r.filename(username)
	if err := RemoveFile(filename); err != nil {
		return err
	}
	r.removeFromIndices(username)
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
