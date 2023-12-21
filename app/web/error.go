package web

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"git.jfam.app/one-way-file-send/app"
	"net/http"
	"strings"
)

type ErrorID string

const ErrorIDLength = 8

func ParseAppErr(err error) (ErrorID, int, string) {
	var appErr app.Error
	if !errors.As(err, &appErr) {
		appErr = app.Error{Err: err}
	}

	id := GenerateErrorID()
	switch appErr.Type {
	case app.ErrBadRequest:
		return id, http.StatusBadRequest, appErr.Err.Error()
	case app.ErrAuthnFailed:
		return id, http.StatusUnauthorized, "Authentication failed"
	default:
		return id, http.StatusInternalServerError, "Internal error"
	}
}

var GenerateErrorID = func() ErrorID {
	v, err := randomBase64String(ErrorIDLength)
	if err != nil {
		panic(err)
	}
	return ErrorID(v)
}

func randomBase64String(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return strings.ToLower(base64.StdEncoding.EncodeToString(b)), nil
}
