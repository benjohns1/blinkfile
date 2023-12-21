package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

type stubReader struct {
	readErr error
}

func (r stubReader) Read([]byte) (int, error) {
	return 0, r.readErr
}

func readCloser(b []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(b))
}

func jsonMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type spyLog struct {
	errors []string
}

func (l *spyLog) Errorf(_ context.Context, format string, v ...any) {
	l.errors = append(l.errors, fmt.Sprintf(format, v...))
}
