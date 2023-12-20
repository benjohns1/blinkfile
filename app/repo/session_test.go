package repo_test

import (
	"context"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/repo"
	"reflect"
	"testing"
)

func TestSession_Save(t *testing.T) {
	type args struct {
		session app.Session
	}
	tests := []struct {
		name    string
		repo    *repo.Session
		args    args
		wantErr error
	}{
		{
			name: "should save a session",
			repo: repo.NewSession(),
			args: args{
				session: app.Session{},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.Save(context.Background(), tt.args.session)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_GetByToken(t *testing.T) {
	ctx := context.Background()
	type args struct {
		token app.Token
	}
	tests := []struct {
		name    string
		repo    *repo.Session
		args    args
		want    app.Session
		wantOK  bool
		wantErr error
	}{
		{
			name: "should not get a non-existent token",
			repo: repo.NewSession(),
			args: args{
				token: "token1",
			},
			wantOK: false,
		},
		{
			name: "should get a token",
			repo: func() *repo.Session {
				r := repo.NewSession()
				if err := r.Save(ctx, app.Session{Token: "token1"}); err != nil {
					t.Fatal(err)
				}
				return r
			}(),
			args: args{
				token: "token1",
			},
			want:   app.Session{Token: "token1"},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK, err := tt.repo.GetByToken(ctx, tt.args.token)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByToken() got = %v, want %v", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("GetByToken() gotOK = %v, wantOK %v", gotOK, tt.wantOK)
			}
		})
	}
}

func TestSession_Delete(t *testing.T) {
	ctx := context.Background()
	type args struct {
		token app.Token
	}
	tests := []struct {
		name    string
		repo    *repo.Session
		args    args
		wantErr error
		assert  func(*testing.T, *repo.Session)
	}{
		{
			name: "should no-op a non-existent token",
			repo: repo.NewSession(),
			args: args{
				token: "token1",
			},
			wantErr: nil,
		},
		{
			name: "should delete a token",
			repo: func() *repo.Session {
				r := repo.NewSession()
				if err := r.Save(ctx, app.Session{Token: "token1"}); err != nil {
					t.Fatal(err)
				}
				return r
			}(),
			args: args{
				token: "token1",
			},
			assert: func(t *testing.T, r *repo.Session) {
				_, ok, _ := r.GetByToken(ctx, "token1")
				if ok {
					t.Errorf("Delete() did not delete token")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.Delete(ctx, tt.args.token)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.assert != nil {
				tt.assert(t, tt.repo)
			}
		})
	}
}
