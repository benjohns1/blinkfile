package domain

import (
	"fmt"
	"time"
)

type (
	Username string
	UserID   string

	User struct {
		Username
		Created time.Time
	}
)

func CreateUser(id UserID, name Username, now func() time.Time) (User, error) {
	if id == "" {
		return User{}, fmt.Errorf("user ID cannot be empty")
	}
	if name == "" {
		return User{}, fmt.Errorf("user name cannot be empty")
	}
	if now == nil {
		return User{}, fmt.Errorf("now() service cannot be empty")
	}
	return User{
		Username: name,
		Created:  now(),
	}, nil
}
