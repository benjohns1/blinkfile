package main

import (
	"context"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/repo"
	"git.jfam.app/one-way-file-send/app/web/html"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	log.Println("Exited")
}

func run(ctx context.Context) error {
	cfg := parseConfig()

	adminCredentials, err := app.NewCredentials(cfg.AdminUsername, cfg.AdminPassword)
	if err != nil {
		return err
	}
	application, appErr := app.New(app.Config{
		AdminCredentials:  adminCredentials,
		SessionRepo:       repo.NewSession(),
		SessionExpiration: 7 * 24 * time.Hour,
	})
	if appErr != nil {
		return appErr
	}

	srv, err := html.New(ctx, html.Config{
		App:                      application,
		Port:                     cfg.Port,
		BrowserSessionExpiration: 4 * 7 * 24 * time.Hour,
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
}

func parseConfig() config {
	return config{
		Port:          envDefaultInt("PORT", 8000),
		AdminUsername: envDefaultString("ADMIN_USERNAME", "admin"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
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
