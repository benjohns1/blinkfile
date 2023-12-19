package main

import (
	"context"
	"git.jfam.app/one-way-file-send/web"
	"log"
	"os"
	"strconv"
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
	log.Printf("Starting server on port %d", cfg.Port)
	server, err := web.NewServer(ctx, web.ServerConfig{Port: cfg.Port})
	if err != nil {
		return err
	}
	done := server.Start(ctx)
	return <-done
}

type config struct {
	Port int
}

func parseConfig() config {
	return config{
		Port: envDefaultInt("PORT", 8000),
	}
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
