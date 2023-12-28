package html

import (
	"context"
	"errors"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/request"
	"net/http"
)

func ParseAppErr(ctx context.Context, err error) ErrorView {
	var appErr app.Error
	if !errors.As(err, &appErr) {
		appErr = app.Error{Err: err}
	}

	status, msg := getStatusMsg(appErr)
	return ErrorView{
		ID:      request.GetID(ctx),
		Status:  status,
		Message: msg,
	}
}

func getStatusMsg(appErr app.Error) (int, string) {
	switch appErr.Type {
	case app.ErrBadRequest:
		return http.StatusBadRequest, appErr.Err.Error()
	case app.ErrAuthnFailed:
		return http.StatusUnauthorized, "Authentication failed"
	case app.ErrAuthzFailed:
		return http.StatusUnauthorized, "Authorization failed"
	default:
		return http.StatusInternalServerError, "Internal error"
	}
}
