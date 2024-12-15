package app

import "context"

var FeatureFlagIsOn = func(context.Context, FeatureFlag) bool { return false }

type FeatureFlag string

const (
	FeatureUserAccounts FeatureFlag = "UserAccounts"
	LogAllAuthnCalls    FeatureFlag = "LogAllAuthnCalls"
)
