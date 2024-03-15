package web

import (
	"fmt"
	"net/http"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"github.com/kataras/iris/v12"
)

type LoginView struct {
	LayoutView
	MessageView
}

const authnTokenCookieName = "token"

func showLogin(ctx iris.Context, a App) error {
	if isAuthenticated(ctx, a) {
		ctx.Redirect("/")
		return nil
	}
	ctx.ViewData("content", LoginView{})
	return ctx.View("login.html")
}

func isAuthenticated(ctx iris.Context, a App) bool {
	authnToken := ctx.GetCookie(authnTokenCookieName)
	if authnToken == "" {
		return false
	}
	userID, isAuthn, err := a.IsAuthenticated(ctx, app.Token(authnToken))
	if err != nil {
		a.Errorf(ctx, "checking authentication state of token %q: %v", authnToken, err)
		return false
	}
	if !isAuthn {
		return false
	}
	if sess, sessErr := getSession(ctx); sessErr == nil {
		sess.setAuthenticated(userID)
	}
	return true
}

func loginRequired(ctx iris.Context, a App) error {
	if !isAuthenticated(ctx, a) {
		ctx.Redirect("/login")
		return nil
	}
	ctx.Next()
	return nil
}

func requirePermission(permission string) func(iris.Context, App) error {
	return func(ctx iris.Context, a App) error {
		sess, err := getSession(ctx)
		if err != nil {
			return err
		}
		if sess.GetBooleanDefault(fmt.Sprintf("permission.%s", permission), false) {
			ctx.Next()
			return nil
		}
		ctx.StopWithStatus(http.StatusNotFound)
		return nil
	}
}

func logout(ctx iris.Context, a App) error {
	authnToken := ctx.GetCookie(authnTokenCookieName)
	ctx.RemoveCookie(authnTokenCookieName)
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	session.setLogout()
	if authnToken != "" {
		err = a.Logout(ctx, app.Token(authnToken))
		if err != nil {
			return fmt.Errorf("logging out: %w", err)
		}
	}
	ctx.ViewData("content", LoginView{
		MessageView: MessageView{
			SuccessMessage: "Successfully logged out",
		},
	})
	return ctx.View("login.html")
}

func login(ctx iris.Context, a App) error {
	view, err := doLogin(ctx, a)
	if err != nil {
		view.ErrorView = ParseAppErr(ctx, a, err)
	}
	ctx.ViewData("content", view)
	return ctx.View("login.html")
}

func doLogin(ctx iris.Context, a App) (LoginView, error) {
	session, err := getSession(ctx)
	if err != nil {
		return LoginView{}, err
	}
	username := blinkfile.Username(ctx.FormValue("username"))
	session.setUsername(username)
	password := ctx.FormValue("password")
	req := ctx.Request()
	data := app.SessionRequestData{
		UserAgent: req.UserAgent(),
		IP:        req.RemoteAddr,
	}
	authn, err := a.Login(ctx, username, password, data)
	if err != nil {
		return LoginView{}, err
	}
	ctx.SetCookie(&http.Cookie{
		Name:     authnTokenCookieName,
		Value:    string(authn.Token),
		Expires:  authn.Expires,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	session.setAuthenticated(authn.UserID)
	ctx.Redirect("/")

	return LoginView{}, nil
}
