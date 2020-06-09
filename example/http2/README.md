# freedom
### http2和依赖倒置
---

#### 启动
```sh
$  go run domain/main.go
#  本程序是一个购物的小示例, 浏览器可输入http://127.0.0.1:8000/shop/1, 1-4为商品ID
```

#### http2 server
> main.go
```go

//创建应用
app := freedom.NewApplication()
//安装第三方日志中间件
installLogrus(app)

//创建 http2.0 H2cRunner
h2caddrRunner := app.CreateH2CRunner(config.Get().App.Other["listen_addr"].(string))
app.Run(h2caddrRunner, *config.Get().App)

func installLogrus(app freedom.Application) {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	//安装Logrus后，日志输出全部使用Logrus。
	freedom.Logger().Install(logrus.StandardLogger())
}

// 安装中间件
func installMiddleware(app freedom.Application) {
	/*
		设置框架自带中间件,可重写
		NewTrace默认设置了总线, 下游服务和事件消费者服务都会拿到x-request-id
		NewRequestLogger 默认读取了总线里的x-request-id, 所有上下游服务打印日志全部都携带x-request-id
	*/
	app.InstallMiddleware(middleware.NewTrace("x-request-id"))
	app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))
	

	//http client安装普罗米修斯监控
	requests.InstallPrometheus(config.Get().App.Other["service_name"].(string), freedom.Prometheus())
	app.InstallBusMiddleware(newBus(config.Get().App.Other["service_name"].(string)))
}

// newBus 自定义总线中间件示例.
func newBus(serviceName string) func(freedom.Worker) {
	//调用下游服务和事件消费者将传递service-name， 下游服务和mq事件消费者，使用 Worker.Bus() 可获取到service-name。
	return func(run freedom.Worker) {
		bus := run.Bus()
		bus.Add("x-service-name", serviceName)
	}
}

```

#### 依赖倒置
``` go
package repositorys
//声明接口
type GoodsInterface interface {
	GetGoods(goodsID int) objects.GoodsModel
}
```
```go
//使用接口
func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *ShopService {
			return &ShopService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *ShopService) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// ShopService .
type ShopService struct {
	Worker freedom.Worker
	Goods  repositorys.GoodsInterface //注入接口对象
}

func (s *ShopService) Shopping(goodsID int) string {
	//使用接口的方法
	entity := s.Goods.GetGoods(goodsID)
	return entity.Name
}
```

#### 实现接口和http2 client
- http NewHttpRequest()
- http2 NewH2CRequest()
```go
func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *GoodsRepository {
			return &GoodsRepository{}
		})
	})
}

// GoodsRepository .
type GoodsRepository struct {
	freedom.Repository
}

// 实现接口
func (repo *GoodsRepository) GetGoods(goodsID int) (result objects.GoodsModel) {
	//篇幅有限 不调用其他微服务API，直接调用自身的API
	addr := "http://127.0.0.1:8000/goods/" + strconv.Itoa(goodsID)
	repo.NewH2CRequest(addr).Get().ToJSON(&result)

	//开启go 并发,并且没有group wait。请求结束触发相关对象回收，会快于当前并发go的读取数据，所以使用DeferRecycle
	repo.Worker.DeferRecycle()
	go func() {
		var model objects.GoodsModel
		repo.NewH2CRequest(addr).Get().ToJSON(&model)
		repo.NewHttpRequest(addr, false).Get().ToJSON(&model)
	}()
	return result
}
```