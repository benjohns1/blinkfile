package domain

import (
	"fmt"
	"io"
	"time"
)

type (
	FileID string

	FileHeader struct {
		ID      FileID
		Name    string
		Owner   UserID
		Created time.Time
		Size    int64
	}

	File struct {
		FileHeader
		Data io.ReadCloser
	}
)

func UploadFile(id FileID, name string, owner UserID, reader io.ReadCloser, size int64, now func() time.Time) (File, error) {
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
		FileHeader: FileHeader{
			ID:      id,
			Name:    name,
			Owner:   owner,
			Created: now(),
			Size:    size,
		},
		Data: reader,
	}, nil
}
