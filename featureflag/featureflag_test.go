package featureflag_test

import (
	"context"
	"fmt"
	"github.com/benjohns1/blinkfile/featureflag"
	"os"
	"reflect"
	"strings"
	"testing"
)

func newTestFeatureFlags(t *testing.T, opts ...featureflag.ConfigOption) *featureflag.FeatureFlags {
	ff, err := featureflag.New(opts...)
	if err != nil {
		t.Fatal(err)
	}
	return ff
}

func TestFeatureFlags_IsOn(t *testing.T) {
	ctx := context.Background()
	type args struct {
		feature string
		opts    []featureflag.CheckOption
	}
	tests := []struct {
		name    string
		ff      *featureflag.FeatureFlags
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "should fail if feature name is empty",
			args: args{
				feature: "",
			},
			wantErr: fmt.Errorf("feature name is required"),
		},
		{
			name: "should return false if feature is not known",
			args: args{
				feature: "feature-a",
			},
			want: false,
		},
		{
			name: "should return false if feature flag is not set",
			ff: func() *featureflag.FeatureFlags {
				ff := newTestFeatureFlags(t, featureflag.WithFeature("feature-a", false))
				return ff
			}(),
			args: args{
				feature: "feature-a",
			},
			want: false,
		},
		{
			name: "should return true if feature flag is set",
			ff: func() *featureflag.FeatureFlags {
				ff := newTestFeatureFlags(t, featureflag.WithFeature("feature-a", true))
				return ff
			}(),
			args: args{
				feature: "feature-a",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ff == nil {
				tt.ff = &featureflag.FeatureFlags{}
			}
			got, err := tt.ff.IsOn(ctx, tt.args.feature, tt.args.opts...)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("IsOn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsOn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func cleanEnvVars(t *testing.T, prefix string) func() {
	t.Helper()
	saved := make(map[string]string)
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		s := strings.Split(env, "=")
		k, v := s[0], s[1]
		saved[k] = v
		if err := os.Unsetenv(k); err != nil {
			t.Fatal(err)
		}
	}
	return func() {
		for k, v := range saved {
			if err := os.Setenv(k, v); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func anyFatalErr(t *testing.T, errs ...error) {
	t.Helper()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func Test_New(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		arrange    func(*testing.T) func()
		opts       []featureflag.ConfigOption
		wantErr    error
		wantValues map[string]bool
	}{
		{
			name: "should fail loading from env vars with an invalid boolean value",
			arrange: func(t *testing.T) func() {
				restore := cleanEnvVars(t, "TEST_FEATURE_FLAG_")
				anyFatalErr(t, os.Setenv("TEST_FEATURE_FLAG_feature-a", "123"))
				return restore
			},
			opts: []featureflag.ConfigOption{
				featureflag.WithFeaturesFromEnvironment("TEST_FEATURE_FLAG_"),
			},
			wantErr: fmt.Errorf(`parsing boolean feature flag value for env var "TEST_FEATURE_FLAG_feature-a": "123" is not a valid boolean value`),
		},
		{
			name: "should fail loading from env vars if attempting to re-register a feature flag with a duplicate name",
			arrange: func(t *testing.T) func() {
				restore := cleanEnvVars(t, "TEST_FEATURE_FLAG_")
				anyFatalErr(t, os.Setenv("TEST_FEATURE_FLAG_feature-a", "1"))
				return restore
			},
			opts: []featureflag.ConfigOption{
				featureflag.WithFeature("feature-a", false),
				featureflag.WithFeaturesFromEnvironment("TEST_FEATURE_FLAG_"),
			},
			wantErr: fmt.Errorf(`registering feature flag "feature-a": %w`, fmt.Errorf(`feature "feature-a" already registered`)),
		},
		{
			name: "should load 1 feature from env vars",
			arrange: func(t *testing.T) func() {
				restore := cleanEnvVars(t, "TEST_FEATURE_FLAG_")
				anyFatalErr(t, os.Setenv("TEST_FEATURE_FLAG_feature-a", "1"))
				return restore
			},
			opts: []featureflag.ConfigOption{
				featureflag.WithFeaturesFromEnvironment("TEST_FEATURE_FLAG_"),
			},
			wantValues: map[string]bool{
				"feature-a": true,
			},
		},
		{
			name: "should load features from env vars - any unregistered features should be false",
			arrange: func(t *testing.T) func() {
				restore := cleanEnvVars(t, "TEST_FEATURE_FLAG_")
				anyFatalErr(t,
					os.Setenv("TEST_FEATURE_FLAG_feature-a", "1"),
					os.Setenv("TEST_FEATURE_FLAG_feature-b", "0"),
					os.Setenv("TEST_FEATURE_FLAG_feature-c", "true"),
					os.Setenv("TEST_FEATURE_FLAG_feature-d", "FALSE"),
					os.Setenv("TEST_FEATURE_FLAG_feature-e", "TRUE"),
				)
				return restore
			},
			opts: []featureflag.ConfigOption{
				featureflag.WithFeaturesFromEnvironment("TEST_FEATURE_FLAG_"),
			},
			wantValues: map[string]bool{
				"feature-a": true,
				"feature-b": false,
				"feature-c": true,
				"feature-d": false,
				"feature-e": true,
				"feature-f": false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.arrange != nil {
				cleanup := tt.arrange(t)
				defer cleanup()
			}
			ff, err := featureflag.New(tt.opts...)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("New() error:\n\t%v\nwantErr:\n\t%v", err, tt.wantErr)
			}
			for feature, want := range tt.wantValues {
				got, _ := ff.IsOn(ctx, feature)
				if got != want {
					t.Errorf("New() got flag %q = %v, want %v", feature, got, want)
				}
			}
		})
	}
}
