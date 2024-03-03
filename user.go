package blinkfile

import (
	"fmt"
	"time"
)

type (
	Username string
	UserID   string

	User struct {
		ID UserID
		Username
		Created time.Time
	}
)

var (
	ErrEmptyUserID     = fmt.Errorf("user ID cannot be empty")
	ErrEmptyUsername   = fmt.Errorf("user name cannot be empty")
	ErrEmptyNowService = fmt.Errorf("now() service cannot be empty")
)

func CreateUser(id UserID, name Username, now func() time.Time) (User, error) {
	if id == "" {
		return User{}, ErrEmptyUserID
	}
	if name == "" {
		return User{}, ErrEmptyUsername
	}
	if now == nil {
		return User{}, ErrEmptyNowService
	}
	return User{
		ID:       id,
		Username: name,
		Created:  now(),
	}, nil
}
