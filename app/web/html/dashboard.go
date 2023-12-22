package html

import "github.com/kataras/iris/v12"

type DashboardView struct {
	LayoutView
}

func showDashboard(ctx iris.Context, _ App) error {
	ctx.ViewData("content", DashboardView{})
	return ctx.View("dashboard.html")
}
