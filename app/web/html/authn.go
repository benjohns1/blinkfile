package html

import (
	"fmt"
	"git.jfam.app/one-way-file-send/app/web"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
)

type LoginView struct {
	LayoutView
	SuccessMessage string
	ErrorView
}

func showLogin(ctx iris.Context, _ App) error {
	if isAuthenticated(ctx) {
		ctx.Redirect("/")
		return nil
	}
	ctx.ViewData("content", LoginView{})
	return ctx.View("login.html")
}

func isAuthenticated(ctx iris.Context) bool {
	sess := sessions.Get(ctx)
	if sess == nil {
		return false
	}
	return sess.GetBooleanDefault("authenticated", false)
}

func loginRequired(ctx iris.Context, _ App) error {
	if !isAuthenticated(ctx) {
		ctx.Redirect("/login")
		return nil
	}
	ctx.Next()
	return nil
}

func setAuthenticated(ctx iris.Context, value bool) error {
	sess := sessions.Get(ctx)
	if sess == nil {
		return fmt.Errorf("unable to get session from request context")
	}
	sess.Set("authenticated", value)
	return nil
}

func logout(ctx iris.Context, _ App) error {
	if sessErr := setAuthenticated(ctx, false); sessErr != nil {
		return sessErr
	}
	ctx.ViewData("content", LoginView{
		SuccessMessage: "Successfully logged out",
	})
	return ctx.View("login.html")
}

func doLogin(ctx iris.Context, a App) error {
	view, err := login(ctx, a)
	if err != nil {
		errID, errStatus, errMsg := web.ParseAppErr(err)
		web.LogError(ctx, errID, err)
		view.ErrorView = ErrorView{
			ID:      errID,
			Status:  errStatus,
			Message: errMsg,
		}
	}
	ctx.ViewData("content", view)
	return ctx.View("login.html")
}

func login(ctx iris.Context, a App) (LoginView, error) {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")
	if err := a.Login(ctx, username, password); err != nil {
		return LoginView{}, err
	}
	if sessErr := setAuthenticated(ctx, true); sessErr != nil {
		return LoginView{}, sessErr
	}
	ctx.Redirect("/")

	return LoginView{}, nil
}
