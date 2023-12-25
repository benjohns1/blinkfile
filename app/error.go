package app

import "fmt"

type (
	Error struct {
		Type ErrorType
		Err  error
	}

	ErrorType string
)

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Err.Error())
}

func (e Error) Unwrap() error {
	return e.Err
}

const (
	ErrUnknown     ErrorType = "unknown"
	ErrBadRequest  ErrorType = "bad-request"
	ErrInternal    ErrorType = "internal"
	ErrAuthnFailed ErrorType = "authn-failed"
	ErrAuthzFailed ErrorType = "authz-failed"
	ErrRepo        ErrorType = "repo"
	ErrNotFound    ErrorType = "not-found"
)

var ErrFileNotFound = Error{ErrNotFound, fmt.Errorf("file not found")}
