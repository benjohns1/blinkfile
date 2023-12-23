package domain

import (
	"fmt"
	"io"
	"time"
)

type (
	FileID string
	File   struct {
		ID      FileID
		Name    string
		Owner   UserID
		File    io.ReadCloser
		Created time.Time
	}
)

func UploadFile(id FileID, name string, owner UserID, reader io.ReadCloser, now func() time.Time) (File, error) {
	if id == "" {
		return File{}, fmt.Errorf("file ID cannot be empty")
	}
	if name == "" {
		return File{}, fmt.Errorf("file name cannot be empty")
	}
	if owner == "" {
		return File{}, fmt.Errorf("file owner cannot be empty")
	}
	if reader == nil {
		return File{}, fmt.Errorf("file reader cannot be empty")
	}
	if now == nil {
		return File{}, fmt.Errorf("now() service cannot be empty")
	}
	return File{
		ID:      id,
		Name:    name,
		Owner:   owner,
		File:    reader,
		Created: now(),
	}, nil
}
