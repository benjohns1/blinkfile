package app

import (
	"context"
	"fmt"
	"log"
)

var Log Logger = &DefaultLogger{}

type Logger interface {
	Printf(ctx context.Context, format string, v ...any)
	Errorf(ctx context.Context, format string, v ...any)
}

type DefaultLogger struct{}

func (l *DefaultLogger) Printf(_ context.Context, format string, v ...any) {
	log.Printf(format, v...)
}

func (l *DefaultLogger) Errorf(_ context.Context, format string, v ...any) {
	log.Printf("Error: %s", fmt.Sprintf(format, v...))
}
