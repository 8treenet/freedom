# freedom
### http2和依赖倒置
---

#### 启动
```sh
$  go run application/main.go
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
		NewRequestLogger和NewRuntimeLogger 默认读取了总线里的x-request-id, 所有上下游服务打印日志全部都携带x-request-id
	*/
	app.InstallMiddleware(middleware.NewTrace("x-request-id"))
	app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))
	app.InstallMiddleware(middleware.NewRuntimeLogger("x-request-id"))

	//http client安装普罗米修斯监控
	requests.InstallPrometheus(config.Get().App.Other["service_name"].(string), freedom.Prometheus())
	app.InstallBusMiddleware(newBus(config.Get().App.Other["service_name"].(string)))
}

// newBus 自定义总线中间件示例.
func newBus(serviceName string) func(freedom.Runtime) {
	//调用下游服务和事件消费者将传递service-name， 下游服务和mq事件消费者，使用 Runtime.Bus() 可获取到service-name。
	return func(run freedom.Runtime) {
		bus := run.Bus()
		bus.Add("x-service-name", serviceName)
	}
}

```

#### 依赖倒置
###### 依赖倒置的本质是以 依赖注入 + 接口隔离 的方式隔离 实现和调用依赖，可以灵活的应对多变的业务和多人协作。
```go
/*
	购买的接口定义 
	ShoppingInterface 接口声明在 service包， 由controller使用
	文件 application/interface.go
*/
type ShoppingInterface interface {
	Shopping(goodsID int) string
}

/*
	文件 controllers/shop.go
*/
type ShopController struct {
	Shopping services.ShoppingInterface	//定义接口变量, 框架会自动注入该接口的实现
}

// Get handles the GET: /shop/:id route.
func (s *ShopController) GetBy(id int) string {
	//该接口的使用 s.Shopping
	return s.Shopping.Shopping(id)
}

/*
	商品的接口定义
	Goods 接口声明在 repositorys包， 由service使用
	文件 repositorys/interface.go
*/
type GoodsInterface interface {
	GetGoods(goodsID int) objects.GoodsModel
}

/*
	文件 application/shop.go
*/
type ShopService struct {
	Goods   repositorys.GoodsInterface //定义接口变量, 框架会自动注入该接口的实现
}

// Shopping implment Shopping interface
func (s *ShopService) Shopping(goodsID int) string {
	//该接口的使用 s.Goods
	entity := s.Goods.GetGoods(goodsID)
	return fmt.Sprintf("您花费了%d元, 购买了一件 %s。", entity.Price, entity.Name)
}
```

#### http2 client
```go
/*
	Repo 数据源访问
	1.实例必须继承 freedom.Repository
	2.创建http请求 .NewHttpRequest()
	3.创建http2.0请求 .NewH2CRequest()
*/
type GoodsRepository struct {
	freedom.Repository
}

// GetGoods implment Goods interface
func (repo *GoodsRepository) GetGoods(goodsID int) (result objects.GoodsModel) {
	//篇幅所限，示例直接调用自身服务的其他http接口，而不是下游。
	addr := "http://127.0.0.1:8000/goods/" + strconv.Itoa(goodsID)
	repo.NewH2CRequest(addr).Get().ToJSON(&result)

	correctReq := repo.NewH2CRequest(addr)
	go func() {
		//panic 错误方式, repo.NewH2CRequest方法使用了Runtime.Bus(),用来设置传递给下游的总线相关数据，并发中请求运行时结束会导致panic
		var model objects.GoodsModel
		repo.NewH2CRequest(addr).Get().ToJSON(&model)

		//正确方式1, 在请求内创建
		correctReq.Get().ToJSON(&model)

		//正确方式2, 不向下游传递trace相关数据,不会访问请求运行时。通常访问外部服务设置false。
		repo.NewH2CRequest(addr, false).Get().ToJSON(&model)
	}()
	return result
}

```