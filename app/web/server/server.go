package server

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/web"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"time"
)

const shutdownWaitTime = 10 * time.Second

type (
	Config struct {
		Port      int
		APIRoutes map[string]http.HandlerFunc
		App
	}

	App interface {
		Login(context.Context, app.Credentials) (app.Session, error)
	}

	Server struct {
		*http.Server
		Config
	}

	Controllers struct {
		App
	}
)

var (
	_ App = &app.App{}

	//go:embed static
	staticFS embed.FS

	//go:embed template
	templateFS embed.FS
)

func routes(ctrl Controllers, cfg Config) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	handleStatic, err := staticHandler()
	if err != nil {
		return nil, err
	}
	mux.Handle("/", handleStatic)
	mux.HandleFunc("/login.html", ctrl.handleLogin)
	for pattern, handler := range cfg.APIRoutes {
		mux.HandleFunc(pattern, handler)
	}
	return mux, nil
}

func staticHandler() (http.Handler, error) {
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(static)), nil
}

func New(ctx context.Context, cfg Config) (Server, error) {
	ctrl := Controllers{App: cfg.App}
	mux, err := routes(ctrl, cfg)
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

var cachedTemplates = map[string]*template.Template{}

func getTemplate(name string) (*template.Template, error) {
	if tpl, ok := cachedTemplates[name]; ok {
		return tpl, nil
	}
	tpl, err := template.ParseFS(templateFS, fmt.Sprintf("template/%s", name))
	if err != nil {
		return nil, err
	}
	cachedTemplates[name] = tpl
	return tpl, nil
}

func renderTemplate[T any](ctx context.Context, name string, w http.ResponseWriter, view T) bool {
	tpl, err := getTemplate(name)
	if err != nil {
		if !renderError(ctx, w, err) {
			writeRawError(ctx, w, err)
		}
		return false
	}

	var buf bytes.Buffer
	if execErr := tpl.Execute(&buf, view); execErr != nil {
		writeRawError(ctx, w, execErr)
		return false
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if _, writeErr := buf.WriteTo(w); writeErr != nil {
		web.Log.Errorf(ctx, "writing template buffer to response: %v", writeErr)
		return false
	}

	return true
}

type ErrorView struct {
	ID      web.ErrorID
	Status  int
	Message string
}

func renderError(ctx context.Context, w http.ResponseWriter, err error) bool {
	errID, errStatus, errMsg := web.ParseAppErr(err)
	web.LogError(ctx, errID, err)
	return renderTemplate(ctx, "error.html", w, ErrorView{
		ID:      errID,
		Status:  errStatus,
		Message: errMsg,
	})
}

func writeRawError(ctx context.Context, w http.ResponseWriter, err error) {
	errID, errStatus, errMsg := web.ParseAppErr(err)
	web.LogError(ctx, errID, err)
	web.WriteResponse(w, errStatus, []byte(fmt.Sprintf(`<html><head><title>Error</title></head><body><h1>Fatal Error</h1><p>ID: %s<br/>%s</p></body></html>`, errID, errMsg)))
}
