package repo

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"io/fs"
	"path/filepath"
	"sort"
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
		ownerIndex map[blinkfile.UserID]map[blinkfile.FileID]fileHeader
		idIndex    map[blinkfile.FileID]fileHeader
		Log
	}

	fileHeader struct {
		ID           blinkfile.FileID
		Name         string
		Location     string
		Owner        blinkfile.UserID
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
		make(map[blinkfile.UserID]map[blinkfile.FileID]fileHeader),
		make(map[blinkfile.FileID]fileHeader),
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

func (r *FileRepo) Dir() string {
	return r.dir
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
		_, _, headerFilename := filenames(dir, blinkfile.FileID(d.Name()))
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
		r.ownerIndex[header.Owner] = make(map[blinkfile.FileID]fileHeader, 1)
	}
	r.ownerIndex[header.Owner][header.ID] = header
	r.idIndex[header.ID] = header
}

func (r *FileRepo) removeFromIndices(header blinkfile.FileHeader) {
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

func (r *FileRepo) Save(_ context.Context, file blinkfile.File) error {
	if file.ID == "" {
		return fmt.Errorf("file ID cannot be empty")
	}
	if file.Data == nil {
		return fmt.Errorf("file data cannot be nil")
	}
	if file.Owner == "" {
		return fmt.Errorf("file owner cannot be empty")
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
	var deleteList []blinkfile.FileHeader
	for _, header := range r.idIndex {
		if header.Expires.IsZero() {
			continue
		}
		if header.Expires.After(t) {
			continue
		}
		deleteList = append(deleteList, blinkfile.FileHeader(header))
	}
	var count int
	for _, header := range sortFiles(deleteList) {
		err := r.deleteFile(header)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func sortFiles(files []blinkfile.FileHeader) []blinkfile.FileHeader {
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
	return files
}

func (r *FileRepo) ListByUser(_ context.Context, userID blinkfile.UserID) ([]blinkfile.FileHeader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ownedFiles := r.ownerIndex[userID]
	out := make([]blinkfile.FileHeader, 0, len(ownedFiles))
	for _, header := range ownedFiles {
		out = append(out, blinkfile.FileHeader(header))
	}
	return sortFiles(out), nil
}

func (r *FileRepo) Get(_ context.Context, fileID blinkfile.FileID) (blinkfile.FileHeader, error) {
	if fileID == "" {
		return blinkfile.FileHeader{}, fmt.Errorf("file ID cannot be empty")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	header, found := r.idIndex[fileID]
	if !found {
		return blinkfile.FileHeader{}, app.ErrFileNotFound
	}
	if header.Location == "" {
		_, header.Location, _ = filenames(r.dir, header.ID)
	}
	return blinkfile.FileHeader(header), nil
}

func (r *FileRepo) Delete(_ context.Context, owner blinkfile.UserID, deleteFiles []blinkfile.FileID) error {
	if owner == "" {
		return fmt.Errorf("file owner ID cannot be empty")
	}
	if len(deleteFiles) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	ownedFiles := r.ownerIndex[owner]

	toRemove := make([]blinkfile.FileHeader, 0, len(deleteFiles))
	for _, fileID := range deleteFiles {
		header, exists := ownedFiles[fileID]
		if !exists {
			return fmt.Errorf("file %q not found to delete by user %q", fileID, owner)
		}
		toRemove = append(toRemove, blinkfile.FileHeader(header))
	}

	var count int
	for _, file := range toRemove {
		err := r.deleteFile(file)
		if err != nil {
			return fmt.Errorf("successfully deleted the first %d file(s) but failed deleting file %q: %w", count, file.ID, err)
		}
		count++
	}
	return nil
}

func (r *FileRepo) deleteFile(file blinkfile.FileHeader) error {
	dir, _, _ := filenames(r.dir, file.ID)
	if err := RemoveAll(dir); err != nil {
		return err
	}
	r.removeFromIndices(file)
	return nil
}

func filenames(root string, fileID blinkfile.FileID) (dir, file, header string) {
	dir = fmt.Sprintf("%s/%s", root, fileID)
	file = fmt.Sprintf("%s/file", dir)
	header = fmt.Sprintf("%s/header.json", dir)
	return filepath.Clean(dir), filepath.Clean(file), filepath.Clean(header)
}
