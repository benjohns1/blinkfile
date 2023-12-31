package domain

import (
	"fmt"
	"io"
	"time"
)

type (
	FileID string

	FileHeader struct {
		ID           FileID
		Name         string
		Location     string
		Owner        UserID
		Created      time.Time
		Expires      time.Time
		Size         int64
		PasswordHash string
	}

	File struct {
		FileHeader
		Data io.ReadCloser
	}

	NowFunc func() time.Time

	PasswordHashFunc func(password string) (hash string)

	PasswordMatchFunc func(hashedPassword string, checkPassword string) (matched bool, err error)

	UploadFileArgs struct {
		ID       FileID
		Name     string
		Owner    UserID
		Reader   io.ReadCloser
		Size     int64
		Now      NowFunc
		Password string
		HashFunc PasswordHashFunc
		Expires  time.Time
	}
)

func UploadFile(args UploadFileArgs) (file File, err error) {
	if args.ID == "" {
		return File{}, fmt.Errorf("file ID cannot be empty")
	}
	if args.Name == "" {
		return File{}, fmt.Errorf("file name cannot be empty")
	}
	if args.Owner == "" {
		return File{}, fmt.Errorf("file owner cannot be empty")
	}
	if args.Reader == nil {
		return File{}, fmt.Errorf("file reader cannot be empty")
	}
	if args.Now == nil {
		return File{}, fmt.Errorf("now() service cannot be empty")
	}
	now := args.Now()
	var hash string
	if args.Password != "" {
		if args.HashFunc == nil {
			return File{}, fmt.Errorf("a password is set, so hashFunc() service cannot be empty")
		}
		hash = args.HashFunc(args.Password)
	}
	var expires time.Time
	if !args.Expires.IsZero() {
		expires = args.Expires
	}
	if !expires.IsZero() && !expires.After(now) {
		return File{}, fmt.Errorf("expiration cannot be set in the past")
	}
	return File{
		FileHeader: FileHeader{
			ID:           args.ID,
			Name:         args.Name,
			Owner:        args.Owner,
			Created:      now,
			Size:         args.Size,
			PasswordHash: hash,
			Expires:      expires,
		},
		Data: args.Reader,
	}, nil
}

var (
	ErrFilePasswordRequired = fmt.Errorf("file access requires password")
	ErrFilePasswordInvalid  = fmt.Errorf("invalid file password")
	ErrFileExpired          = fmt.Errorf("file has expired")
)

func (f *FileHeader) Download(user UserID, password string, matchFunc PasswordMatchFunc, nowFunc NowFunc) error {
	if matchFunc == nil {
		return fmt.Errorf("matchFunc() service cannot be empty")
	}
	if nowFunc == nil {
		return fmt.Errorf("now() service cannot be empty")
	}
	now := nowFunc()
	if !f.Expires.IsZero() && !now.Before(f.Expires) {
		return ErrFileExpired
	}
	if f.Owner != "" && f.Owner == user {
		return nil
	}
	if f.PasswordHash != "" {
		if password == "" {
			return ErrFilePasswordRequired
		}
		match, err := matchFunc(f.PasswordHash, password)
		if err != nil {
			return err
		}
		if !match {
			return ErrFilePasswordInvalid
		}
	}
	return nil
}
