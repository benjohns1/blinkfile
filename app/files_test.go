package app_test

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"reflect"
	"testing"
	"time"
)

func TestApp_ListFiles(t *testing.T) {
	ctx := context.Background()
	type args struct {
		owner blinkfile.UserID
	}
	tests := []struct {
		name    string
		cfg     app.Config
		args    args
		want    []blinkfile.FileHeader
		wantErr error
	}{
		{
			name: "should fail if owner is empty",
			args: args{
				"",
			},
			wantErr: &app.Error{
				Type: app.ErrBadRequest,
				Err:  fmt.Errorf("owner is required"),
			},
		},
		{
			name: "should fail with repo error",
			cfg: app.Config{
				FileRepo: &StubFileRepo{
					ListByUserFunc: func(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error) {
						return nil, fmt.Errorf("list file err")
					},
				},
			},
			args: args{
				"user1",
			},
			wantErr: &app.Error{
				Type: app.ErrRepo,
				Err:  fmt.Errorf("retrieving file list: %w", fmt.Errorf("list file err")),
			},
		},
		{
			name: "should sort files by created time descending, then name, then internal file ID",
			cfg: app.Config{
				FileRepo: &StubFileRepo{
					ListByUserFunc: func(context.Context, blinkfile.UserID) ([]blinkfile.FileHeader, error) {
						return []blinkfile.FileHeader{
							{
								ID:      "1",
								Name:    "File1",
								Created: time.Unix(1, 0),
							},
							{
								ID:      "2",
								Name:    "File2",
								Created: time.Unix(2, 0),
							},
							{
								ID:      "3b",
								Name:    "File3b",
								Created: time.Unix(3, 0),
							},
							{
								ID:      "3a",
								Name:    "File3a",
								Created: time.Unix(3, 0),
							},
							{
								ID:      "4",
								Name:    "File4b",
								Created: time.Unix(4, 0),
							},
							{
								ID:      "4",
								Name:    "File4a",
								Created: time.Unix(4, 0),
							},
						}, nil
					},
				},
			},
			args: args{
				"user1",
			},
			want: []blinkfile.FileHeader{
				{
					ID:      "4",
					Name:    "File4a",
					Created: time.Unix(4, 0),
				},
				{
					ID:      "4",
					Name:    "File4b",
					Created: time.Unix(4, 0),
				},
				{
					ID:      "3a",
					Name:    "File3a",
					Created: time.Unix(3, 0),
				},
				{
					ID:      "3b",
					Name:    "File3b",
					Created: time.Unix(3, 0),
				},
				{
					ID:      "2",
					Name:    "File2",
					Created: time.Unix(2, 0),
				},
				{
					ID:      "1",
					Name:    "File1",
					Created: time.Unix(1, 0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AppConfigDefaults(tt.cfg)
			application := NewTestApp(ctx, t, cfg)
			got, err := application.ListFiles(ctx, tt.args.owner)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListFiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}
