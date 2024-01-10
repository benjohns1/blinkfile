package blinkfile_test

import (
	"fmt"
	"github.com/benjohns1/blinkfile"
	"reflect"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	type args struct {
		id   blinkfile.UserID
		name blinkfile.Username
		now  func() time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    blinkfile.User
		wantErr error
	}{
		{
			name: "should fail with empty user ID",
			args: args{
				id: "",
			},
			wantErr: fmt.Errorf("user ID cannot be empty"),
		},
		{
			name: "should fail with empty username",
			args: args{
				id:   "user1",
				name: "",
			},
			wantErr: fmt.Errorf("user name cannot be empty"),
		},
		{
			name: "should fail with a nil now() service",
			args: args{
				id:   "user1",
				name: "user1",
				now:  nil,
			},
			wantErr: fmt.Errorf("now() service cannot be empty"),
		},
		{
			name: "should create a new user",
			args: args{
				id:   "user1",
				name: "user1",
				now:  func() time.Time { return time.Unix(1, 0).UTC() },
			},
			want: blinkfile.User{
				Username: "user1",
				Created:  time.Unix(1, 0).UTC(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := blinkfile.CreateUser(tt.args.id, tt.args.name, tt.args.now)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}
