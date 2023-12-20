package api

import (
	"context"
	"git.jfam.app/one-way-file-send/app"
	"net/http"
)

type (
	LoginRequest struct {
		Username string
	}
	LoginResponse struct {
		AuthzToken string
	}
)

func (api *API) login(ctx context.Context, req *http.Request) (successResponse, error) {
	var login LoginRequest
	if parseErr := unmarshalRequestBody(req, &login); parseErr != nil {
		return successResponse{}, parseErr
	}
	token, loginErr := api.App.Login(ctx, app.Credentials{Username: login.Username})
	if loginErr != nil {
		return successResponse{}, loginErr
	}
	resp := LoginResponse{
		AuthzToken: string(token),
	}
	return marshalResponseBody(resp)
}
