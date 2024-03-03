package web

import (
	"fmt"

	"github.com/benjohns1/blinkfile"

	"github.com/benjohns1/blinkfile/app"

	"github.com/kataras/iris/v12"
)

type (
	UsersView struct {
		LayoutView
		Users []UserView
		MessageView
	}

	UserView struct {
		ID       string
		Username string
	}
)

func showUsers(ctx iris.Context, a App) error {
	users, err := a.ListUsers(ctx)
	if err != nil {
		return err
	}
	userList := make([]UserView, 0, len(users))
	for _, user := range users {
		userList = append(userList, userToView(user))
	}
	ctx.ViewData("content", UsersView{
		Users:       userList,
		MessageView: flashMessageView(ctx),
	})
	return ctx.View("users.html")
}

func userToView(u blinkfile.User) UserView {
	return UserView{
		ID:       string(u.ID),
		Username: string(u.Username),
	}
}

func createUser(ctx iris.Context, a App) error {
	username, err := doCreateUser(ctx, a)
	if err != nil {
		setFlashErr(ctx, a, err)
	} else {
		setFlashSuccess(ctx, fmt.Sprintf("Created new user %q", username))
	}
	ctx.Redirect("/users")
	return nil
}

func doCreateUser(ctx iris.Context, a App) (string, error) {
	user, pass := ctx.FormValue("username"), ctx.FormValue("password")
	err := a.CreateUser(ctx, app.CreateUserArgs{Username: user, Password: pass})
	if err != nil {
		return "", err
	}
	return user, nil
}
