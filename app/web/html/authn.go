package html

import (
	"git.jfam.app/one-way-file-send/app/web"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
)

type LoginView struct {
	LayoutView
	ErrorView
	SuccessMessage string
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

func logout(ctx iris.Context, _ App) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	session.setLogout()
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
	session, err := getSession(ctx)
	if err != nil {
		return LoginView{}, err
	}
	username := ctx.FormValue("username")
	session.setUsername(username)
	password := ctx.FormValue("password")
	if err := a.Authenticate(ctx, username, password); err != nil {
		return LoginView{}, err
	}
	session.setAuthenticated()
	ctx.Redirect("/")

	return LoginView{}, nil
}
