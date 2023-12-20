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

const (
	ErrUnknown     ErrorType = "unknown"
	ErrBadRequest  ErrorType = "bad-request"
	ErrInternal    ErrorType = "internal"
	ErrAuthnFailed ErrorType = "authn-failed"
	ErrRepo        ErrorType = "repo"
)
