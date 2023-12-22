package html

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/web"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/rate"
	"github.com/kataras/iris/v12/sessions"
	"time"
)

type (
	Config struct {
		App               App
		Port              int
		SessionExpiration time.Duration
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
	}

	HTML struct {
		i   *iris.Application
		cfg Config
	}

	App interface {
		Authenticate(ctx context.Context, username, password string) error
	}

	wrapper struct {
		App
	}

	LayoutView struct {
		Title string
	}

	ErrorView struct {
		LayoutView
		ID      web.ErrorID
		Status  int
		Message string
	}
)

func (c Config) parse(ctx context.Context) (Config, error) {
	cfg := c
	if cfg.Port <= 0 {
		return cfg, fmt.Errorf("port must be a positive number")
	}
	if cfg.ReadTimeout == 0 {
		const defaultTimeout = time.Hour
		app.Log.Printf(ctx, "setting ReadTimeout to default %v", defaultTimeout)
		cfg.ReadTimeout = defaultTimeout
	}
	if cfg.WriteTimeout == 0 {
		const defaultTimeout = time.Hour
		app.Log.Printf(ctx, "setting WriteTimeout to default %v", defaultTimeout)
		cfg.WriteTimeout = defaultTimeout
	}
	return cfg, nil
}

var (
	_ App = &app.App{}

	//go:embed assets
	assetsFS embed.FS

	//go:embed favicon
	faviconFS embed.FS

	//go:embed templates
	templateFS embed.FS
)

func New(ctx context.Context, cfg Config) (html *HTML, err error) {
	cfg, err = cfg.parse(ctx)
	if err != nil {
		return nil, err
	}
	i := iris.New()

	i.HandleDir("/assets", assetsFS)
	i.HandleDir("/", faviconFS)

	tpl := iris.HTML(templateFS, ".html").RootDir("templates")
	tpl.Layout("layouts/main.html")
	i.RegisterView(tpl)

	sess := sessions.New(sessions.Config{
		AllowReclaim: true,
		Cookie:       "session",
		Expires:      cfg.SessionExpiration,
		SessionIDGenerator: func(iris.Context) string {
			return randomBase64String(128)
		},
	})

	i.Use(iris.Compression)
	i.Use(sess.Handler())
	i.Use(setSessionViewData)
	w := wrapper{cfg.App}

	authenticated := i.Party("/")
	{
		authenticated.Use(w.f(loginRequired))
		authenticated.Get("/", w.f(showDashboard))
	}

	unauthenticated := i.Party("/")
	{
		limit := rate.Limit(1, 3, rate.PurgeEvery(time.Minute, 5*time.Minute))
		unauthenticated.Use(limit)
		unauthenticated.Get("/login", w.f(showLogin))
		unauthenticated.Post("/login", w.f(doLogin))
		unauthenticated.Get("/logout", w.f(logout))
	}

	return &HTML{i, cfg}, nil
}

func setSessionViewData(ctx iris.Context) {
	sess := sessions.Get(ctx)
	ctx.ViewData("session", sess)
	ctx.Next()
}

func randomBase64String(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (w wrapper) f(h func(ctx iris.Context, a App) error) func(ctx iris.Context) {
	return handleErrors(injectApp(w.App, h))
}

func injectApp(a App, h func(ctx iris.Context, a App) error) func(ctx iris.Context) error {
	return func(ctx iris.Context) error {
		return h(ctx, a)
	}
}

func handleErrors(h func(ctx iris.Context) error) func(iris.Context) {
	return func(ctx iris.Context) {
		err := h(ctx)
		if err == nil {
			return
		}
		errID, errStatus, errMsg := web.ParseAppErr(err)
		web.LogError(ctx, errID, err)
		err = ctx.View("error.html", ErrorView{
			ID:      errID,
			Status:  errStatus,
			Message: errMsg,
		})
		if err == nil {
			return
		}
		_, err = ctx.HTML("<h3>%s</h3>", err)
		if err == nil {
			return
		}
		_, err = ctx.WriteString(err.Error())
		web.LogError(ctx, errID, err)
	}
}

const shutdownWaitTime = 10 * time.Second

func (html *HTML) Start(ctx context.Context) <-chan error {
	done := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownWaitTime)
		defer cancel()
		if err := html.i.Shutdown(shutdownCtx); err != nil {
			done <- err
		}
	}()
	go func() {
		defer close(done)
		addr := fmt.Sprintf(":%d", html.cfg.Port)
		if err := html.i.Listen(addr); err != nil {
			done <- err
		}
	}()
	return done
}
