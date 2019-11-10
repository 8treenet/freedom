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
	installMiddleware()

	//http2 h2c 服务
	h2caddrRunner := freedom.CreateH2CRunner(config.Get().App.Other["listen_addr"].(string))
	freedom.Run(h2caddrRunner, config.Get().App)
}

func installMiddleware() {
	freedom.UseMiddleware(middleware.NewTrace("TRACE-ID"))
	freedom.UseMiddleware(middleware.NewLogger("TRACE-ID"))
	freedom.UseMiddleware(middleware.NewRuntimeLogger("TRACE-ID"))
	installLogrus()
}

func installLogrus() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	freedom.Logger().Install(logrus.StandardLogger())
}
