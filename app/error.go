package app

import "fmt"

type (
	Error struct {
		Type   ErrorType
		Title  string
		Detail string
		Err    error
		Status int
	}

	ErrorType string
)

func (e *Error) Error() string {
	var moreDetail string
	if e.Status != 0 || e.Title != "" || e.Detail != "" {
		moreDetail = fmt.Sprintf(" (%d %s: %s)", e.Status, e.Title, e.Detail)
	}
	return fmt.Sprintf("%s: %s%s", e.Type, e.Err.Error(), moreDetail)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Err(t ErrorType, err error) *Error {
	return &Error{
		Type: t,
		Err:  err,
	}
}

func ErrUser(title, detail string, err error) *Error {
	if err == nil {
		err = fmt.Errorf("%s: %s", title, detail)
	}
	return &Error{
		Type:   ErrBadRequest,
		Title:  title,
		Detail: detail,
		Err:    err,
	}
}

func (e *Error) AddDetail(detail string) *Error {
	e.Detail = detail
	return e
}

func (e *Error) AddStatus(status int) *Error {
	e.Status = status
	return e
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

var ErrFileNotFound = fmt.Errorf("file not found")
