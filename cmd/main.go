package main

import (
	"context"
	"fmt"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/repo"
	"git.jfam.app/one-way-file-send/app/web"
	"git.jfam.app/one-way-file-send/hash"
	"git.jfam.app/one-way-file-send/log"
	"git.jfam.app/one-way-file-send/request"
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

	log.Printf("Starting server on port %d", cfg.Port)
	done := srv.Start(ctx)
	return <-done
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
