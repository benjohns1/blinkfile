package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"git.jfam.app/one-way-file-send/app/web/api"
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
	errors map[api.ErrorID]string
}

func (l *spyLog) Error(_ context.Context, errID api.ErrorID, err error) {
	if l.errors == nil {
		l.errors = make(map[api.ErrorID]string, 1)
	}
	l.errors[errID] = err.Error()
}
