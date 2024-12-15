package blinkfile

import (
	"fmt"
	"strings"
	"time"
)

type (
	Username string
	UserID   string

	User struct {
		ID UserID
		Username
		Created    time.Time
		LastEdited time.Time
	}
)

const MinUsernameLength = 4

var (
	ErrEmptyUserID      = fmt.Errorf("user ID cannot be empty")
	ErrEmptyUsername    = fmt.Errorf("user name cannot be empty")
	ErrSameUsername     = fmt.Errorf("previous and new usernames cannot be the same")
	ErrUsernameTooShort = fmt.Errorf("user name must be at least %d characters long", MinUsernameLength)
	ErrEmptyNowService  = fmt.Errorf("now() service cannot be empty")
)

func CreateUser(id UserID, name Username, now func() time.Time) (User, error) {
	if id == "" {
		return User{}, ErrEmptyUserID
	}
	parsedName, err := parseUsername(name)
	if err != nil {
		return User{}, err
	}
	if now == nil {
		return User{}, ErrEmptyNowService
	}
	return User{
		ID:       id,
		Username: parsedName,
		Created:  now(),
	}, nil
}

func (u *User) ChangeUsername(name Username, now func() time.Time) (User, error) {
	parsedName, err := parseUsername(name)
	if err != nil {
		return User{}, err
	}
	if u.Username == parsedName {
		return User{}, ErrSameUsername
	}
	changedUser := u.copy()
	changedUser.Username = parsedName
	changedUser.LastEdited = now()
	return changedUser, nil
}

func parseUsername(name Username) (Username, error) {
	name = Username(strings.Trim(string(name), " "))
	if name == "" {
		return name, ErrEmptyUsername
	}
	if len(name) < MinUsernameLength {
		return name, ErrUsernameTooShort
	}
	return name, nil
}

func (u *User) copy() User {
	return User{
		ID:         u.ID,
		Username:   u.Username,
		Created:    u.Created,
		LastEdited: u.LastEdited,
	}
}
