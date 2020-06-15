package main

import (
	"fmt"
	"time"

	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/event-example/adapter/controllers"
	"github.com/8treenet/freedom/example/event-example/server/conf"
	"github.com/8treenet/freedom/infra/kafka"
	"github.com/8treenet/freedom/infra/requests"
	"github.com/8treenet/freedom/middleware"
	"github.com/Shopify/sarama"
)

// mac 启动kafka: zookeeper-server-start /usr/local/etc/kafka/zookeeper.properties & kafka-server-start /usr/local/etc/kafka/server.properties
func main() {
	//如果使用默认kafka配置，无需设置
	kafka.SettingConfig(func(conf *sarama.Config, other map[string]interface{}) {
		conf.Producer.Retry.Max = 3
		conf.Producer.Retry.Backoff = 5 * time.Second
		conf.Consumer.Offsets.Initial = sarama.OffsetOldest
		fmt.Println(other)
	})
	app := freedom.NewApplication()
	installMiddleware(app)
	addrRunner := app.CreateH2CRunner(conf.Get().App.Other["listen_addr"].(string))

	//获取领域事件的kafka基础设施并安装
	app.InstallDomainEventInfra(kafka.GetDomainEventInfra())
	app.Run(addrRunner, *conf.Get().App)
}

func installMiddleware(app freedom.Application) {
	app.InstallMiddleware(middleware.NewRecover())
	app.InstallMiddleware(middleware.NewTrace("x-request-id"))
	app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))

	app.InstallBusMiddleware(middleware.NewLimiter())
	requests.InstallPrometheus(conf.Get().App.Other["service_name"].(string), freedom.Prometheus())
}
