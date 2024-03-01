package app

import (
	"context"
)

type (
	CreateUserArgs struct {
		Username string
		Password string
	}

	User struct {
		ID       string
		Username string
	}
)

func (a *App) CreateUser(ctx context.Context, args CreateUserArgs) error {
	return nil
}

func (a *App) ListUsers(ctx context.Context) ([]User, error) {
	return nil, nil
}
