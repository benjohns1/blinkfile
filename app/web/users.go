package web

import "github.com/kataras/iris/v12"

func showUsers(ctx iris.Context, a App) error {
	return ctx.View("users.html")
}
