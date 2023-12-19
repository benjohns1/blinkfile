package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"
)

const shutdownWaitTime = 10 * time.Second

type ServerConfig struct {
	Port int
}

type Server struct {
	*http.Server
	ServerConfig
}

//go:embed static/*
var content embed.FS

func routes() (*http.ServeMux, error) {
	mux := http.NewServeMux()
	handleStatic, err := staticHandler()
	if err != nil {
		return nil, err
	}
	mux.Handle("/", handleStatic)
	return mux, nil
}

func staticHandler() (http.Handler, error) {
	static, err := fs.Sub(content, "static")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(static)), nil
}

func NewServer(ctx context.Context, cfg ServerConfig) (Server, error) {
	mux, err := routes()
	if err != nil {
		return Server{}, err
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	return Server{srv, cfg}, nil
}

func (s *Server) Start(ctx context.Context) <-chan error {
	done := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownWaitTime)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			done <- err
		}
	}()
	go func() {
		defer close(done)
		if err := s.ListenAndServe(); err != nil {
			done <- err
		}
	}()
	return done
}
