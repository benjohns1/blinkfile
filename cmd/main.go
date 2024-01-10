package main

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/app/repo"
	"github.com/benjohns1/blinkfile/app/web"
	"github.com/benjohns1/blinkfile/hash"
	"github.com/benjohns1/blinkfile/log"
	"github.com/benjohns1/blinkfile/request"
	"os"
	"strconv"
	"time"
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

	application, appErr := app.New(ctx, app.Config{
		Log:               log.New(log.Config{GetRequestID: request.GetID}),
		AdminUsername:     cfg.AdminUsername,
		AdminPassword:     cfg.AdminPassword,
		SessionExpiration: 7 * 24 * time.Hour,
		SessionRepo:       sessionRepo,
		FileRepo:          fileRepo,
		PasswordHasher:    &hash.Argon2idDefault,
	})
	if appErr != nil {
		return appErr
	}

	srv, err := web.New(ctx, web.Config{
		App:  application,
		Port: cfg.Port,
	})
	if err != nil {
		return err
	}

	done := srv.Start(ctx)
	log.Printf("Started server on port %d", cfg.Port)

	go startExpiredFileCleanup(ctx, application)

	return <-done
}

const expireCheckCycleTime = 15 * time.Minute

func startExpiredFileCleanup(ctx context.Context, a *app.App) {
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
	Port          int
	AdminUsername string
	AdminPassword string
	DataDir       string
}

func parseConfig() config {
	return config{
		Port:          envDefaultInt("PORT", 8000),
		AdminUsername: os.Getenv("ADMIN_USERNAME"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		DataDir:       envDefaultString("DATA_DIR", "./data"),
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
