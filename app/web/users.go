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

	EditUserView struct {
		LayoutView
		User UserView
		MessageView
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

func showEditUser(ctx iris.Context, a App) error {
	userID := blinkfile.UserID(ctx.Params().Get("user_id"))
	user, err := a.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	ctx.ViewData("content", EditUserView{
		User:        userToView(user),
		MessageView: flashMessageView(ctx),
	})
	return ctx.View("user_edit.html")
}

func editUser(ctx iris.Context, a App) error {
	userID := blinkfile.UserID(ctx.Params().Get("user_id"))
	successMsg, err := doEditUser(ctx, a, userID)
	if err != nil {
		setFlashErr(ctx, a, err)
	} else {
		setFlashSuccess(ctx, successMsg)
	}
	ctx.Redirect(fmt.Sprintf("/users/%s/edit", userID))
	return nil
}

func doEditUser(ctx iris.Context, a App, userID blinkfile.UserID) (string, error) {
	action := ctx.FormValue("action")
	switch action {
	case "change_username":
		username := ctx.FormValue("username")
		err := a.ChangeUsername(ctx, app.ChangeUsernameArgs{ID: userID, Username: blinkfile.Username(username)})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Username changed to %q", username), nil
	case "change_password":
		password := ctx.FormValue("password")
		username, err := a.ChangePassword(ctx, app.ChangePasswordArgs{ID: userID, Password: password})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Password changed for %q", username), nil
	default:
		return "", fmt.Errorf("unknown action %q", action)
	}
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
			var plural string
			if len(deleteUserIDs) != 1 {
				plural = "s"
			}
			setFlashSuccess(ctx, fmt.Sprintf("Deleted %d user%s.", len(deleteUserIDs), plural))
		}
	}

	ctx.Redirect("/users")
	return nil
}
