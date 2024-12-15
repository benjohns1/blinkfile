package web

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/benjohns1/blinkfile/app/testautomation"
	"github.com/benjohns1/blinkfile/longduration"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/request"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/rate"
	"github.com/kataras/iris/v12/sessions"
)

type (
	Config struct {
		Title                         string
		App                           App
		Port                          int
		MaxFileByteSize               int64
		BrowserSessionExpiration      time.Duration
		ReadTimeout                   time.Duration
		WriteTimeout                  time.Duration
		RateLimitUnauthenticated      float64
		RateLimitBurstUnauthenticated int
		TestAutomator                 TestAutomator
	}

	HTML struct {
		i   *iris.Application
		cfg Config
	}

	App interface {
		Login(ctx context.Context, username blinkfile.Username, password string, requestData app.SessionRequestData) (app.Session, error)
		Logout(context.Context, app.Token) error
		IsAuthenticated(context.Context, app.Token) (blinkfile.UserID, bool, error)
		ListFiles(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error)
		UploadFile(context.Context, app.UploadFileArgs) error
		DownloadFile(ctx context.Context, userID blinkfile.UserID, fileID blinkfile.FileID, pass string) (blinkfile.FileHeader, error)
		DeleteFiles(context.Context, blinkfile.UserID, []blinkfile.FileID) error
		SubscribeToFileChanges(blinkfile.UserID) (<-chan app.FileEvent, func())
		CreateUser(context.Context, app.CreateUserArgs) error
		ChangeUsername(context.Context, app.ChangeUsernameArgs) error
		ListUsers(context.Context) ([]blinkfile.User, error)
		GetUserByID(context.Context, blinkfile.UserID) (blinkfile.User, error)
		DeleteUsers(context.Context, []blinkfile.UserID) error

		app.Log
	}

	TestAutomator interface {
		TestAutomation(ctx context.Context, args testautomation.Args) error
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
)

func (c Config) parse(ctx context.Context) (Config, error) {
	cfg := c
	if cfg.MaxFileByteSize <= 0 {
		const defaultMaxFilesize = 2 * iris.GB
		cfg.MaxFileByteSize = defaultMaxFilesize
	}
	if cfg.Title == "" {
		const defaultTitle = "Blinkfile"
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
	if cfg.RateLimitUnauthenticated <= 0 {
		cfg.RateLimitUnauthenticated = 2
	}
	if cfg.RateLimitBurstUnauthenticated <= 0 {
		cfg.RateLimitBurstUnauthenticated = 5
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

	//go:embed node_modules/the-basicest-tablesort
	tablesortFS embed.FS

	//go:embed favicon
	faviconFS embed.FS

	//go:embed templates
	templateFS embed.FS
)

func verifyTmpDir() error {
	tmpDir := os.TempDir()
	if err := os.MkdirAll(tmpDir, os.ModeDir); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp("", "init_check_")
	if err != nil {
		return fmt.Errorf("creating test tmp file: %w", err)
	}
	defer func() { _ = tmpFile.Close() }()
	tmpName := tmpFile.Name()
	_, err = fmt.Fprintf(tmpFile, "write_test_data")
	if err != nil {
		return fmt.Errorf("writing to test tmp file %q: %w", tmpName, err)
	}
	return nil
}

func New(ctx context.Context, cfg Config) (html *HTML, err error) {
	cfg, err = cfg.parse(ctx)
	if err != nil {
		return nil, err
	}
	err = verifyTmpDir()
	if err != nil {
		return nil, fmt.Errorf("verifying tmp directory: %w", err)
	}

	i := iris.New()

	i.HandleDir("/assets", assetsFS)
	i.HandleDir("/datepicker", datepickerFS)
	i.HandleDir("/dayjs", dayjsFS)
	i.HandleDir("/tablesort", tablesortFS)
	i.HandleDir("/", faviconFS)

	tpl := iris.HTML(templateFS, ".html").RootDir("templates")
	tpl.Layout("layouts/main.html")
	tpl.AddFunc("featureFlagIsOn", app.FeatureFlagIsOn)
	i.RegisterView(tpl)

	const sessionIDLength = 64
	sess := sessions.New(sessions.Config{
		AllowReclaim: true,
		Cookie:       "session",
		Expires:      cfg.BrowserSessionExpiration,
		SessionIDGenerator: func(ctx iris.Context) string {
			return randomBase64String(sessionIDLength)
		},
	})
	i.Use(iris.Compression)
	i.Use(func(ctx iris.Context) {
		ctx = withRequestID(ctx)
		ctx.Next()
	})
	i.Use(logRequest(cfg.App))
	i.UseError(func(ctx iris.Context) {
		reqID := request.GetID(ctx)
		if reqID == "" {
			ctx = withRequestID(ctx)
		}
		logRequestStart(ctx, cfg.App)
		appErr := parseIrisErr(ctx)
		showError(ctx, cfg.App, appErr)
	})
	i.Use(sess.Handler())
	i.Use(setDefaultViewData(cfg.Title))
	w := wrapper{cfg.App}

	authenticated := i.Party("/")
	{
		authenticated.Use(w.f(loginRequired))
		authenticated.Get("/", w.f(showFiles))
		upload := authenticated.Post("/files", w.f(uploadFile))
		upload.Use(maxSize(cfg.MaxFileByteSize))
		authenticated.Post("/files/delete", w.f(deleteFiles))
		authenticated.Any("/files/notifications", w.f(fileNotifications))

		if app.FeatureFlagIsOn(ctx, "UserAccounts") {
			userMgmt := authenticated.Party("/users")
			{
				userMgmt.Use(w.f(requirePermission("user_management")))
				userMgmt.Get("/", w.f(showUsers))
				userMgmt.Post("/", w.f(createUser))
				userMgmt.Get("/{user_id:string}/edit", w.f(showEditUser))
				userMgmt.Post("/{user_id:string}/edit", w.f(editUser))
				userMgmt.Post("/delete", w.f(deleteUsers))
			}
		}

		if cfg.TestAutomator != nil {
			authenticated.Post("/test-automation", func(ctx iris.Context) {
				var deleteUserFiles blinkfile.UserID
				doDeleteUserFiles, _ := strconv.ParseBool(ctx.FormValue("delete_user_files"))
				if doDeleteUserFiles {
					deleteUserFiles = loggedInUser(ctx)
				}
				deleteAllUsers, _ := strconv.ParseBool(ctx.FormValue("delete_all_users"))
				if aErr := cfg.TestAutomator.TestAutomation(ctx, testautomation.Args{
					DeleteUserFiles: deleteUserFiles,
					TimeOffset:      longduration.LongDuration(ctx.FormValue("time_offset")),
					DeleteAllUsers:  deleteAllUsers,
				}); aErr != nil {
					panic(aErr)
				}
			})
		}
	}

	unauthenticated := i.Party("/")
	{
		const (
			purgeEvery       = time.Minute
			purgeMaxLifetime = 5 * time.Minute
		)
		limit := rate.Limit(cfg.RateLimitUnauthenticated, cfg.RateLimitBurstUnauthenticated, rate.PurgeEvery(purgeEvery, purgeMaxLifetime))
		unauthenticated.Use(limit)
		unauthenticated.Get("/login", w.f(showLogin))
		unauthenticated.Post("/login", w.f(login))
		unauthenticated.Get("/logout", w.f(logout))
		unauthenticated.Get("/file/{file_id:string}", w.f(downloadFile))
		unauthenticated.Post("/file/{file_id:string}", w.f(downloadFile))
	}

	return &HTML{i, cfg}, nil
}

func logRequestStart(ctx iris.Context, l app.Log) {
	req := ctx.Request()
	l.Printf(ctx, "%s %s | RemoteAddr: %s | UserAgent: %s", req.Method, req.RequestURI, req.RemoteAddr, req.UserAgent())
}

func logRequest(l app.Log) func(iris.Context) {
	return func(ctx iris.Context) {
		start := time.Now()
		logRequestStart(ctx, l)
		ctx.Next()
		resp := ctx.ResponseWriter()
		l.Printf(ctx, "%d response took %v", resp.StatusCode(), time.Since(start))
	}
}

func maxSize(byteSize int64) func(iris.Context) {
	return func(ctx iris.Context) {
		ctx.SetMaxRequestBodySize(byteSize)
		ctx.Next()
	}
}

func withRequestID(ctx iris.Context) iris.Context {
	newCtx := request.CtxWithNewID(ctx)
	r := ctx.Request().WithContext(newCtx)
	ctx.ResetRequest(r)
	return ctx
}

func setDefaultViewData(title string) func(iris.Context) {
	return func(ctx iris.Context) {
		sess := sessions.Get(ctx)
		ctx.ViewData("session", sess)
		ctx.ViewData("title", title)
		ctx.ViewData("ctx", ctx)
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
