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
		Owner        UserID
		Created      time.Time
		Size         int64
		PasswordHash string
	}

	File struct {
		FileHeader
		Data io.ReadCloser
	}

	PasswordHashFunc func(password string) (hash string, err error)

	PasswordMatchFunc func(hashedPassword string, checkPassword string) (matched bool, err error)

	UploadFileArgs struct {
		ID       FileID
		Name     string
		Owner    UserID
		Reader   io.ReadCloser
		Size     int64
		Now      func() time.Time
		Password string
		HashFunc PasswordHashFunc
	}
)

func UploadFile(args UploadFileArgs) (File, error) {
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
	var hash string
	if args.Password != "" {
		var err error
		hash, err = args.HashFunc(args.Password)
		if err != nil {
			return File{}, err
		}
	}
	return File{
		FileHeader: FileHeader{
			ID:           args.ID,
			Name:         args.Name,
			Owner:        args.Owner,
			Created:      args.Now(),
			Size:         args.Size,
			PasswordHash: hash,
		},
		Data: args.Reader,
	}, nil
}

var (
	ErrFilePasswordRequired = fmt.Errorf("file access requires password")
	ErrFilePasswordInvalid  = fmt.Errorf("invalid file password")
)

func (f *File) Download(user UserID, password string, matchFunc PasswordMatchFunc) error {
	if f.PasswordHash == "" {
		return nil
	}
	if f.Owner == user {
		return nil
	}
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
	return nil
}
