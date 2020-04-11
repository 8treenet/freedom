package main

import (
	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/http2/adapter/controllers"
	"github.com/8treenet/freedom/example/http2/infra/config"
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
	/*
		设置框架自带中间件,可重写
		NewTrace默认设置了总线, 下游服务和事件消费者服务都会拿到TRACE-ID
		NewLogger和NewRuntimeLogger 默认读取了总线里的TRACE-ID, 所有上下游服务打印日志全部都携带TRACE-ID
	*/
	app.InstallMiddleware(middleware.NewTrace("TRACE-ID"))
	app.InstallMiddleware(middleware.NewLogger("TRACE-ID", true))
	app.InstallMiddleware(middleware.NewRuntimeLogger("TRACE-ID"))

	app.InstallBusMiddleware(newBus(config.Get().App.Other["service_name"].(string)))
}

func installLogrus(app freedom.Application) {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	freedom.Logger().Install(logrus.StandardLogger())
}

// newBus 自定义总线中间件示例.
func newBus(serviceName string) func(freedom.Runtime) {
	//调用下游服务和事件消费者将传递service-name， 下游服务和mq事件消费者，使用 Runtime.Bus() 可获取到service-name。
	return func(run freedom.Runtime) {
		bus := run.Bus()
		bus.Add("service-name", serviceName)
	}
}
