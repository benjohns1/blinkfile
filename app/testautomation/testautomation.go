package testautomation

import (
	"context"
	"fmt"
	"time"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/longduration"
)

type (
	Automator struct {
		Log app.Log
		Clock
		FileRepo
		UserRepo UserRepo
	}

	Args struct {
		DeleteUserFiles blinkfile.UserID
		TimeOffset      longduration.LongDuration
		DeleteAllUsers  bool
	}

	Clock interface {
		Now() time.Time
		SetTimeOffset(time.Duration)
	}

	FileRepo interface {
		ListByUser(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error)
		Delete(context.Context, blinkfile.UserID, []blinkfile.FileID) error
	}

	UserRepo interface {
		ListAll(context.Context) ([]blinkfile.User, error)
		Delete(context.Context, blinkfile.UserID) error
	}
)

func (a *Automator) TestAutomation(ctx context.Context, args Args) error {
	if args.DeleteUserFiles != "" {
		files, err := a.FileRepo.ListByUser(ctx, args.DeleteUserFiles)
		if err != nil {
			return err
		}
		fileIDs := make([]blinkfile.FileID, 0, len(files))
		for _, file := range files {
			fileIDs = append(fileIDs, file.ID)
		}
		err = a.Delete(ctx, args.DeleteUserFiles, fileIDs)
		if err != nil {
			return fmt.Errorf("deleting files: %v", err)
		}
		a.Log.Printf(ctx, "deleted all %q files", args.DeleteUserFiles)
	}
	if args.TimeOffset != "" {
		d, err := args.TimeOffset.Duration()
		if err != nil {
			return fmt.Errorf("parsing duration for fast-forward: %v", err)
		}
		a.Clock.SetTimeOffset(d)
		a.Log.Printf(ctx, "fast-forwarded time by %v, new clock time is %v", args.TimeOffset, a.Clock.Now())
	}
	if args.DeleteAllUsers {
		users, err := a.UserRepo.ListAll(ctx)
		if err != nil {
			return err
		}
		for _, user := range users {
			err = a.UserRepo.Delete(ctx, user.ID)
			if err != nil {
				return err
			}
		}
		a.Log.Printf(ctx, "deleted all users")
	}
	return nil
}
