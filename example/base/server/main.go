package main

import (
	"time"

	"github.com/8treenet/freedom/example/base/server/conf"

	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/base/adapter/controller"
	"github.com/8treenet/freedom/infra/requests"
	"github.com/8treenet/freedom/middleware"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
)

func main() {
	app := freedom.NewApplication()
	/*
		installDatabase(app) //安装数据库
		installRedis(app) //安装redis
		installLogrus(app) //安装第三方logger

		http2 h2c 服务
		h2caddrRunner := app.CreateH2CRunner(conf.Get().App.Other["listen_addr"].(string))
	*/

	installMiddleware(app)
	addrRunner := app.CreateRunner(conf.Get().App.Other["listen_addr"].(string))
	app.Run(addrRunner, *conf.Get().App)
}

func installMiddleware(app freedom.Application) {
	app.InstallMiddleware(middleware.NewRecover())
	app.InstallMiddleware(middleware.NewTrace("x-request-id"))
	app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))

	requests.InstallPrometheus(conf.Get().App.Other["service_name"].(string), freedom.Prometheus())
	app.InstallBusMiddleware(middleware.NewLimiter())
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

func installRedis(app freedom.Application) {
	app.InstallRedis(func() (client redis.Cmdable) {
		cfg := conf.Get().Redis
		opt := &redis.Options{
			Addr:               cfg.Addr,
			Password:           cfg.Password,
			DB:                 cfg.DB,
			MaxRetries:         cfg.MaxRetries,
			PoolSize:           cfg.PoolSize,
			ReadTimeout:        time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:       time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:        time.Duration(cfg.IdleTimeout) * time.Second,
			IdleCheckFrequency: time.Duration(cfg.IdleCheckFrequency) * time.Second,
			MaxConnAge:         time.Duration(cfg.MaxConnAge) * time.Second,
			PoolTimeout:        time.Duration(cfg.PoolTimeout) * time.Second,
		}
		redisClient := redis.NewClient(opt)
		if e := redisClient.Ping().Err(); e != nil {
			freedom.Logger().Fatal(e.Error())
		}
		client = redisClient
		return
	})
}

func installLogrus(app freedom.Application) {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	freedom.Logger().Install(logrus.StandardLogger())
}
