package repo_test

import (
	"os"
	"testing"
)

func cleanDir(t *testing.T, dir string) {
	t.Helper()
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}
