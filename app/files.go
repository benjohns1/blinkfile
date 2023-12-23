package app

import (
	"context"
	"encoding/base64"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"io"
)

func (a *App) UploadFile(ctx context.Context, filename string, owner domain.Username, reader io.ReadCloser) error {
	ownerID, err := a.getUserID(owner)
	if err != nil {
		return Error{ErrBadRequest, err}
	}
	fileID, err := generateFileID()
	if err != nil {
		return Error{ErrInternal, fmt.Errorf("generating file ID: %w", err)}
	}
	file, err := domain.UploadFile(fileID, filename, ownerID, reader, a.cfg.Now)
	if err != nil {
		return Error{ErrBadRequest, err}
	}
	err = a.cfg.FileRepo.Save(ctx, file)
	if err != nil {
		return Error{ErrRepo, err}
	}
	return nil
}

func generateFileID() (domain.FileID, error) {
	b, err := generateRandomBytes(64)
	if err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(b)
	return domain.FileID(id), nil
}
