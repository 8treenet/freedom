# freedom
### HTTP/H2C和依赖倒置
---

#### 启动
```sh
$  go run server/main.go
#  本程序是一个购物的小示例, 浏览器可输入http://127.0.0.1:8000/shop/1, 1-4为商品ID
```

#### HTTP/H2C Server
> main.go
```go

//创建应用
app := freedom.NewApplication()
//创建 http2.0 H2cRunner
h2caddrRunner := app.NewH2CRunner(conf.Get().App.Other["listen_addr"].(string))
app.Run(h2caddrRunner, *conf.Get().App)

// 安装中间件
func installMiddleware(app freedom.Application) {
	/*
		设置框架自带中间件,可重写
		NewTrace默认设置了总线, 下游服务和事件消费者服务都会拿到x-request-id
		NewRequestLogger 默认读取了总线里的x-request-id, 所有上下游服务打印日志全部都携带x-request-id
	*/
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
	//自定义Bus
	app.InstallBusMiddleware(newBus(conf.Get().App.Other["service_name"].(string)))
}

// newBus 自定义总线中间件示例.
func newBus(serviceName string) func(freedom.Worker) {
	//调用下游服务和事件消费者将传递service-name， 下游服务和mq事件消费者，使用 Worker.Bus() 可获取到service-name。
	return func(run freedom.Worker) {
		bus := run.Bus()
		//Bus 主要处理http的Header，所有下游调用都会携带Header:x-service-name
		bus.Add("x-service-name", serviceName)
	}
}
```
```go
//默认的每行log中间件,支持自定义.
func DefaultLogRowHandle(value *freedom.LogRow) bool {
	//logRow中间件，每一行日志都会触发回调。如果返回true，将停止中间件遍历回调。
	fieldKeys := []string{}
	for k := range value.Fields {
		fieldKeys = append(fieldKeys, k)
	}
	sort.Strings(fieldKeys)
	for i := 0; i < len(fieldKeys); i++ {
		fieldMsg := value.Fields[fieldKeys[i]]
		if value.Message != "" {
			value.Message += "  "
		}
		value.Message += fmt.Sprintf("%s:%v", fieldKeys[i], fieldMsg)
	}
	return false

	/*
		logrus.WithFields(value.Fields).Info(value.Message)
		return true
	*/
	/*
		zapLogger, _ := zap.NewProduction()
		zapLogger.Info(value.Message)
		return true
	*/
}
```

#### 依赖倒置
``` go
package repository
//声明接口
type GoodsInterface interface {
	GetGoods(goodsID int) vo.GoodsModel
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
	Goods  repository.GoodsInterface //注入接口对象
}

func (s *ShopService) Shopping(goodsID int) string {
	//使用接口的方法
	entity := s.Goods.GetGoods(goodsID)
	return entity.Name
}
```

#### 实现接口和HTTP/H2C Client
- HTTP     NewHttpRequest()
- HTTP/H2C NewH2CRequest()
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
func (repo *GoodsRepository) GetGoods(goodsID int) (result vo.GoodsModel) {
	//篇幅有限 不调用其他微服务API，直接调用自身的API
	addr := "http://127.0.0.1:8000/goods/" + strconv.Itoa(goodsID)
	repo.NewH2CRequest(addr).Get().ToJSON(&result)

	//开启go 并发,并且没有group wait。请求结束触发相关对象回收，会快于当前并发go的读取数据，所以使用DeferRecycle
	repo.Worker().DeferRecycle()
	go func() {
		var model vo.GoodsModel
		repo.NewH2CRequest(addr).Get().ToJSON(&model)
		repo.NewHttpRequest(addr, false).Get().ToJSON(&model)
	}()
	return result
}
```