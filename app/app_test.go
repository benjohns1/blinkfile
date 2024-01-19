package app_test

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/hash"
	"github.com/benjohns1/blinkfile/log"
	"reflect"
	"testing"
	"time"
)

func AppConfigDefaults(cfg app.Config) app.Config {
	out := cfg
	if cfg.SessionRepo == nil {
		out.SessionRepo = &StubSessionRepo{}
	}

	if cfg.FileRepo == nil {
		out.FileRepo = &StubFileRepo{}
	}

	if cfg.Log == nil {
		out.Log = log.New(log.Config{})
	}

	if cfg.PasswordHasher == nil {
		out.PasswordHasher = &hash.Argon2idDefault
	}
	return out
}

func NewTestApp(ctx context.Context, t *testing.T, cfg app.Config) *app.App {
	application, err := app.New(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return application
}

type SpySessionRepo struct {
	repo        app.SessionRepo
	SaveCalls   []app.Session
	DeleteCalls []app.Token
	GetCalls    []app.Token
}

func (sr *SpySessionRepo) Save(ctx context.Context, session app.Session) error {
	sr.SaveCalls = append(sr.SaveCalls, session)
	return sr.repo.Save(ctx, session)
}

func (sr *SpySessionRepo) Delete(ctx context.Context, token app.Token) error {
	sr.DeleteCalls = append(sr.DeleteCalls, token)
	return sr.repo.Delete(ctx, token)
}

func (sr *SpySessionRepo) Get(ctx context.Context, token app.Token) (app.Session, bool, error) {
	sr.GetCalls = append(sr.GetCalls, token)
	return sr.repo.Get(ctx, token)
}

type StubSessionRepo struct {
	SaveFunc   func(context.Context, app.Session) error
	GetFunc    func(context.Context, app.Token) (app.Session, bool, error)
	DeleteFunc func(context.Context, app.Token) error
}

func (sr *StubSessionRepo) Save(ctx context.Context, s app.Session) error {
	if sr.SaveFunc != nil {
		return sr.SaveFunc(ctx, s)
	}
	return nil
}

func (sr *StubSessionRepo) Get(ctx context.Context, t app.Token) (app.Session, bool, error) {
	if sr.GetFunc != nil {
		return sr.GetFunc(ctx, t)
	}
	return app.Session{}, false, nil
}

func (sr *StubSessionRepo) Delete(ctx context.Context, t app.Token) error {
	if sr.DeleteFunc != nil {
		return sr.DeleteFunc(ctx, t)
	}
	return nil
}

type StubFileRepo struct {
	SaveFunc                func(context.Context, blinkfile.File) error
	ListByUserFunc          func(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error)
	DeleteExpiredBeforeFunc func(context.Context, time.Time) (int, error)
	GetFunc                 func(context.Context, blinkfile.FileID) (blinkfile.FileHeader, error)
	DeleteFunc              func(context.Context, blinkfile.UserID, []blinkfile.FileID) error
}

func (fr *StubFileRepo) Save(ctx context.Context, f blinkfile.File) error {
	if fr.SaveFunc != nil {
		return fr.SaveFunc(ctx, f)
	}
	return nil
}
func (fr *StubFileRepo) ListByUser(ctx context.Context, uID blinkfile.UserID) ([]blinkfile.FileHeader, error) {
	if fr.ListByUserFunc != nil {
		return fr.ListByUserFunc(ctx, uID)
	}
	return nil, nil
}
func (fr *StubFileRepo) DeleteExpiredBefore(ctx context.Context, t time.Time) (int, error) {
	if fr.DeleteExpiredBeforeFunc != nil {
		return fr.DeleteExpiredBeforeFunc(ctx, t)
	}
	return 0, nil
}
func (fr *StubFileRepo) Get(ctx context.Context, fID blinkfile.FileID) (blinkfile.FileHeader, error) {
	if fr.GetFunc != nil {
		return fr.GetFunc(ctx, fID)
	}
	return blinkfile.FileHeader{}, nil
}
func (fr *StubFileRepo) Delete(ctx context.Context, uID blinkfile.UserID, fID []blinkfile.FileID) error {
	if fr.DeleteFunc != nil {
		return fr.DeleteFunc(ctx, uID, fID)
	}
	return nil
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	type args struct {
		cfg app.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		assert  func(*testing.T, *app.Config, *app.App)
	}{
		{
			name: "should fail with a nil logger",
			args: args{
				cfg: app.Config{
					Log: nil,
				},
			},
			wantErr: fmt.Errorf("log instance is required"),
		},
		{
			name: "should fail with a nil session repo",
			args: args{
				cfg: app.Config{
					Log:         log.New(log.Config{}),
					SessionRepo: nil,
				},
			},
			wantErr: fmt.Errorf("session repo is required"),
		},
		{
			name: "should fail with a nil file repo",
			args: args{
				cfg: app.Config{
					Log:         log.New(log.Config{}),
					SessionRepo: &StubSessionRepo{},
					FileRepo:    nil,
				},
			},
			wantErr: fmt.Errorf("file repo is required"),
		},
		{
			name: "should fail with a nil password hasher",
			args: args{
				cfg: app.Config{
					Log:            log.New(log.Config{}),
					SessionRepo:    &StubSessionRepo{},
					FileRepo:       &StubFileRepo{},
					PasswordHasher: nil,
				},
			},
			wantErr: fmt.Errorf("password hasher is required"),
		},
		{
			name: "should succeed without an admin user if admin username is empty",
			args: args{
				cfg: app.Config{
					Log:            log.New(log.Config{}),
					SessionRepo:    &StubSessionRepo{},
					FileRepo:       &StubFileRepo{},
					PasswordHasher: &hash.Argon2idDefault,
					AdminUsername:  "",
				},
			},
		},
		{
			name: "should fail if admin password is less than 16 characters",
			args: args{
				cfg: app.Config{
					Log:            log.New(log.Config{}),
					SessionRepo:    &StubSessionRepo{},
					FileRepo:       &StubFileRepo{},
					PasswordHasher: &hash.Argon2idDefault,
					AdminUsername:  "admin",
					AdminPassword:  "123456781234567",
				},
			},
			wantErr: fmt.Errorf("password must be at least 16 characters long"),
		},
		{
			name: "should succeed with an admin password",
			args: args{
				cfg: app.Config{
					Log:            log.New(log.Config{}),
					SessionRepo:    &StubSessionRepo{},
					FileRepo:       &StubFileRepo{},
					PasswordHasher: &hash.Argon2idDefault,
					AdminUsername:  "admin",
					AdminPassword:  "1234567812345678",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.New(ctx, tt.args.cfg)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.assert != nil {
				tt.assert(t, &tt.args.cfg, got)
			}
		})
	}
}
