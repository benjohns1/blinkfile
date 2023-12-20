package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"io"
	"net/http"
)

type (
	API struct {
		App
		GenerateErrorID func() (ErrorID, error)
		Log
	}

	Log interface {
		Error(context.Context, ErrorID, error)
	}

	App interface {
		Login(context.Context, app.Credentials) (app.Token, error)
	}

	ResponseErrorBody struct {
		ID  ErrorID
		Err string
	}
	ErrorID string

	successResponse struct {
		Status int
		Body   []byte
	}
)

func (api *API) GetRoutes() map[string]func(http.ResponseWriter, *http.Request) {
	return map[string]func(http.ResponseWriter, *http.Request){
		"/api/login": api.httpResponder(api.login),
	}
}

func unmarshalRequestBody(rawReq *http.Request, v any) error {
	b, parseErr := io.ReadAll(rawReq.Body)
	if parseErr != nil {
		return app.Error{
			Type: app.ErrInternal,
			Err:  fmt.Errorf("reading request body: %w", parseErr),
		}
	}
	if unmarshalErr := json.Unmarshal(b, v); unmarshalErr != nil {
		return app.Error{
			Type: app.ErrBadRequest,
			Err:  fmt.Errorf("parsing request body: %w", unmarshalErr),
		}
	}
	return nil
}

func marshalResponseBody(v any) (successResponse, error) {
	if v == nil {
		return successResponse{http.StatusNoContent, nil}, nil
	}
	resp, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return successResponse{}, app.Error{
			Type: app.ErrInternal,
			Err:  fmt.Errorf("marshaling response body: %w", marshalErr),
		}
	}
	return successResponse{http.StatusOK, resp}, nil
}

func parseAppErr(appErr app.Error) (int, string) {
	switch appErr.Type {
	case app.ErrBadRequest:
		return http.StatusBadRequest, appErr.Err.Error()
	case app.ErrAuthnFailed:
		return http.StatusUnauthorized, "Authentication failed"
	default:
		return http.StatusInternalServerError, "Internal error"
	}
}

func (api *API) httpResponder(handler func(ctx context.Context, req *http.Request) (successResponse, error)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		success, err := handler(ctx, req)
		if err != nil {
			var appErr app.Error
			if !errors.As(err, &appErr) {
				appErr = app.Error{Err: err}
			}
			errStatus, errMsg := parseAppErr(appErr)

			errID, genErr := api.GenerateErrorID()
			if genErr != nil {
				api.Log.Error(ctx, "", fmt.Errorf("generating error ID: %w", genErr))
			}
			b, marshalErr := json.Marshal(ResponseErrorBody{
				ID:  errID,
				Err: errMsg,
			})
			api.Log.Error(ctx, errID, err)
			if marshalErr != nil {
				api.Log.Error(ctx, errID, fmt.Errorf("attempting to encode error response: %w", marshalErr))
				api.writeResponse(w, http.StatusInternalServerError, nil)
				return
			}
			if !api.writeResponse(w, errStatus, b) {
				return
			}
			return
		}
		if !api.writeResponse(w, success.Status, success.Body) {
			return
		}
	}
}

func (api *API) writeResponse(w http.ResponseWriter, status int, b []byte) bool {
	w.WriteHeader(status)
	_, writeErr := w.Write(b)
	if writeErr == nil {
		return true
	}
	api.Log.Error(context.Background(), "", fmt.Errorf("attempting to write response: %v, response: %s", writeErr, b))
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(nil)
	return false
}
