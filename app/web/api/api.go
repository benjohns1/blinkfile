package api

import (
	"context"
	"encoding/json"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/web"
	"io"
	"net/http"
)

type (
	API struct {
		App
	}

	App interface {
		Login(context.Context, app.Credentials) (app.Session, error)
	}

	ResponseErrorBody struct {
		ID  web.ErrorID
		Err string
	}

	successResponse struct {
		Status int
		Body   []byte
	}

	route struct {
		pattern    string
		method     string
		middleware []middlewareFunc
		handler    func(context.Context, *http.Request) (successResponse, error)
	}

	middlewareFunc func(http.HandlerFunc) http.HandlerFunc
)

var _ App = &app.App{}

func (api *API) GetRoutes() map[string]http.HandlerFunc {
	cfg := []route{
		{
			pattern: "/api/login",
			method:  http.MethodPost,
			handler: api.login,
		},
	}
	return api.compileRoutes(cfg)
}

func (api *API) compileRoutes(cfg []route) map[string]http.HandlerFunc {
	routes := make(map[string]http.HandlerFunc, len(cfg))
	for _, r := range cfg {
		if r.method != "" {
			r.middleware = append(r.middleware, api.httpMethod(r.method))
		}
		handler := api.httpResponder(r.handler)
		for _, m := range r.middleware {
			handler = m(handler)
		}
		routes[r.pattern] = handler
	}
	return routes
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

func (api *API) httpMethod(method string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			if req.Method != method {
				web.WriteResponse(w, http.StatusMethodNotAllowed, nil)
				return
			}
			next(w, req)
		}
	}
}

func (api *API) httpResponder(handler func(ctx context.Context, req *http.Request) (successResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		success, err := handler(ctx, req)
		if err != nil {
			errID, errStatus, errMsg := web.ParseAppErr(err)
			b, marshalErr := json.Marshal(ResponseErrorBody{
				ID:  errID,
				Err: errMsg,
			})
			web.LogError(ctx, errID, err)
			if marshalErr != nil {
				web.LogError(ctx, errID, fmt.Errorf("attempting to encode error response: %w", marshalErr))
				web.WriteResponse(w, http.StatusInternalServerError, nil)
				return
			}
			web.WriteResponse(w, errStatus, b)
			return
		}
		web.WriteResponse(w, success.Status, success.Body)
	}
}
