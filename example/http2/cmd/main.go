package main

import (
	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/http2/controllers"
	"github.com/8treenet/freedom/example/http2/models/config"
	_ "github.com/8treenet/freedom/example/http2/repositorys"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
)

func main() {
	install()

	//http2 h2c 服务
	h2caddrRunner := freedom.CreateH2CRunner(config.Get().App.Other["listen_addr"].(string))
	freedom.Run(h2caddrRunner, config.Get().App)
}

func install() {
	installLogrus()
}

func installLogrus() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	freedom.Logger().Install(logrus.StandardLogger())
}
