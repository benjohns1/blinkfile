package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"io"
	"sort"
)

func (a *App) ListFiles(ctx context.Context, owner domain.UserID) ([]domain.FileHeader, error) {
	files, err := a.cfg.FileRepo.ListByUser(ctx, owner)
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		x, y := files[i], files[j]
		if x.Name != y.Name {
			return x.Name < y.Name
		}
		if !x.Created.Equal(y.Created) {
			return x.Created.Before(y.Created)
		}
		return x.ID < y.ID
	})
	return files, nil
}

func (a *App) UploadFile(ctx context.Context, filename string, owner domain.UserID, reader io.ReadCloser, size int64, password string) error {
	fileID, err := generateFileID()
	if err != nil {
		return Error{ErrInternal, fmt.Errorf("generating file ID: %w", err)}
	}
	hashFunc := func(password string) (hash string, err error) {
		return a.cfg.PasswordHasher.Hash([]byte(password))
	}
	file, err := domain.UploadFile(domain.UploadFileArgs{
		ID:       fileID,
		Name:     filename,
		Owner:    owner,
		Reader:   reader,
		Size:     size,
		Now:      a.cfg.Now,
		Password: password,
		HashFunc: hashFunc,
	})
	if err != nil {
		return Error{ErrBadRequest, err}
	}
	err = a.cfg.FileRepo.Save(ctx, file)
	if err != nil {
		return Error{ErrRepo, err}
	}
	return nil
}

func (a *App) DownloadFile(ctx context.Context, userID domain.UserID, fileID domain.FileID, password string) (domain.File, error) {
	matchFunc := func(hashedPassword string, checkPassword string) (matched bool, err error) {
		return a.cfg.PasswordHasher.Match(hashedPassword, []byte(checkPassword))
	}
	file, err := a.cfg.FileRepo.Get(ctx, fileID)
	if err != nil {
		// Mimic responses for files that don't exist
		if errors.Is(err, ErrFileNotFound) {
			if password == "" {
				err = domain.ErrFilePasswordRequired
			} else {
				err = Error{ErrAuthzFailed, domain.ErrFilePasswordInvalid}
			}
		}
		return domain.File{}, err
	}
	err = file.Download(userID, password, matchFunc)
	if err != nil {
		if errors.Is(err, domain.ErrFilePasswordInvalid) {
			err = Error{ErrAuthzFailed, err}
		}
		return file, err
	}
	return file, nil
}

func (a *App) DeleteFiles(ctx context.Context, owner domain.UserID, deleteFiles []domain.FileID) error {
	return a.cfg.FileRepo.Delete(ctx, owner, deleteFiles)
}

func generateFileID() (domain.FileID, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(b)
	return domain.FileID(id), nil
}
