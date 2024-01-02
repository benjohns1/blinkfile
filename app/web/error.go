package web

import (
	"context"
	"errors"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/request"
	"net/http"
)

type ErrorView struct {
	LayoutView
	ID       string
	Type     string
	Title    string
	Detail   string
	Instance string
	Status   int
}

func ParseAppErr(ctx context.Context, a App, err error) ErrorView {
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		appErr = app.Err(app.ErrInternal, err)
	}

	a.Errorf(ctx, appErr.Error())

	return parseErrorView(ctx, appErr)
}

var defaultErrors = map[app.ErrorType]ErrorView{
	app.ErrBadRequest: {
		Type:   "/problems/bad-request",
		Title:  "Bad Request",
		Detail: "Sorry, but your request could not be handled and we don't have any more detail about what went wrong. Please try again.",
		Status: http.StatusBadRequest,
	},
	app.ErrAuthnFailed: {
		Type:   "/problems/authn-failed",
		Title:  "Authentication failed",
		Detail: "Your credentials were not correct.",
		Status: http.StatusUnauthorized,
	},
	app.ErrAuthzFailed: {
		Type:   "/problems/authz-failed",
		Title:  "Authorization failed",
		Detail: "You do not have permission.",
		Status: http.StatusForbidden,
	},
}

var defaultUnknownError = ErrorView{
	Type:   "/problems/internal",
	Title:  "Internal error",
	Status: http.StatusInternalServerError,
}

func parseErrorView(ctx context.Context, appErr *app.Error) ErrorView {
	view, ok := defaultErrors[appErr.Type]
	if !ok {
		view = defaultUnknownError
	}
	if appErr.Status != 0 {
		view.Status = appErr.Status
		view.Title = http.StatusText(appErr.Status)
	}

	reqID := request.GetID(ctx)
	view.ID = reqID
	view.Instance = fmt.Sprintf("/problems/instances/%s", reqID)
	if appErr.Title != "" {
		view.Title = appErr.Title
	}
	if appErr.Detail != "" {
		view.Detail = appErr.Detail
	}
	return view
}
