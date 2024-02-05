package featureflag

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type (
	FeatureFlags struct {
		mu         sync.RWMutex
		features   map[Name]bool
		defaultVal bool
	}

	Name string

	ConfigOption func(*FeatureFlags) error

	CheckConfig struct{}

	CheckOption func(*CheckConfig)
)

const defaultKey = "DEFAULT"

func WithFeature(name string, value bool) ConfigOption {
	return func(ff *FeatureFlags) error {
		return ff.register(Name(name), value)
	}
}

func WithFeaturesFromEnvironment(prefix string) ConfigOption {
	return func(ff *FeatureFlags) error {
		for _, pair := range os.Environ() {
			if !strings.HasPrefix(pair, prefix) {
				continue
			}
			split := strings.Split(pair, "=")
			if len(split) != 2 {
				return fmt.Errorf("invalid feature flag environment variable split, expecting 2 parts separated by =, but got %d", len(split))
			}
			key := strings.TrimPrefix(split[0], prefix)
			value, err := strconv.ParseBool(split[1])
			if err != nil {
				return fmt.Errorf("parsing boolean feature flag value for env var %q: %q is not a valid boolean value", split[0], split[1])
			}
			if key == defaultKey {
				ff.setDefault(value)
				continue
			}
			if err = ff.register(Name(key), value); err != nil {
				return fmt.Errorf("registering feature flag %q: %w", key, err)
			}
		}
		return nil
	}
}

func New(opts ...ConfigOption) (*FeatureFlags, error) {
	ff := FeatureFlags{
		features: make(map[Name]bool),
	}
	for _, opt := range opts {
		if err := opt(&ff); err != nil {
			return nil, err
		}
	}

	return &ff, nil
}

func (ff *FeatureFlags) register(feature Name, value bool) error {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	if _, ok := ff.features[feature]; ok {
		return fmt.Errorf("feature %q already registered", feature)
	}
	ff.features[feature] = value
	return nil
}

func (ff *FeatureFlags) setDefault(value bool) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	ff.defaultVal = value
}

func (ff *FeatureFlags) get(feature Name) bool {
	ff.mu.RLock()
	defer ff.mu.RUnlock()
	if val, ok := ff.features[feature]; ok {
		return val
	}
	return ff.defaultVal
}

func (ff *FeatureFlags) IsOn(_ context.Context, feature string, opts ...CheckOption) (bool, error) {
	cc := CheckConfig{}
	for _, opt := range opts {
		opt(&cc)
	}
	name := Name(feature)

	if name == "" {
		return false, fmt.Errorf("feature name is required")
	}
	return ff.get(name), nil
}
