package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"git.jfam.app/one-way-file-send/domain"
	"io"
	"sort"
	"time"
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

type UploadFileArgs struct {
	Filename  string
	Owner     domain.UserID
	Reader    io.ReadCloser
	Size      int64
	Password  string
	ExpiresIn LongDuration
	Expires   time.Time
}

func (a *App) UploadFile(ctx context.Context, args UploadFileArgs) error {
	fileID, err := generateFileID()
	if err != nil {
		return Error{ErrInternal, fmt.Errorf("generating file ID: %w", err)}
	}
	hashFunc := func(password string) (hash string, err error) {
		return a.cfg.PasswordHasher.Hash([]byte(password))
	}
	if args.ExpiresIn != "" {
		if !args.Expires.IsZero() {
			return Error{ErrBadRequest, fmt.Errorf("cannot set both Expires In and Expires On fields")}
		}
		args.Expires, err = args.ExpiresIn.AddTo(a.cfg.Now())
		if err != nil {
			return Error{ErrBadRequest, err}
		}
	}
	file, err := domain.UploadFile(domain.UploadFileArgs{
		ID:       fileID,
		Name:     args.Filename,
		Owner:    args.Owner,
		Reader:   args.Reader,
		Size:     args.Size,
		Now:      a.cfg.Now,
		Password: args.Password,
		HashFunc: hashFunc,
		Expires:  args.Expires,
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
func (a *App) mimicErr(ctx context.Context, password string, err error) error {
	if errors.Is(err, ErrFileNotFound) || errors.Is(err, domain.ErrFileExpired) {
		a.Errorf(ctx, fmt.Sprintf("mimicking a valid response for security, but real error was: %s", err))
		if password == "" {
			return Error{ErrAuthzFailed, domain.ErrFilePasswordRequired}
		}
		return Error{ErrAuthzFailed, domain.ErrFilePasswordInvalid}
	}
	return err
}

func (a *App) DownloadFile(ctx context.Context, userID domain.UserID, fileID domain.FileID, password string) (domain.File, error) {
	matchFunc := func(hashedPassword string, checkPassword string) (matched bool, err error) {
		return a.cfg.PasswordHasher.Match(hashedPassword, []byte(checkPassword))
	}
	file, err := a.cfg.FileRepo.Get(ctx, fileID)
	if err != nil {
		// Mimic responses for files that don't exist
		err = a.mimicErr(ctx, password, err)
		return domain.File{}, err
	}
	err = file.Download(userID, password, matchFunc, a.cfg.Now)
	if err != nil {
		if errors.Is(err, domain.ErrFilePasswordInvalid) {
			err = Error{ErrAuthzFailed, err}
		}
		err = a.mimicErr(ctx, password, err)
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
