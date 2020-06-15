package main

import (
	"time"

	"github.com/8treenet/extjson"
	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/fshop/adapter/controller"
	"github.com/8treenet/freedom/example/fshop/server/conf"
	"github.com/8treenet/freedom/infra/kafka" //需要开启 server/conf/infra/kafka.toml open = true
	"github.com/8treenet/freedom/infra/requests"
	"github.com/8treenet/freedom/middleware"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {

	app := freedom.NewApplication()
	installDatabase(app)
	installRedis(app)
	installMiddleware(app)

	//安装领域事件的基础设施
	app.InstallDomainEventInfra(kafka.GetDomainEventInfra())
	addrRunner := app.CreateH2CRunner(conf.Get().App.Other["listen_addr"].(string))
	app.Run(addrRunner, *conf.Get().App)
}

func installMiddleware(app freedom.Application) {
	app.InstallMiddleware(middleware.NewRecover())
	app.InstallMiddleware(middleware.NewTrace("x-request-id"))
	app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))

	app.InstallBusMiddleware(middleware.NewLimiter())
	requests.InstallPrometheus(conf.Get().App.Other["service_name"].(string), freedom.Prometheus())

	//安装序列化和反序列化
	extjson.SetDefaultOption(extjson.ExtJSONEntityOption{
		NamedStyle:       extjson.NamedStyleLowerCamelCase,
		SliceNotNull:     true, //空数组不返回null, 返回[]
		StructPtrNotNull: true, //nil结构体指针不返回null, 返回{}})
	})
	app.InstallSerializer(extjson.Marshal, extjson.Unmarshal)
}

func installDatabase(app freedom.Application) {
	app.InstallGorm(func() (db *gorm.DB) {
		conf := conf.Get().DB
		var e error
		db, e = gorm.Open("mysql", conf.Addr)
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}
		db = db.Debug()

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
		client = redis.NewClient(opt)
		if e := client.Ping().Err(); e != nil {
			freedom.Logger().Fatal(e.Error())
		}
		return
	})
}
