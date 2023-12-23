package repo

import (
	"fmt"
	"os"
)

func mkdirValidate(dir string) error {
	err := os.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("making directory %q: %w", dir, err)
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return fmt.Errorf("getting directory %q info: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}
	return nil
}
