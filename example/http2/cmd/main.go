package main

import (
	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/http2/business/controllers"
	_ "github.com/8treenet/freedom/example/http2/business/repositorys"
	"github.com/8treenet/freedom/example/http2/components/config"
	"github.com/8treenet/freedom/middleware"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
)

func main() {
	app := freedom.NewApplication()
	installMiddleware(app)
	installLogrus(app)

	//http2 h2c 服务
	h2caddrRunner := app.CreateH2CRunner(config.Get().App.Other["listen_addr"].(string))
	app.Run(h2caddrRunner, *config.Get().App)
}

func installMiddleware(app freedom.Application) {
	app.InstallMiddleware(middleware.NewTrace("TRACE-ID"))
	app.InstallMiddleware(middleware.NewLogger("TRACE-ID"))
	app.InstallMiddleware(middleware.NewRuntimeLogger("TRACE-ID"))
}

func installLogrus(app freedom.Application) {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	app.Logger().Install(logrus.StandardLogger())
}
