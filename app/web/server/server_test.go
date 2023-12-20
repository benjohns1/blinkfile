package server_test

import (
	"bytes"
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app/web/server"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func findOpenPort() (int, error) {
	ln, err := net.Listen("tcp", ":")
	defer func() { _ = ln.Close() }()
	if err != nil {
		return 0, err
	}
	addr := ln.Addr().String()
	parts := strings.Split(addr, ":")
	port, err := strconv.ParseInt(parts[len(parts)-1], 10, 32)
	return int(port), err
}

func findOpenTestPort(t *testing.T) int {
	t.Helper()
	port, err := findOpenPort()
	if err != nil {
		t.Fatalf("finding open port: %v", err)
	}
	return port
}

type stubAPI struct {
	routes map[string]func(http.ResponseWriter, *http.Request)
}

func (api *stubAPI) GetRoutes() map[string]func(http.ResponseWriter, *http.Request) {
	return api.routes
}

func TestStartServerLocalhost(t *testing.T) {
	testPort := findOpenTestPort(t)
	type args struct {
		cfg server.Config
	}
	tests := []struct {
		name    string
		args    args
		assert  func(*testing.T)
		wantErr error
	}{
		{
			name: "should serve a static file 200 response",
			args: args{
				cfg: server.Config{
					Port: testPort,
					API:  &stubAPI{},
				},
			},
			assert: func(t *testing.T) {
				resp, err := http.Get(fmt.Sprintf("http://localhost:%d", testPort))
				if err != nil {
					t.Errorf("serve response had unexpected error = %v", err)
					return
				}
				got, want := resp.StatusCode, http.StatusOK
				if got != want {
					t.Errorf("serve response status = %v, want %v", got, want)
				}
			},
		},
		{
			name: "should serve an API endpoint success response",
			args: args{
				cfg: server.Config{
					Port: testPort,
					API: &stubAPI{routes: map[string]func(http.ResponseWriter, *http.Request){
						"/test/route": func(w http.ResponseWriter, req *http.Request) {
							w.WriteHeader(http.StatusTeapot)
							_, err := w.Write([]byte("response body"))
							if err != nil {
								t.Fatal(err)
							}
						},
					}},
				},
			},
			assert: func(t *testing.T) {
				resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test/route", testPort))
				if err != nil {
					t.Errorf("serve response had unexpected error = %v", err)
					return
				}
				gotStatus, wantStatus := resp.StatusCode, http.StatusTeapot
				if gotStatus != wantStatus {
					t.Errorf("serve response status = %v, want %v", gotStatus, wantStatus)
				}
				gotBody, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				wantBody := []byte("response body")
				if !bytes.Equal(gotBody, wantBody) {
					t.Errorf("serve response body = %s, want %s", gotBody, wantBody)
				}
			},
		},
		{
			name: "should serve an API endpoint error response",
			args: args{
				cfg: server.Config{
					Port: testPort,
					API: &stubAPI{routes: map[string]func(http.ResponseWriter, *http.Request){
						"/test/error/route": func(w http.ResponseWriter, req *http.Request) {
							w.WriteHeader(http.StatusNotAcceptable)
							_, err := w.Write([]byte("error response"))
							if err != nil {
								t.Fatal(err)
							}
						},
					}},
				},
			},
			assert: func(t *testing.T) {
				resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test/error/route", testPort))
				if err != nil {
					t.Errorf("serve response had unexpected error = %v", err)
					return
				}
				gotStatus, wantStatus := resp.StatusCode, http.StatusNotAcceptable
				if gotStatus != wantStatus {
					t.Errorf("serve response status = %v, want %v", gotStatus, wantStatus)
				}
				gotBody, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				wantBody := []byte("error response")
				if !bytes.Equal(gotBody, wantBody) {
					t.Errorf("serve response body = %s, want %s", gotBody, wantBody)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			srv, err := server.New(ctx, tt.args.cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			srv.Start(ctx)
			if tt.assert != nil {
				tt.assert(t)
			}
		})
	}
}
