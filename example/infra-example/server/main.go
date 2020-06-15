package main

import (
	"time"

	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/infra-example/adapter/controllers"
	"github.com/8treenet/freedom/example/infra-example/server/conf"
	"github.com/8treenet/freedom/infra/requests"
	"github.com/8treenet/freedom/middleware"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	app := freedom.NewApplication()
	installDatabase(app)

	installMiddleware(app)
	addrRunner := app.CreateRunner(conf.Get().App.Other["listen_addr"].(string))
	app.Run(addrRunner, *conf.Get().App)
}

func installMiddleware(app freedom.Application) {
	app.InstallMiddleware(middleware.NewRecover())
	app.InstallMiddleware(middleware.NewTrace("x-request-id"))
	app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))

	app.InstallBusMiddleware(middleware.NewLimiter())
	requests.InstallPrometheus(conf.Get().App.Other["service_name"].(string), freedom.Prometheus())
}

func installDatabase(app freedom.Application) {
	app.InstallGorm(func() (db *gorm.DB) {
		conf := conf.Get().DB
		var e error
		db, e = gorm.Open("mysql", conf.Addr)
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}

		db.DB().SetMaxIdleConns(conf.MaxIdleConns)
		db.DB().SetMaxOpenConns(conf.MaxOpenConns)
		db.DB().SetConnMaxLifetime(time.Duration(conf.ConnMaxLifeTime) * time.Second)
		return
	})
}
