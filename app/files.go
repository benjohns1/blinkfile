package app

import (
	"context"
	"encoding/base64"
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

func (a *App) UploadFile(ctx context.Context, filename string, owner domain.UserID, reader io.ReadCloser, size int64) error {
	fileID, err := generateFileID()
	if err != nil {
		return Error{ErrInternal, fmt.Errorf("generating file ID: %w", err)}
	}
	file, err := domain.UploadFile(fileID, filename, owner, reader, size, a.cfg.Now)
	if err != nil {
		return Error{ErrBadRequest, err}
	}
	err = a.cfg.FileRepo.Save(ctx, file)
	if err != nil {
		return Error{ErrRepo, err}
	}
	return nil
}

func (a *App) DownloadFile(ctx context.Context, userID domain.UserID, fileID domain.FileID) (domain.File, error) {
	file, err := a.cfg.FileRepo.Get(ctx, fileID, FileFilter{&userID})
	if err != nil {
		return domain.File{}, err
	}
	return file, nil
}

func generateFileID() (domain.FileID, error) {
	b, err := generateRandomBytes(64)
	if err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(b)
	return domain.FileID(id), nil
}
