package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

var (
	RemoveFile = os.Remove
	WriteFile  = os.WriteFile
	ReadFile   = os.ReadFile
	CreateFile = os.Create
	MkdirAll   = os.MkdirAll
	RemoveAll  = os.RemoveAll
	Copy       = io.Copy
	Lstat      = os.Lstat
	Unmarshal  = json.Unmarshal
	Marshal    = json.Marshal
)

const (
	ModeDir  = os.ModeDir
)

func mkdirValidate(dir string) error {
	err := MkdirAll(dir, ModeDir | os.ModePerm)
	if err != nil {
		return fmt.Errorf("making directory %q: %w", dir, err)
	}
	info, err := Lstat(dir)
	if err != nil {
		return fmt.Errorf("getting directory %q info: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}
	return nil
}
