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
		idIndex    map[domain.FileID]fileHeader
	}

	fileHeader struct {
		ID           domain.FileID
		Name         string
		Owner        domain.UserID
		Created      time.Time
		Size         int64
		PasswordHash string
	}
)

func NewFileRepo(ctx context.Context, cfg FileRepoConfig) (*FileRepo, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	if err != nil {
		return nil, err
	}
	r := &FileRepo{
		dir,
		make(map[domain.UserID]map[domain.FileID]fileHeader),
		make(map[domain.FileID]fileHeader),
	}
	err = r.buildIndices(ctx, dir)
	if err != nil {
		return nil, err
	}
	return r, err
}

func (r *FileRepo) buildIndices(ctx context.Context, dir string) error {
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
		r.addToIndices(header)
		return nil
	})
}

func (r *FileRepo) addToIndices(header fileHeader) {
	if _, ok := r.ownerIndex[header.Owner]; !ok {
		r.ownerIndex[header.Owner] = make(map[domain.FileID]fileHeader, 1)
	}
	r.ownerIndex[header.Owner][header.ID] = header
	r.idIndex[header.ID] = header
}

func (r *FileRepo) removeFromIndices(header fileHeader) {
	delete(r.ownerIndex[header.Owner], header.ID)
	delete(r.idIndex, header.ID)
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
	r.addToIndices(header)

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

func (r *FileRepo) Get(_ context.Context, fileID domain.FileID) (domain.File, error) {
	header, found := r.idIndex[fileID]
	if !found {
		return domain.File{}, app.ErrFileNotFound
	}
	fh := domain.FileHeader(header)
	_, filename, _ := filenames(r.dir, fileID)
	file, err := os.Open(filename)
	if err != nil {
		return domain.File{}, err
	}
	return domain.File{
		FileHeader: fh,
		Data:       file,
	}, nil
}

func (r *FileRepo) Delete(_ context.Context, owner domain.UserID, deleteFiles []domain.FileID) error {
	ownedFiles := r.ownerIndex[owner]

	type deleteFile struct {
		dir    string
		header fileHeader
	}
	toRemove := make([]deleteFile, 0, len(deleteFiles))
	for _, fileID := range deleteFiles {
		header, exists := ownedFiles[fileID]
		if !exists {
			return fmt.Errorf("file %q not found to delete by user %q", fileID, owner)
		}
		dir, _, _ := filenames(r.dir, fileID)
		toRemove = append(toRemove, deleteFile{dir, header})
	}

	for _, file := range toRemove {
		if err := os.RemoveAll(file.dir); err != nil {
			return err
		}
		r.removeFromIndices(file.header)
	}
	return nil
}

func filenames(root string, fileID domain.FileID) (dir, file, header string) {
	dir = fmt.Sprintf("%s/%s", root, fileID)
	file = fmt.Sprintf("%s/file", dir)
	header = fmt.Sprintf("%s/header.json", dir)
	return dir, file, header
}
