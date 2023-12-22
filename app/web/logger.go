package web

import (
	"context"
	"git.jfam.app/one-way-file-send/app"
)

func LogError(ctx context.Context, errID ErrorID, err error) {
	app.Log.Errorf(ctx, "error ID %s: %v", errID, err)
}
