package web

import (
	"context"
	"fmt"
	"net/http"
)

func WriteResponse(w http.ResponseWriter, status int, b []byte) bool {
	w.WriteHeader(status)
	if _, writeErr := w.Write(b); writeErr != nil {
		LogError(context.Background(), "", fmt.Errorf("attempting to write response: %v, response: %s", writeErr, b))
		_, _ = w.Write(nil)
		return false
	}
	return true
}
