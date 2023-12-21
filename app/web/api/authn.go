package api

import (
	"context"
	"git.jfam.app/one-way-file-send/app"
	"net/http"
	"time"
)

type (
	LoginRequest struct {
		Username string
	}
	LoginResponse struct {
		Token   string
		Expires time.Time
	}
)

func (api *API) login(ctx context.Context, req *http.Request) (successResponse, error) {
	var login LoginRequest
	if parseErr := unmarshalRequestBody(req, &login); parseErr != nil {
		return successResponse{}, parseErr
	}
	session, loginErr := api.App.Login(ctx, app.Credentials{Username: login.Username})
	if loginErr != nil {
		return successResponse{}, loginErr
	}
	resp := LoginResponse{
		Token:   string(session.Token),
		Expires: session.Expires,
	}
	return marshalResponseBody(resp)
}
