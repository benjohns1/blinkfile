package testautomation

import (
	"context"
	"fmt"
	"time"

	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/longduration"
)

type (
	Automator struct {
		Log app.Log
		Clock
		App
	}

	Args struct {
		DeleteAllFiles bool
		TimeOffset     longduration.LongDuration
	}

	Clock interface {
		Now() time.Time
		SetTimeOffset(time.Duration)
		TimeOffset() time.Duration
	}

	App interface {
		DeleteExpiredFiles(ctx context.Context) error
	}
)

const maxDuration = time.Duration(^uint(0) >> 1)

func (a *Automator) TestAutomation(ctx context.Context, args Args) {
	if args.DeleteAllFiles {
		prevOffset := a.Clock.TimeOffset()
		a.Clock.SetTimeOffset(maxDuration)
		if err := a.App.DeleteExpiredFiles(ctx); err != nil {
			panic(fmt.Errorf("deleting files: %v", err))
		}
		a.Clock.SetTimeOffset(prevOffset)
	}
	if args.TimeOffset != "" {
		d, err := args.TimeOffset.Duration()
		if err != nil {
			panic(fmt.Errorf("parsing duration for fast-forward: %v", err))
		}
		a.Clock.SetTimeOffset(d)
		a.Log.Printf(ctx, "fast-forwarded time by %v, new clock time is %v", args.TimeOffset, a.Clock.Now())
	}
}
