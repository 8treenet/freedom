package main

import (
	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/base/application/controllers"
	"github.com/8treenet/freedom/example/base/infra/config"
	"github.com/8treenet/freedom/middleware"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kataras/iris"
)

func main() {
	app := freedom.NewApplication()
	installMiddleware(app)
	addrRunner := iris.Addr(config.Get().App.Other["listen_addr"].(string))
	app.Run(addrRunner, *config.Get().App)
}

func installMiddleware(app freedom.Application) {
	app.InstallMiddleware(middleware.NewTrace("TRACE-ID"))
	app.InstallMiddleware(middleware.NewLogger("TRACE-ID", true))
	app.InstallMiddleware(middleware.NewRuntimeLogger("TRACE-ID"))
}
