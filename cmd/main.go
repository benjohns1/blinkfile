package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/repo"
	"git.jfam.app/one-way-file-send/app/web/api"
	"git.jfam.app/one-way-file-send/app/web/server"
	"log"
	"os"
	"strconv"
	"strings"
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

	application, appErr := app.New(app.Config{
		AdminCredentials: app.Credentials{
			Username: cfg.AdminUsername,
		},
		SessionExpiration: 7 * 24 * time.Hour,
		SessionRepo:       repo.NewSession(),
		GenerateToken: func() (app.Token, error) {
			v, err := randomBase64String(128)
			return app.Token(v), err
		},
	})
	if appErr != nil {
		return appErr
	}

	webAPI := &api.API{App: application}

	log.Printf("Starting server on port %d", cfg.Port)
	srv, srvErr := server.New(ctx, server.Config{
		Port:      cfg.Port,
		APIRoutes: webAPI.GetRoutes(),
		App:       application,
	})
	if srvErr != nil {
		return srvErr
	}
	done := srv.Start(ctx)
	return <-done
}

func randomBase64String(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return strings.ToLower(base64.StdEncoding.EncodeToString(b)), nil
}

type config struct {
	Port          int
	AdminUsername string
}

func parseConfig() config {
	return config{
		Port:          envDefaultInt("PORT", 8000),
		AdminUsername: envDefaultString("ADMIN_USERNAME", "admin"),
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
