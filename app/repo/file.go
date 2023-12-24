package repo

import (
	"context"
	"encoding/json"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"git.jfam.app/one-way-file-send/app"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type (
	FileRepoConfig struct {
		Dir string
	}

	FileRepo struct {
		dir        string
		ownerIndex map[domain.UserID]map[domain.FileID]fileHeader
	}

	fileHeader struct {
		ID      domain.FileID
		Name    string
		Owner   domain.UserID
		Created time.Time
		Size    int64
	}
)

func NewFileRepo(ctx context.Context, cfg FileRepoConfig) (*FileRepo, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	r := &FileRepo{
		dir,
		make(map[domain.UserID]map[domain.FileID]fileHeader),
	}
	err = r.indexByOwner(ctx, dir)
	if err != nil {
		return nil, err
	}
	return r, err
}

func (r *FileRepo) indexByOwner(ctx context.Context, dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return filepath.SkipAll
		}
		if err != nil {
			app.Log.Errorf(ctx, "Loading file from %q: %v", path, err)
			return nil
		}
		if path == dir {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		_, _, headerFilename := filenames(dir, domain.FileID(d.Name()))
		header, err := loadFileHeader(ctx, headerFilename)
		if err != nil {
			app.Log.Errorf(ctx, "Loading file header %q: %v", headerFilename, err)
			return nil
		}
		r.addToOwnerIndex(header)
		return nil
	})
}

func (r *FileRepo) addToOwnerIndex(header fileHeader) {
	if _, ok := r.ownerIndex[header.Owner]; !ok {
		r.ownerIndex[header.Owner] = make(map[domain.FileID]fileHeader, 1)
	}
	r.ownerIndex[header.Owner][header.ID] = header
}

func loadFileHeader(_ context.Context, path string) (header fileHeader, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return header, err
	}
	return header, json.Unmarshal(data, &header)
}

func (r *FileRepo) Save(_ context.Context, file domain.File) error {
	header := fileHeader(file.FileHeader)
	dir, filename, headerFilename := filenames(r.dir, header.ID)
	err := os.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("making directory %q: %w", dir, err)
	}

	data, err := json.Marshal(header)
	if err != nil {
		return err
	}
	err = os.WriteFile(headerFilename, data, os.ModePerm)
	if err != nil {
		return err
	}

	target, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", filename, err)
	}
	defer func() { _ = target.Close() }()
	_, err = io.Copy(target, file.Data)
	if err != nil {
		return fmt.Errorf("writing file %q: %w", filename, err)
	}
	r.addToOwnerIndex(header)

	return nil
}

func (r *FileRepo) ListByUser(_ context.Context, userID domain.UserID) ([]domain.FileHeader, error) {
	ownedFiles := r.ownerIndex[userID]
	out := make([]domain.FileHeader, 0, len(ownedFiles))
	for _, header := range ownedFiles {
		out = append(out, domain.FileHeader(header))
	}
	return out, nil
}

func filenames(root string, fileID domain.FileID) (dir, file, header string) {
	dir = fmt.Sprintf("%s/%s", root, fileID)
	file = fmt.Sprintf("%s/file", dir)
	header = fmt.Sprintf("%s/header.json", dir)
	return dir, file, header
}
