package main

import (
	"time"

	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/infra-example/business/controllers"
	"github.com/8treenet/freedom/example/infra-example/infra/config"
	"github.com/8treenet/freedom/middleware"
	"github.com/8treenet/gcache"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kataras/iris"
)

func main() {
	app := freedom.NewApplication()
	installDatabase(app)

	installMiddleware(app)
	addrRunner := iris.Addr(config.Get().App.Other["listen_addr"].(string))
	app.Run(addrRunner, *config.Get().App)
}

func installMiddleware(app freedom.Application) {
	app.InstallMiddleware(middleware.NewTrace("TRACE-ID"))
	app.InstallMiddleware(middleware.NewLogger("TRACE-ID", true))
	app.InstallMiddleware(middleware.NewRuntimeLogger("TRACE-ID"))
}

func installDatabase(app freedom.Application) {
	app.InstallGorm(func() (db *gorm.DB, cache gcache.Plugin) {
		conf := config.Get().DB
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
