package web

import (
	"context"
	"fmt"
	"log"
)

var Log Logger = &DefaultLogger{}

type Logger interface {
	Errorf(_ context.Context, format string, v ...any)
}

type DefaultLogger struct{}

func (l *DefaultLogger) Errorf(_ context.Context, format string, v ...any) {
	log.Printf("Error: %s", fmt.Sprintf(format, v...))
}

func LogError(ctx context.Context, errID ErrorID, err error) {
	Log.Errorf(ctx, "error ID %s: %v", errID, err)
}
