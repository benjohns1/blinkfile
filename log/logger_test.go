package log_test

import (
	"context"
	"fmt"
	"git.jfam.app/blinkfile/log"
	"reflect"
	"testing"
)

type spy struct {
	logs []string
}

func (s *spy) Printf(format string, v ...any) {
	s.logs = append(s.logs, fmt.Sprintf(format, v...))
}

func TestLog_Printf(t *testing.T) {
	type args struct {
		ctx    context.Context
		format string
		v      []any
	}
	tests := []struct {
		name    string
		l       log.Log
		args    args
		wantSpy []string
	}{
		{
			name: "should log a simple string",
			l:    log.New(log.Config{}),
			args: args{
				ctx:    context.Background(),
				format: "log string",
			},
			wantSpy: []string{"log string"},
		},
		{
			name: "should log a formatted string",
			l:    log.New(log.Config{}),
			args: args{
				ctx:    context.Background(),
				format: "format %s %d",
				v:      []any{"s", 1},
			},
			wantSpy: []string{"format s 1"},
		},
		{
			name: "should log message with a request ID",
			l: log.New(log.Config{
				GetRequestID: func(context.Context) string {
					return "12345"
				},
			}),
			args: args{
				ctx:    context.Background(),
				format: "msg",
			},
			wantSpy: []string{"msg, Request ID: 12345"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &spy{}
			log.Printf = s.Printf
			tt.l.Printf(tt.args.ctx, tt.args.format, tt.args.v...)
			if !reflect.DeepEqual(s.logs, tt.wantSpy) {
				t.Errorf("Printf() = %v, want %v", s.logs, tt.wantSpy)
			}
		})
	}
}

func TestLog_Errorf(t *testing.T) {
	type args struct {
		ctx    context.Context
		format string
		v      []any
	}
	tests := []struct {
		name    string
		l       log.Log
		args    args
		wantSpy []string
	}{
		{
			name: "should log an error",
			l:    log.New(log.Config{}),
			args: args{
				ctx:    context.Background(),
				format: "err msg",
			},
			wantSpy: []string{"Error: err msg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &spy{}
			log.Printf = s.Printf
			tt.l.Errorf(tt.args.ctx, tt.args.format, tt.args.v...)
			if !reflect.DeepEqual(s.logs, tt.wantSpy) {
				t.Errorf("Errorf() = %v, want %v", s.logs, tt.wantSpy)
			}
		})
	}
}
