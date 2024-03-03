package repo_test

import (
	"context"
	"fmt"
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

type spyLog struct {
	errors []string
}

func (l *spyLog) Errorf(_ context.Context, format string, v ...any) {
	l.errors = append(l.errors, fmt.Sprintf(format, v...))
}
