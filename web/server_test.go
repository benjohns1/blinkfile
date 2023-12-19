package web

import (
	"context"
	"fmt"
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

func TestStartServerLocalhost(t *testing.T) {
	testPort := findOpenTestPort(t)
	type args struct {
		cfg ServerConfig
	}
	tests := []struct {
		name    string
		args    args
		assert  func(*testing.T)
		wantErr error
	}{
		{
			name: fmt.Sprintf("should start an HTTP server on port %d and serve a 200 response", testPort),
			args: args{
				cfg: ServerConfig{Port: testPort},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			srv, err := NewServer(ctx, tt.args.cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			srv.Start(ctx)
			if tt.assert != nil {
				tt.assert(t)
			}
		})
	}
}
