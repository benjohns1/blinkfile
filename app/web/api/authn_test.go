package api_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/web"
	"git.jfam.app/one-way-file-send/app/web/api"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type stubAppLogin struct {
	LoginSession app.Session
	LoginErr     error
}

func (a *stubAppLogin) Login(context.Context, app.Credentials) (app.Session, error) {
	return a.LoginSession, a.LoginErr
}

type spyAppLogin struct {
	api.App
	stubAppLogin
	app.Credentials
}

func (a *spyAppLogin) Login(ctx context.Context, creds app.Credentials) (app.Session, error) {
	a.Credentials = creds
	return a.stubAppLogin.Login(ctx, creds)
}

func TestAPI_Login(t *testing.T) {
	type args struct {
		pattern string
		req     *http.Request
	}
	tests := []struct {
		name          string
		args          args
		stubApp       stubAppLogin
		wantStatus    int
		wantBody      []byte
		wantSpy       app.Credentials
		wantErrorLogs []string
	}{
		{
			name: "should fail if HTTP method is GET",
			args: args{
				req: &http.Request{
					Method: http.MethodGet,
				},
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   []byte{},
		},
		{
			name: "should fail if request body cannot be read",
			args: args{
				req: &http.Request{
					Method: http.MethodPost,
					Body:   io.NopCloser(stubReader{readErr: fmt.Errorf("read-err")}),
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "Internal error"}),
			wantErrorLogs: []string{
				fmt.Sprintf("error ID 1: %v", app.Error{
					Type: app.ErrInternal,
					Err:  fmt.Errorf("reading request body: read-err"),
				}.Error()),
			},
		},
		{
			name: "should fail with an invalid request body",
			args: args{
				req: &http.Request{
					Method: http.MethodPost,
					Body:   readCloser([]byte("invalid-json")),
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "parsing request body: invalid character 'i' looking for beginning of value"}),
			wantErrorLogs: []string{
				fmt.Sprintf("error ID 1: %v", app.Error{
					Type: app.ErrBadRequest,
					Err:  fmt.Errorf("parsing request body: invalid character 'i' looking for beginning of value"),
				}.Error()),
			},
		},
		{
			name: "should fail if the application returns a generic error",
			args: args{
				req: &http.Request{
					Method: http.MethodPost,
					Body:   readCloser(jsonMarshal(t, api.LoginRequest{Username: "Bob Johansson"})),
				},
			},
			stubApp: stubAppLogin{
				LoginErr: fmt.Errorf("app err"),
			},
			wantStatus:    http.StatusInternalServerError,
			wantBody:      jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "Internal error"}),
			wantErrorLogs: []string{"error ID 1: app err"},
			wantSpy:       app.Credentials{Username: "Bob Johansson"},
		},
		{
			name: "should fail with a 401 if the application returns an authentication failed error",
			args: args{
				req: &http.Request{
					Method: http.MethodPost,
					Body:   readCloser(jsonMarshal(t, api.LoginRequest{Username: "Bob Johansson"})),
				},
			},
			stubApp: stubAppLogin{
				LoginErr: app.Error{Type: app.ErrAuthnFailed, Err: fmt.Errorf("additional err detail")},
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "Authentication failed"}),
			wantErrorLogs: []string{
				fmt.Sprintf("error ID 1: %v", app.Error{
					Type: app.ErrAuthnFailed,
					Err:  fmt.Errorf("additional err detail"),
				}.Error()),
			},
			wantSpy: app.Credentials{Username: "Bob Johansson"},
		},
		{
			name: "should succeed if the application succeeds",
			args: args{
				req: &http.Request{
					Method: http.MethodPost,
					Body:   readCloser(jsonMarshal(t, api.LoginRequest{Username: "Bob Johansson"})),
				},
			},
			stubApp: stubAppLogin{
				LoginSession: app.Session{
					Token:   "token1",
					Expires: time.Unix(1, 0),
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   jsonMarshal(t, api.LoginResponse{Token: "token1", Expires: time.Unix(1, 0)}),
			wantSpy:    app.Credentials{Username: "Bob Johansson"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appSpy := &spyAppLogin{stubAppLogin: tt.stubApp}
			logSpy := &spyLog{}
			var errIDIdx int
			web.Log = logSpy
			web.GenerateErrorID = func() web.ErrorID {
				errIDIdx++
				return web.ErrorID(fmt.Sprintf("%d", errIDIdx))
			}
			a := api.API{App: appSpy}
			pattern := "/api/login"
			handler := a.GetRoutes()[pattern]
			w := httptest.NewRecorder()
			handler(w, tt.args.req)

			resp := w.Result()
			gotStatus := resp.StatusCode
			gotBody, _ := io.ReadAll(resp.Body)

			if gotStatus != tt.wantStatus {
				t.Errorf("%s gotStatus = %v, wantStatus %v", pattern, gotStatus, tt.wantStatus)
			}
			if !reflect.DeepEqual(gotBody, tt.wantBody) {
				t.Errorf("%s gotBody = %s, wantBody %s", pattern, gotBody, tt.wantBody)
			}
			if !reflect.DeepEqual(appSpy.Credentials, tt.wantSpy) {
				t.Errorf("%s spied passed parameters to app = %+v, want %+v", pattern, appSpy.Credentials, tt.wantSpy)
			}
			if !reflect.DeepEqual(logSpy.errors, tt.wantErrorLogs) {
				t.Errorf("%s spied error logs = %#v, want %#v", pattern, logSpy.errors, tt.wantErrorLogs)
			}
		})
	}
}
