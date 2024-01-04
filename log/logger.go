package log

import (
	"context"
	"fmt"
	"log"
)

type (
	Config struct {
		GetRequestID func(context.Context) string
	}
	Log struct {
		cfg          Config
		getRequestID func(context.Context) string
	}
)

var Printf = log.Printf

func New(cfg Config) Log {
	l := Log{cfg, cfg.GetRequestID}
	if l.getRequestID == nil {
		l.getRequestID = func(context.Context) string { return "" }
	}
	return l
}

func (l Log) Printf(ctx context.Context, format string, v ...any) {
	var prefix string
	if reqID := l.getRequestID(ctx); reqID != "" {
		prefix = fmt.Sprintf("%s: ", reqID)
	}
	Printf("%s%s", prefix, fmt.Sprintf(format, v...))
}

func (l Log) Errorf(ctx context.Context, format string, v ...any) {
	l.Printf(ctx, "Error: %s", fmt.Sprintf(format, v...))
}
