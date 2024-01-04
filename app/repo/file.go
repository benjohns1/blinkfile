package repo

import (
	"context"
	"fmt"
	"git.jfam.app/blinkfile/app"
	"git.jfam.app/blinkfile/domain"
	"io/fs"
	"path/filepath"
	"sync"
	"time"
)

type (
	FileRepoConfig struct {
		Log
		Dir string
	}

	FileRepo struct {
		mu         sync.RWMutex
		dir        string
		ownerIndex map[domain.UserID]map[domain.FileID]fileHeader
		idIndex    map[domain.FileID]fileHeader
		Log
	}

	fileHeader struct {
		ID           domain.FileID
		Name         string
		Location     string
		Owner        domain.UserID
		Created      time.Time
		Expires      time.Time
		Size         int64
		PasswordHash string
	}

	Log interface {
		Errorf(ctx context.Context, format string, v ...any)
	}
)

func NewFileRepo(ctx context.Context, cfg FileRepoConfig) (*FileRepo, error) {
	dir := filepath.Clean(cfg.Dir)
	err := mkdirValidate(dir)
	if err != nil {
		return nil, err
	}
	r := &FileRepo{
		sync.RWMutex{},
		dir,
		make(map[domain.UserID]map[domain.FileID]fileHeader),
		make(map[domain.FileID]fileHeader),
		cfg.Log,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	err = r.buildIndices(ctx, dir)
	if err != nil {
		return nil, err
	}
	return r, err
}

func (r *FileRepo) buildIndices(ctx context.Context, dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			r.Errorf(ctx, "Loading file from %q: %v", path, err)
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
			r.Errorf(ctx, "Loading file header %q: %v", headerFilename, err)
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
	data, err := ReadFile(path)
	if err != nil {
		return header, err
	}
	return header, Unmarshal(data, &header)
}

func (r *FileRepo) Save(_ context.Context, file domain.File) error {
	if file.Data == nil {
		return fmt.Errorf("file data cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	header := fileHeader(file.FileHeader)
	dir, filename, headerFilename := filenames(r.dir, header.ID)
	header.Location = filename
	err := MkdirAll(dir, ModeDir)
	if err != nil {
		return fmt.Errorf("making directory %q: %w", dir, err)
	}

	data, err := Marshal(header)
	if err != nil {
		return fmt.Errorf("marshaling file header: %w", err)
	}
	err = WriteFile(headerFilename, data, ModePerm)
	if err != nil {
		return fmt.Errorf("writing file header: %w", err)
	}

	target, err := CreateFile(filename)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", filename, err)
	}
	defer func() { _ = target.Close() }()
	_, err = Copy(target, file.Data)
	if err != nil {
		return fmt.Errorf("writing file %q: %w", filename, err)
	}
	r.addToIndices(header)

	return nil
}

func (r *FileRepo) DeleteExpiredBefore(_ context.Context, t time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int
	for _, header := range r.idIndex {
		if header.Expires.IsZero() {
			continue
		}
		if header.Expires.After(t) {
			continue
		}
		err := r.deleteFile(header)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (r *FileRepo) ListByUser(_ context.Context, userID domain.UserID) ([]domain.FileHeader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ownedFiles := r.ownerIndex[userID]
	out := make([]domain.FileHeader, 0, len(ownedFiles))
	for _, header := range ownedFiles {
		out = append(out, domain.FileHeader(header))
	}
	return out, nil
}

func (r *FileRepo) Get(_ context.Context, fileID domain.FileID) (domain.FileHeader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	header, found := r.idIndex[fileID]
	if !found {
		return domain.FileHeader{}, app.ErrFileNotFound
	}
	if header.Location == "" {
		_, header.Location, _ = filenames(r.dir, header.ID)
	}
	return domain.FileHeader(header), nil
}

func (r *FileRepo) Delete(_ context.Context, owner domain.UserID, deleteFiles []domain.FileID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ownedFiles := r.ownerIndex[owner]

	toRemove := make([]fileHeader, 0, len(deleteFiles))
	for _, fileID := range deleteFiles {
		header, exists := ownedFiles[fileID]
		if !exists {
			return fmt.Errorf("file %q not found to delete by user %q", fileID, owner)
		}
		toRemove = append(toRemove, header)
	}

	for _, file := range toRemove {
		err := r.deleteFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *FileRepo) deleteFile(file fileHeader) error {
	dir, _, _ := filenames(r.dir, file.ID)
	if err := RemoveAll(dir); err != nil {
		return err
	}
	r.removeFromIndices(file)
	return nil
}

func filenames(root string, fileID domain.FileID) (dir, file, header string) {
	dir = fmt.Sprintf("%s/%s", root, fileID)
	file = fmt.Sprintf("%s/file", dir)
	header = fmt.Sprintf("%s/header.json", dir)
	return filepath.Clean(dir), filepath.Clean(file), filepath.Clean(header)
}
