package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/benjohns1/blinkfile/app/testautomation"

	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/app/repo"
	"github.com/benjohns1/blinkfile/app/web"
	"github.com/benjohns1/blinkfile/featureflag"
	"github.com/benjohns1/blinkfile/hash"
	"github.com/benjohns1/blinkfile/log"
	"github.com/benjohns1/blinkfile/request"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(0)
	}
	log.Printf("Exited")
}

func run(ctx context.Context) (err error) {
	cfg := parseConfig()

	sessionRepo, err := repo.NewSession(repo.SessionConfig{
		Dir: fmt.Sprintf("%s/sessions", cfg.DataDir),
	})
	if err != nil {
		return err
	}

	fileRepo, err := repo.NewFileRepo(ctx, repo.FileRepoConfig{
		Dir: fmt.Sprintf("%s/files", cfg.DataDir),
	})
	if err != nil {
		return err
	}

	l := log.New(log.Config{GetRequestID: request.GetID})

	ff, err := featureflag.New(featureflag.WithFeaturesFromEnvironment("FEATURE_FLAG_"))
	if err != nil {
		return err
	}
	app.FeatureFlagIsOn = func(ctx context.Context, feature string) bool {
		v, ffErr := ff.IsOn(ctx, feature)
		if ffErr != nil {
			l.Errorf(ctx, "checking feature flag %q: %v", feature, ffErr)
		}
		return v
	}

	appConfig := app.Config{
		Log:               l,
		AdminUsername:     cfg.AdminUsername,
		AdminPassword:     cfg.AdminPassword,
		SessionExpiration: 7 * 24 * time.Hour,
		SessionRepo:       sessionRepo,
		FileRepo:          fileRepo,
		PasswordHasher:    &hash.Argon2idDefault,
	}

	var automator *testautomation.Automator
	if cfg.EnableTestAutomation {
		l.Printf(ctx, "WARNING: Server running with test automation enabled! DO NOT RUN THESE IN PRODUCTION!")
		testClock := &testautomation.TestClock{}
		automator = &testautomation.Automator{
			Log:      l,
			Clock:    testClock,
			FileRepo: fileRepo,
		}
		appConfig.Clock = testClock
	}

	application, appErr := app.New(ctx, appConfig)
	if appErr != nil {
		return appErr
	}

	srv, err := web.New(ctx, web.Config{
		App:                           application,
		Port:                          cfg.Port,
		RateLimitUnauthenticated:      cfg.RateLimitUnauthenticated,
		RateLimitBurstUnauthenticated: cfg.RateLimitBurstUnauthenticated,
		TestAutomator:                 automator,
	})
	if err != nil {
		return err
	}

	done := srv.Start(ctx)
	log.Printf("Started server on port %d", cfg.Port)

	go startExpiredFileCleanup(ctx, application, cfg.ExpireCheckCycleTime)

	return <-done
}

func startExpiredFileCleanup(ctx context.Context, a *app.App, expireCheckCycleTime time.Duration) {
	log.Printf("Starting expired file deletion process, running every %v", expireCheckCycleTime)
	for {
		if err := ctx.Err(); err != nil {
			break
		}
		time.Sleep(expireCheckCycleTime)
		err := a.DeleteExpiredFiles(ctx)
		if err != nil {
			log.Printf("Error deleting expired files: %v", err)
		}
	}
	log.Printf("Stopped expired file deletion process")
}

type config struct {
	Port                          int
	AdminUsername                 string
	AdminPassword                 string
	DataDir                       string
	RateLimitUnauthenticated      float64
	RateLimitBurstUnauthenticated int
	EnableTestAutomation          bool
	ExpireCheckCycleTime          time.Duration
}

func parseConfig() config {
	return config{
		Port:                          envDefaultInt("PORT", 8000),
		AdminUsername:                 os.Getenv("ADMIN_USERNAME"),
		AdminPassword:                 os.Getenv("ADMIN_PASSWORD"),
		DataDir:                       envDefaultString("DATA_DIR", "./data"),
		RateLimitUnauthenticated:      envDefaultFloat("RATE_LIMIT_UNAUTHENTICATED", 2),
		RateLimitBurstUnauthenticated: envDefaultInt("RATE_LIMIT_BURST_UNAUTHENTICATED", 5),
		EnableTestAutomation:          envDefaultBool("ENABLE_TEST_AUTOMATION", false),
		ExpireCheckCycleTime:          envDefaultDuration("EXPIRE_CHECK_CYCLE_TIME", 15*time.Minute),
	}
}

func envDefaultString(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func envDefaultInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return def
	}
	return int(i)
}

func envDefaultFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func envDefaultBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envDefaultDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
