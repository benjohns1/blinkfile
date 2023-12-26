package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestAddLongDuration(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		ld      LongDuration
		args    args
		want    time.Time
		wantErr error
	}{
		{
			name: "should add zero hours",
			ld:   "0h",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(1, 0),
		},
		{
			name: "should add zero days",
			ld:   "0d",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(1, 0),
		},
		{
			name: "should add zero weeks",
			ld:   "0w",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(1, 0),
		},
		{
			name: "should add zero",
			ld:   "0",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(1, 0),
		},
		{
			name: "should add one minute",
			ld:   "1m",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(60+1, 0),
		},
		{
			name: "should add one hour",
			ld:   "1h",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(60*60+1, 0),
		},
		{
			name: "should add one day",
			ld:   "1d",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(60*60*24+1, 0),
		},
		{
			name: "should add one week",
			ld:   "1w",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix(60*60*24*7+1, 0),
		},
		{
			name: "should add one and 1/2 hours",
			ld:   "1.5h",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix((60*60)+(30*60)+1, 0),
		},
		{
			name: "should add one and 1/2 days",
			ld:   "1.5d",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix((60*60*24)+(30*60*24)+1, 0),
		},
		{
			name: "should add one and 1/2 weeks",
			ld:   "1.5w",
			args: args{
				t: time.Unix(1, 0),
			},
			want: time.Unix((60*60*24*7)+(30*60*24*7)+1, 0),
		},
		{
			name: "should subtract 1/2 hours",
			ld:   "-0.5h",
			args: args{
				t: time.Unix(30*60, 0),
			},
			want: time.Unix(0, 0),
		},
		{
			name: "should subtract 1/2 days",
			ld:   "-0.5d",
			args: args{
				t: time.Unix(30*60*24, 0),
			},
			want: time.Unix(0, 0),
		},
		{
			name: "should subtract 1/2 weeks",
			ld:   "-0.5w",
			args: args{
				t: time.Unix(30*60*24*7, 0),
			},
			want: time.Unix(0, 0),
		},
		{
			name: "should add 13.01789 weeks",
			ld:   "13.01789w",
			args: args{
				t: time.Unix(0, 0),
			},
			want: time.Unix(13*60*60*24*7, 17_890_000*60*60*24*7),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ld.AddTo(tt.args.t)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("AddLongDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddLongDuration() got = %v, want %v", got, tt.want)
			}
		})
	}
}
