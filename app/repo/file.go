package repo

import (
	"context"
	"encoding/json"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"io"
	"os"
	"path/filepath"
	"time"
)

type (
	FileRepoConfig struct {
		Dir string
	}

	FileRepo struct {
		dir string
	}

	fileData struct {
		ID      domain.FileID `json:"-"`
		Name    string
		Owner   domain.UserID
		File    io.ReadCloser `json:"-"`
		Created time.Time
	}
)

func NewFileRepo(cfg FileRepoConfig) (*FileRepo, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	return &FileRepo{dir}, err
}

func (r *FileRepo) Save(_ context.Context, file domain.File) error {
	fd := fileData(file)
	dir, filename, dataFilename := r.filenames(fd)
	err := os.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("making directory %q: %w", dir, err)
	}

	data, err := json.Marshal(fd)
	if err != nil {
		return err
	}
	err = os.WriteFile(dataFilename, data, os.ModePerm)
	if err != nil {
		return err
	}

	target, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", filename, err)
	}
	defer func() { _ = target.Close() }()
	_, err = io.Copy(target, fd.File)
	if err != nil {
		return fmt.Errorf("writing file %q: %w", filename, err)
	}

	return nil
}

func (r *FileRepo) filenames(fd fileData) (dir, file, data string) {
	dir = fmt.Sprintf("%s/%s", r.dir, fd.ID)
	file = fmt.Sprintf("%s/file", dir)
	data = fmt.Sprintf("%s/data.json", dir)
	return dir, file, data
}
