package web

import (
	"fmt"
	"strings"

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
	err := a.CreateUser(ctx, app.CreateUserArgs{Username: blinkfile.Username(user), Password: pass})
	if err != nil {
		return "", err
	}
	return user, nil
}

func deleteUsers(ctx iris.Context, a App) error {
	req := ctx.Request()
	err := req.ParseForm()
	if err != nil {
		return err
	}
	deleteUserIDs := make([]blinkfile.UserID, 0, len(req.Form))
	for name, values := range req.Form {
		if len(values) == 0 || values[0] != "on" {
			continue
		}
		deleteUserIDs = append(deleteUserIDs, blinkfile.UserID(strings.TrimPrefix(name, "select-")))
	}
	if len(deleteUserIDs) > 0 {
		err = a.DeleteUsers(ctx, deleteUserIDs)
		if err != nil {
			setFlashErr(ctx, a, err)
		} else {
			setFlashSuccess(ctx, fmt.Sprintf("Deleted %d users.", len(deleteUserIDs)))
		}
	}

	ctx.Redirect("/users")
	return nil
}
