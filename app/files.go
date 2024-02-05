package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/longduration"
)

func (a *App) ListFiles(ctx context.Context, owner blinkfile.UserID) ([]blinkfile.FileHeader, error) {
	if owner == "" {
		return nil, Err(ErrBadRequest, fmt.Errorf("owner is required"))
	}
	files, err := a.cfg.FileRepo.ListByUser(ctx, owner)
	if err != nil {
		return nil, Err(ErrRepo, fmt.Errorf("retrieving file list: %w", err))
	}
	files = a.filterExpired(files)
	sortFilesByCreatedTimeDesc(files)
	return files, nil
}

func (a *App) filterExpired(files []blinkfile.FileHeader) []blinkfile.FileHeader {
	out := make([]blinkfile.FileHeader, 0, len(files))
	for _, file := range files {
		if !file.Expires.IsZero() && !a.cfg.Now().Before(file.Expires) {
			continue
		}
		out = append(out, file)
	}
	return out
}

func sortFilesByCreatedTimeDesc(files []blinkfile.FileHeader) {
	sort.Slice(files, func(i, j int) bool {
		x, y := files[i], files[j]
		if !x.Created.Equal(y.Created) {
			return x.Created.After(y.Created)
		}
		if x.Name != y.Name {
			return x.Name < y.Name
		}
		return x.ID < y.ID
	})
}

type UploadFileArgs struct {
	Filename      string
	Owner         blinkfile.UserID
	Reader        io.ReadCloser
	Size          int64
	Password      string
	ExpiresIn     longduration.LongDuration
	Expires       time.Time
	DownloadLimit int64
}

func (a *App) UploadFile(ctx context.Context, args UploadFileArgs) error {
	fileID, err := a.cfg.GenerateFileID()
	if err != nil {
		return Err(ErrInternal, fmt.Errorf("generating file ID: %w", err))
	}
	hashFunc := func(password string) (hash string) {
		return a.cfg.PasswordHasher.Hash([]byte(password))
	}
	if args.ExpiresIn != "" {
		if !args.Expires.IsZero() {
			return ErrUser("Error validating file expiration", "Can only set one of the expiration fields at a time.", nil)
		}
		args.Expires, err = args.ExpiresIn.AddTo(a.cfg.Now())
		if err != nil {
			return ErrUser("Error calculating file expiration", "Expires In field is not in a valid format.", err)
		}
	}
	file, err := blinkfile.UploadFile(blinkfile.UploadFileArgs{
		ID:            fileID,
		Name:          args.Filename,
		Owner:         args.Owner,
		Reader:        args.Reader,
		Size:          args.Size,
		Now:           a.cfg.Now,
		Password:      args.Password,
		HashFunc:      hashFunc,
		Expires:       args.Expires,
		DownloadLimit: args.DownloadLimit,
	})
	if err != nil {
		return Err(ErrBadRequest, err)
	}
	err = a.cfg.FileRepo.Save(ctx, file)
	if err != nil {
		return Err(ErrRepo, err)
	}
	return nil
}

func (a *App) mimicErr(ctx context.Context, password string, err error) error {
	if errors.Is(err, ErrFileNotFound) || errors.Is(err, blinkfile.ErrFileExpired) {
		a.Errorf(ctx, fmt.Sprintf("mimicking a valid response for security, but real error was: %s", err))
		if password == "" {
			return Err(ErrAuthzFailed, blinkfile.ErrFilePasswordRequired)
		}
		return Err(ErrAuthzFailed, blinkfile.ErrFilePasswordInvalid)
	}
	return err
}

func (a *App) DownloadFile(ctx context.Context, userID blinkfile.UserID, fileID blinkfile.FileID, password string) (blinkfile.FileHeader, error) {
	if fileID == "" {
		return blinkfile.FileHeader{}, Err(ErrBadRequest, fmt.Errorf("file ID is required"))
	}
	matchFunc := func(hashedPassword string, checkPassword string) (matched bool, err error) {
		return a.cfg.PasswordHasher.Match(hashedPassword, []byte(checkPassword))
	}
	file, err := a.cfg.FileRepo.Get(ctx, fileID)
	if err != nil {
		// Mimic responses for files that don't exist
		err = a.mimicErr(ctx, password, Err(ErrRepo, err))
		return blinkfile.FileHeader{}, err
	}
	err = file.Download(userID, password, matchFunc, a.cfg.Now)
	if err != nil {
		if errors.Is(err, blinkfile.ErrFilePasswordInvalid) {
			err = Err(ErrAuthzFailed, err)
		}
		err = a.mimicErr(ctx, password, err)
		return blinkfile.FileHeader{}, err
	}
	return file, nil
}

func (a *App) DeleteFiles(ctx context.Context, owner blinkfile.UserID, deleteFiles []blinkfile.FileID) error {
	if owner == "" {
		return Err(ErrRepo, fmt.Errorf("owner is required"))
	}
	err := a.cfg.FileRepo.Delete(ctx, owner, deleteFiles)
	if err != nil {
		return Err(ErrRepo, err)
	}
	return nil
}

func (a *App) DeleteExpiredFiles(ctx context.Context) error {
	start := a.cfg.Now()
	count, err := a.cfg.FileRepo.DeleteExpiredBefore(ctx, start)
	if count > 0 {
		a.Log.Printf(ctx, "Deleted %d expired files, took %v", count, time.Since(start))
	}
	if err != nil {
		return Err(ErrRepo, err)
	}
	return nil
}
