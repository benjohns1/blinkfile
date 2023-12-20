package api_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/web/api"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type stubAppLogin struct {
	returnToken app.Token
	returnErr   error
}

func (a *stubAppLogin) Login(context.Context, app.Credentials) (app.Token, error) {
	return a.returnToken, a.returnErr
}

type spyAppLogin struct {
	api.App
	stubAppLogin
	app.Credentials
}

func (a *spyAppLogin) Login(ctx context.Context, creds app.Credentials) (app.Token, error) {
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
		wantErrorLogs map[api.ErrorID]string
	}{
		{
			name: "should fail if request body cannot be read",
			args: args{
				req: &http.Request{
					Body: io.NopCloser(stubReader{readErr: fmt.Errorf("read-err")}),
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "Internal error"}),
			wantErrorLogs: map[api.ErrorID]string{"1": app.Error{
				Type: app.ErrInternal,
				Err:  fmt.Errorf("reading request body: read-err"),
			}.Error()},
		},
		{
			name: "should fail with an invalid request body",
			args: args{
				req: &http.Request{
					Body: readCloser([]byte("invalid-json")),
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "parsing request body: invalid character 'i' looking for beginning of value"}),
			wantErrorLogs: map[api.ErrorID]string{"1": app.Error{
				Type: app.ErrBadRequest,
				Err:  fmt.Errorf("parsing request body: invalid character 'i' looking for beginning of value"),
			}.Error()},
		},
		{
			name: "should fail if the application returns a generic error",
			args: args{
				req: &http.Request{
					Body: readCloser(jsonMarshal(t, api.LoginRequest{Username: "Bob Johansson"})),
				},
			},
			stubApp: stubAppLogin{
				returnErr: fmt.Errorf("app err"),
			},
			wantStatus:    http.StatusInternalServerError,
			wantBody:      jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "Internal error"}),
			wantErrorLogs: map[api.ErrorID]string{"1": "app err"},
			wantSpy:       app.Credentials{Username: "Bob Johansson"},
		},
		{
			name: "should fail with a 401 if the application returns an authentication failed error",
			args: args{
				req: &http.Request{
					Body: readCloser(jsonMarshal(t, api.LoginRequest{Username: "Bob Johansson"})),
				},
			},
			stubApp: stubAppLogin{
				returnErr: app.Error{Type: app.ErrAuthnFailed, Err: fmt.Errorf("additional err detail")},
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   jsonMarshal(t, api.ResponseErrorBody{ID: "1", Err: "Authentication failed"}),
			wantErrorLogs: map[api.ErrorID]string{"1": app.Error{
				Type: app.ErrAuthnFailed,
				Err:  fmt.Errorf("additional err detail"),
			}.Error()},
			wantSpy: app.Credentials{Username: "Bob Johansson"},
		},
		{
			name: "should succeed if the application succeeds",
			args: args{
				req: &http.Request{
					Body: readCloser(jsonMarshal(t, api.LoginRequest{Username: "Bob Johansson"})),
				},
			},
			stubApp: stubAppLogin{
				returnToken: "token1",
			},
			wantStatus: http.StatusOK,
			wantBody:   jsonMarshal(t, api.LoginResponse{AuthzToken: "token1"}),
			wantSpy:    app.Credentials{Username: "Bob Johansson"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appSpy := &spyAppLogin{stubAppLogin: tt.stubApp}
			logSpy := &spyLog{}
			var errIDIdx int
			a := api.API{
				App: appSpy,
				GenerateErrorID: func() (api.ErrorID, error) {
					errIDIdx++
					return api.ErrorID(fmt.Sprintf("%d", errIDIdx)), nil
				},
				Log: logSpy,
			}
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
