package main

import (
	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/http2/adapter/controllers"
	"github.com/8treenet/freedom/example/http2/server/conf"
	"github.com/8treenet/freedom/infra/requests"
	"github.com/8treenet/freedom/middleware"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	app := freedom.NewApplication()
	installMiddleware(app)

	//http2 h2c 服务
	h2caddrRunner := app.NewH2CRunner(conf.Get().App.Other["listen_addr"].(string))
	//app.InstallParty("/http2")
	liveness(app)
	app.Run(h2caddrRunner, *conf.Get().App)
}

func installMiddleware(app freedom.Application) {
	//Recover中间件
	app.InstallMiddleware(middleware.NewRecover())
	//Trace链路中间件
	app.InstallMiddleware(middleware.NewTrace("x-request-id"))
	//日志中间件，每个请求一个logger
	app.InstallMiddleware(middleware.NewRequestLogger("x-request-id"))
	//logRow中间件，每一行日志都会触发回调。如果返回true，将停止中间件遍历回调。
	app.Logger().Handle(middleware.DefaultLogRowHandle)

	//HttpClient 普罗米修斯中间件，监控ClientAPI的请求。
	middle := middleware.NewClientPrometheus(conf.Get().App.Other["service_name"].(string), freedom.Prometheus())
	requests.InstallMiddleware(middle)

	//总线中间件，处理上下游透传的Header
	app.InstallBusMiddleware(middleware.NewBusFilter())
	//自定义
	app.InstallBusMiddleware(newBus(conf.Get().App.Other["service_name"].(string)))
}

// newBus 自定义总线中间件示例.
func newBus(serviceName string) func(freedom.Worker) {
	//调用下游服务和事件消费者将传递service-name， 下游服务和mq事件消费者，使用 Worker.Bus() 可获取到service-name。
	return func(run freedom.Worker) {
		bus := run.Bus()
		bus.Add("x-service-name", serviceName)
	}
}

func liveness(app freedom.Application) {
	app.Iris().Get("/ping", func(ctx freedom.Context) {
		ctx.WriteString("pong")
	})
}
