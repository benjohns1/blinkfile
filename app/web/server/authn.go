package server

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/web"
	"net/http"
)

type LoginView struct {
	Redirect       string
	SuccessMessage string
	ErrorView
}

func (ctrl *Controllers) handleLogin(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	view, err := ctrl.login(ctx, w, req)
	if err != nil {
		errID, errStatus, errMsg := web.ParseAppErr(err)
		web.LogError(ctx, errID, err)
		view.ErrorView = ErrorView{
			ID:      errID,
			Status:  errStatus,
			Message: errMsg,
		}
	}
	renderTemplate(ctx, "login.html", w, view)
}

func (ctrl *Controllers) login(ctx context.Context, w http.ResponseWriter, req *http.Request) (LoginView, error) {
	if req.Method == http.MethodGet {
		return LoginView{}, nil
	}

	if req.Method != http.MethodPost {
		return LoginView{}, app.Error{
			Type: app.ErrBadRequest,
			Err:  fmt.Errorf("invalid HTTP method %q", req.Method),
		}
	}
	if parseErr := req.ParseForm(); parseErr != nil {
		return LoginView{}, app.Error{
			Type: app.ErrBadRequest,
			Err:  parseErr,
		}
	}

	creds := app.Credentials{
		Username: req.Form.Get("username"),
	}
	session, err := ctrl.Login(ctx, creds)
	if err != nil {
		return LoginView{}, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    string(session.Token),
		Expires:  session.Expires,
		HttpOnly: true,
	})

	return LoginView{
		Redirect:       "/landing.html",
		SuccessMessage: "Success! Redirecting...",
	}, nil
}
