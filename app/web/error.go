package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/request"
	"github.com/kataras/iris/v12"
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

func parseIrisErr(ctx iris.Context) *app.Error {
	status := ctx.GetStatusCode()
	text := http.StatusText(status)
	if text == "" {
		text = "Unknown Error"
	}
	irisErr := ctx.GetErr()
	if irisErr == nil {
		irisErr = errors.New(text)
	}
	var errType app.ErrorType
	var detail string
	switch status {
	case http.StatusNotFound:
		errType = app.ErrNotFound
		detail = fmt.Sprintf("Sorry, but the resource at %q could not be found.", ctx.Request().RequestURI)
	default:
		errType = app.ErrUnknown
	}
	appErr := &app.Error{
		Type:   errType,
		Title:  text,
		Detail: detail,
		Err:    irisErr,
		Status: status,
	}
	return appErr
}

func handleErrors(h func(ctx iris.Context, a App) error) func(iris.Context, App) {
	return func(ctx iris.Context, a App) {
		err := h(ctx, a)
		if err != nil {
			showError(ctx, a, err)
		}
	}
}

func showError(ctx iris.Context, a App, err error) {
	ctx.ViewData("content", ParseAppErr(ctx, a, err))
	err = ctx.View("error.html")
	if err == nil {
		return
	}
	_, err = ctx.HTML("<h3>%s</h3>", err)
	if err == nil {
		return
	}
	_, err = ctx.WriteString(err.Error())
	a.Errorf(ctx, err.Error())
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
	app.ErrNotFound: {
		Type:   "/problems/not-found",
		Title:  "Not Found",
		Detail: "Sorry, but the resource could not be found.",
		Status: http.StatusNotFound,
	},
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
