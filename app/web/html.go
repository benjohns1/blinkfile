package web

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/domain"
	"git.jfam.app/one-way-file-send/request"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/rate"
	"github.com/kataras/iris/v12/sessions"
	"time"
)

type (
	Config struct {
		Title                    string
		App                      App
		Port                     int
		BrowserSessionExpiration time.Duration
		ReadTimeout              time.Duration
		WriteTimeout             time.Duration
	}

	HTML struct {
		i   *iris.Application
		cfg Config
	}

	App interface {
		Login(ctx context.Context, username domain.Username, password string, requestData app.SessionRequestData) (app.Session, error)
		Logout(context.Context, app.Token) error
		IsAuthenticated(context.Context, app.Token) (domain.UserID, bool, error)
		ListFiles(context.Context, domain.UserID) ([]domain.FileHeader, error)
		UploadFile(ctx context.Context, args app.UploadFileArgs) error
		DownloadFile(ctx context.Context, userID domain.UserID, fileID domain.FileID, pass string) (domain.File, error)
		DeleteFiles(context.Context, domain.UserID, []domain.FileID) error

		app.Log
	}

	wrapper struct {
		App
	}

	LayoutView struct {
		Title string
	}

	MessageView struct {
		SuccessMessage string
		ErrorView
	}

	ErrorView struct {
		LayoutView
		ID      string
		Status  int
		Message string
	}
)

func (c Config) parse(ctx context.Context) (Config, error) {
	cfg := c
	if cfg.Title == "" {
		const defaultTitle = "File Sender"
		c.App.Printf(ctx, "Setting Title to default %q", defaultTitle)
		cfg.Title = defaultTitle
	}
	if cfg.Port <= 0 {
		return cfg, fmt.Errorf("port must be a positive number")
	}
	if cfg.ReadTimeout == 0 {
		const defaultTimeout = time.Hour
		c.App.Printf(ctx, "Setting ReadTimeout to default %v", defaultTimeout)
		cfg.ReadTimeout = defaultTimeout
	}
	if cfg.WriteTimeout == 0 {
		const defaultTimeout = time.Hour
		c.App.Printf(ctx, "Setting WriteTimeout to default %v", defaultTimeout)
		cfg.WriteTimeout = defaultTimeout
	}
	return cfg, nil
}

var (
	_ App = &app.App{}

	//go:embed assets
	assetsFS embed.FS

	//go:embed node_modules/vanillajs-datepicker/dist
	datepickerFS embed.FS

	//go:embed node_modules/dayjs
	dayjsFS embed.FS

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
	i.HandleDir("/datepicker", datepickerFS)
	i.HandleDir("/dayjs", dayjsFS)
	i.HandleDir("/", faviconFS)

	tpl := iris.HTML(templateFS, ".html").RootDir("templates")
	tpl.Layout("layouts/main.html")
	i.RegisterView(tpl)

	sess := sessions.New(sessions.Config{
		AllowReclaim: true,
		Cookie:       "session",
		Expires:      cfg.BrowserSessionExpiration,
		SessionIDGenerator: func(ctx iris.Context) string {
			return randomBase64String(64)
		},
	})

	i.Use(iris.Compression)
	i.Use(sess.Handler())
	i.Use(setDefaultViewData(cfg.Title))
	i.Use(addRequestID)
	w := wrapper{cfg.App}

	authenticated := i.Party("/")
	{
		authenticated.Use(w.f(loginRequired))
		authenticated.Get("/", w.f(showFiles))
		authenticated.Post("/files", w.f(uploadFile))
		authenticated.Post("/files/delete", w.f(deleteFiles))
	}

	unauthenticated := i.Party("/")
	{
		limit := rate.Limit(1, 3, rate.PurgeEvery(time.Minute, 5*time.Minute))
		unauthenticated.Use(limit)
		unauthenticated.Get("/login", w.f(showLogin))
		unauthenticated.Post("/login", w.f(login))
		unauthenticated.Get("/logout", w.f(logout))
		unauthenticated.Get("/file/{file_id:string}", w.f(downloadFile))
		unauthenticated.Post("/file/{file_id:string}", w.f(downloadFile))
	}

	return &HTML{i, cfg}, nil
}

func addRequestID(ctx iris.Context) {
	newCtx := request.CtxWithNewID(ctx)
	r := ctx.Request().WithContext(newCtx)
	ctx.ResetRequest(r)
	ctx.Next()
}

func setDefaultViewData(title string) func(iris.Context) {
	return func(ctx iris.Context) {
		sess := sessions.Get(ctx)
		ctx.ViewData("session", sess)
		ctx.ViewData("title", title)
		ctx.Next()
	}
}

func randomBase64String(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (w wrapper) f(h func(ctx iris.Context, a App) error) func(ctx iris.Context) {
	return injectApp(w.App, handleErrors(h))
}

func injectApp(a App, h func(ctx iris.Context, a App)) func(ctx iris.Context) {
	return func(ctx iris.Context) {
		h(ctx, a)
	}
}

func handleErrors(h func(ctx iris.Context, a App) error) func(iris.Context, App) {
	return func(ctx iris.Context, a App) {
		err := h(ctx, a)
		if err == nil {
			return
		}
		ctx.ViewData("content", ParseAppErr(ctx, err))
		err = ctx.View("error.html")
		if err == nil {
			return
		}
		_, err = ctx.HTML("<h3>%s</h3>", err)
		if err == nil {
			return
		}
		_, err = ctx.WriteString(err.Error())
		a.Errorf(ctx, err.Error())
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
