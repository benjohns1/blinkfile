package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
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
	files = a.filterDownloaded(files)
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

func (a *App) filterDownloaded(files []blinkfile.FileHeader) []blinkfile.FileHeader {
	out := make([]blinkfile.FileHeader, 0, len(files))
	for _, file := range files {
		if file.DownloadLimit > 0 && file.Downloads >= file.DownloadLimit {
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
	fileChanged(ctx, file.Owner, FileEvent{FileHeader: file.FileHeader, Change: FileUploaded})
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
	err = a.cfg.FileRepo.PutHeader(ctx, file)
	if err != nil {
		return blinkfile.FileHeader{}, Err(ErrRepo, err)
	}
	fileChanged(ctx, file.Owner, FileEvent{FileHeader: file, Change: FileDownloaded})
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
	for _, fileID := range deleteFiles {
		fileChanged(ctx, owner, FileEvent{
			FileHeader: blinkfile.FileHeader{ID: fileID},
			Change:     FileDeleted,
		})
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

var (
	subscriptions = map[blinkfile.UserID]map[uint64]chan FileEvent{}
	subKey        uint64
	subMutex      sync.Mutex
)

func (a *App) SubscribeToFileChanges(userID blinkfile.UserID) (<-chan FileEvent, func()) {
	subMutex.Lock()
	defer subMutex.Unlock()
	if _, ok := subscriptions[userID]; !ok {
		subscriptions[userID] = make(map[uint64]chan FileEvent, 1)
	}
	key := subKey
	subKey++
	sub := make(chan FileEvent)
	subscriptions[userID][key] = sub
	return sub, func() {
		subMutex.Lock()
		defer subMutex.Unlock()
		if _, ok := subscriptions[userID][key]; !ok {
			return
		}
		close(sub)
		delete(subscriptions[userID], key)
		if len(subscriptions[userID]) == 0 {
			delete(subscriptions, userID)
		}
	}
}

type (
	FileEvent struct {
		blinkfile.FileHeader
		Change EventType
	}
	EventType string
)

const (
	FileDownloaded EventType = "downloaded"
	FileUploaded   EventType = "uploaded"
	FileDeleted    EventType = "deleted"
)

func fileChanged(_ context.Context, user blinkfile.UserID, file FileEvent) {
	subMutex.Lock()
	defer subMutex.Unlock()
	subs, ok := subscriptions[user]
	if !ok {
		return
	}
	for _, sub := range subs {
		select {
		case sub <- file:
		default:
		}
	}
}
