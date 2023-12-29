package request_test

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/request"
	"reflect"
	"testing"
)

func TestNewID(t *testing.T) {
	tests := []struct {
		name      string
		randRead  func(b []byte) (n int, err error)
		want      string
		wantPanic any
	}{
		{
			name: "should panic if generating a random byte slice fails",
			randRead: func(b []byte) (n int, err error) {
				return 0, fmt.Errorf("rand read err")
			},
			wantPanic: fmt.Errorf("rand read err"),
		},
		{
			name: "should generate a valid 32-byte hex-encoded request ID",
			randRead: func(b []byte) (n int, err error) {
				for i := range b {
					b[i] = 1
				}
				return 0, nil
			},
			want: "0101010101010101010101010101010101010101010101010101010101010101",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.randRead != nil {
				prev := request.ReadRandomBytes
				request.ReadRandomBytes = tt.randRead
				defer func() { request.ReadRandomBytes = prev }()
			}
			defer func() {
				r := recover()
				if !reflect.DeepEqual(r, tt.wantPanic) {
					t.Errorf("NewID() panic = %#v, wantPanic %#v", r, tt.wantPanic)
				}
			}()
			if got := request.NewID(); got != tt.want {
				t.Errorf("NewID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		randRead func(b []byte) (n int, err error)
		args     args
		want     string
	}{
		{
			name: "should retrieve an empty string if request ID is not set",
			args: args{
				ctx: context.Background(),
			},
			want: "",
		},
		{
			name: "should retrieve a valid 32-byte hex-encoded request ID",
			args: args{
				ctx: func() context.Context {
					prev := request.ReadRandomBytes
					request.ReadRandomBytes = func(b []byte) (n int, err error) {
						for i := range b {
							b[i] = 128
						}
						return 0, nil
					}
					defer func() { request.ReadRandomBytes = prev }()

					return request.CtxWithNewID(context.Background())
				}(),
			},
			want: "8080808080808080808080808080808080808080808080808080808080808080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := request.GetID(tt.args.ctx); got != tt.want {
				t.Errorf("GetID() = %v, want %v", got, tt.want)
			}
		})
	}
}
