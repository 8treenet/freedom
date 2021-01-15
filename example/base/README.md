# Freedom
## 基础示例

#### 目录结构

- domain - 领域模型
    - aggregate - 聚合
    - entity - 实体
    - event - 领域事件
    - vo - 值对象
    - po - 持久化对象
    - *.go - 领域服务

- adapter - 端口适配器
    - controller - 输入适配器
    - repository - 输出适配器

- server - 服务端程序入口
    - conf - 配置文件
    - main.go - 主函数

- infra - 基础设施组件


---
#### 接口介绍
```go
// main 应用安装接口
type Application interface {
    //安装DB 
    InstallDB(f func() interface{})
    //安装redis
    InstallRedis(f func() (client redis.Cmdable))
    //安装路由中间件
    InstallMiddleware(handler iris.Handler)
    //安装总线中间件,参考Http2 example
    InstallBusMiddleware(handle ...BusHandler)
    //安装全局Party http://domian/relativePath/controllerParty
    InstallParty(relativePath string)
    //创建 Runner
    NewRunner(addr string, configurators ...host.Configurator) iris.Runner
    NewH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
    NewAutoTLSRunner(addr string, domain string, email string, configurators ...host.Configurator) iris.Runner
    NewTLSRunner(addr string, certFile, keyFile string, configurators ...host.Configurator) iris.Runner
    //返回iris应用
    Iris() *iris.Application
    //日志
    Logger() *golog.Logger
    //启动
    Run(serve iris.Runner, c iris.Configuration)
    //安装其他, 如mongodb、es 等
    InstallOther(f func() interface{})
    //启动回调: Prepare之后，Run之前.
    Start(f func(starter Starter))
    //安装序列化，未安装默认使用官方json
    InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
}

 /*
    Worker 请求运行时对象，一个请求创建一个运行时对象，可以注入到controller、service、repository。
 */
type Worker interface {
    //获取iris的上下文
    IrisContext() freedom.Context
    //获取带上下文的日志实例。
    Logger() Logger
    //获取一级缓存实例，请求结束，该缓存生命周期结束。
    Store() *memstore.Store
    //获取总线，读写上下游透传的数据
    Bus() *Bus
    //获取标准上下文
    Context() stdContext.Context
    //With标准上下文
    WithContext(stdContext.Context)
    //该worker起始的时间
    StartTime() time.Time
    //延迟回收对象
    DelayReclaiming()
}

// Initiator 实例初始化接口，在Prepare使用。
type Initiator interface {
    //创建 iris.Party，可以指定中间件。
    CreateParty(relativePath string, handlers ...context.Handler) iris.Party
    //绑定控制器到 iris.Party
    BindControllerByParty(party iris.Party, controller interface{})
    //直接绑定控制器到路径，可以指定中间件。
    BindController(relativePath string, controller interface{}, handlers ...context.Handler)
    //绑定服务
    BindService(f interface{})
    //获取服务
    GetService(ctx iris.Context, service interface{})
    //注入到控制器
    InjectController(f interface{})
    //绑定工厂
    BindFactory(f interface{})
    //绑定Repo
    BindRepository(f interface{})
    //绑定基础设施组件 如果组件是单例 com是对象， 如果组件是多例com是创建函数。
    BindInfra(single bool, com interface{})
    //获取基础设施组件, 只有控制器获取组件需要在Prepare内调用， service和repository可直接依赖注入
    GetInfra(ctx iris.Context, com interface{})
    //启动回调: Prepare之后，Run之前.
    Start(f func(starter Starter))
    Iris() *iris.Application
}

```

---
#### 应用生命周期
|外部使用 | 内部调用 |
| ----- | :---: |
|Application.InstallMiddleware| 注册安装的全局中间件|
|Application.InstallDB|触发回调|
|freedom.Prepare|触发回调|
|Initiator.Starter|触发回调|
|infra.Booting|触发组件方法|
|http.Run|开启监听服务|
|infra.RegisterShutdown|触发回调|
|Application.Close|程序关闭|

#### 请求生命周期
###### 每一个请求开始都会创建若干依赖对象，worker、controller、service、repository、infra等。每一个请求独立使用这些对象，不会多请求并发的读写共享对象。当然也无需担心效率问题，框架已经做了池。请求结束会回收这些对象。 如果过程中使用了go func(){//访问相关对象}，请在之前调用 **Worker.DelayReclaiming()**.

---
#### main
```go

import (
    "github.com/8treenet/freedom"
    _ "github.com/8treenet/freedom/example/base/adapter/controller" //引入输入适配器 http路由
    _ "github.com/8treenet/freedom/example/base/adapter/repository" //引入输出适配器 repository资源库
    "github.com/8treenet/freedom/infra/requests"
    "github.com/8treenet/freedom/middleware"
)

func main() {
    app := freedom.NewApplication() //创建应用
    installDatabase(app)
    installRedis(app)
    installMiddleware(app)

    //创建http 监听
    addrRunner := app.NewRunner(conf.Get().App.Other["listen_addr"].(string))
    //创建http2.0 h2c 监听
    addrRunner = app.NewH2CRunner(conf.Get().App.Other["listen_addr"].(string))
    app.Run(addrRunner, *conf.Get().App)
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
}

func installDatabase(app freedom.Application) {
    app.InstallDB(func() interface{} {
        //安装db的回调函数
        conf := conf.Get().DB
        db, e := gorm.Open("mysql", conf.Addr)
        if e != nil {
            freedom.Logger().Fatal(e.Error())
        }
        return db
    })
}

func installRedis(app freedom.Application) {
    app.InstallRedis(func() (client redis.Cmdable) {
        cfg := conf.Get().Redis
        opt := &redis.Options{
            Addr:               cfg.Addr,
        }
        redisClient := redis.NewClient(opt)
        if e := redisClient.Ping().Err(); e != nil {
            freedom.Logger().Fatal(e.Error())
        }
        client = redisClient
        return
    })
}
```
---
#### controllers/default.go
##### [iris路由文档](https://github.com/kataras/iris/wiki/MVC)
```go
package controller

import (
	"github.com/8treenet/freedom/example/base/domain"
	"github.com/8treenet/freedom/example/base/infra"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		/*
		   普通方式绑定 Default控制器到路径 /
		   initiator.BindController("/", &DefaultController{})
		*/

		//中间件方式绑定， 只对本控制器生效，全局中间件请在main加入。
		initiator.BindController("/", &Default{}, func(ctx freedom.Context) {
			worker := freedom.ToWorker(ctx)
			worker.Logger().Info("Hello middleware begin")
			ctx.Next()
			worker.Logger().Info("Hello middleware end")
		})
	})
}

type Default struct {
	Sev    *domain.Default //注入领域服务 Default
	Worker freedom.Worker //注入请求运行时 Worker
}

// Get handles the GET: / route.
func (c *Default) Get() freedom.Result {
    c.Worker.Logger().Infof("我是控制器")
    remote := c.Sev.RemoteInfo() //调用服务方法
    //返回JSON对象
    return &infra.JSONResponse{Object: remote}
}

// GetHello handles the GET: /hello route.
func (c *Default) GetHello() string {
	return "hello"
}

// PutHello handles the PUT: /hello route.
func (c *Default) PutHello() freedom.Result {
	return &infra.JSONResponse{Object: "putHello"}
}

// PostHello handles the POST: /hello route.
func (c *Default) PostHello() freedom.Result {
	return &infra.JSONResponse{Object: "postHello"}
}

func (m *Default) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("ANY", "/custom", "CustomHello")
	//b.Handle("GET", "/custom", "CustomHello")
	//b.Handle("PUT", "/custom", "CustomHello")
	//b.Handle("POST", "/custom", "CustomHello")
}

// PostHello handles the POST: /hello route.
func (c *Default) CustomHello() freedom.Result {
	method := c.Worker.IrisContext().Request().Method
	c.Worker.Logger().Info("CustomHello", freedom.LogFields{"method": method})
	return &infra.JSONResponse{Object: method + "CustomHello"}
}

// GetUserBy handles the GET: /user/{username:string} route.
func (c *Default) GetUserBy(username string) string {
	return username
}

// GetAgeByUserBy handles the GET: /age/{age:int}/user/{user:string} route.
func (c *Default) GetAgeByUserBy(age int, user string) freedom.Result {
	var result struct {
		User string
		Age  int
	}
	result.Age = age
	result.User = user

	return &infra.JSONResponse{Object: result}
}
```

#### 领域服务 domain/default.go
```go
package domain

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/adapter/repository"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
            //绑定 Default Service
            initiator.BindService(func() *Default {
                return &Default{}
            })
            initiator.InjectController(func(ctx freedom.Context) (service *Default) {
                //把 Default 注入到控制器
                initiator.GetService(ctx, &service)
                return
            })
	})
}

// Default .
type Default struct {
	Worker    freedom.Worker    //注入请求运行时
	DefRepo   *repository.Default   //注入资源库对象  DI方式
	DefRepoIF repository.DefaultRepoInterface  //也可以注入资源库接口 DIP方式
}

// RemoteInfo .
func (s *Default) RemoteInfo() (result struct {
	Ip string
	Ua string
}) {
        s.Worker.Logger().Infof("我是service")
        //调用资源库的方法
        result.Ip = s.DefRepo.GetIP()
        result.Ua = s.DefRepoIF.GetUA()
        return
}
```

#### repositorys/default.go
```go
import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
            //绑定 Default Repository
            initiator.BindRepository(func() *Default {
                return &Default{}
            })
	})
}

// Default .
type Default struct {
    //继承 freedom.Repository
    freedom.Repository
}

// GetIP .
func (repo *Default) GetIP() string {
    //只有继承资源库后才有DB、Redis、NewHttp、Other 访问权限, 并且可以直接获取 Worker
    repo.FetchDB(&db)
    repo.FetchSourceDB(&db)
    repo.Redis()
    repo.Other()
    repo.NewHttpRequest()
    repo.NewH2CRequest()
    repo.Worker().Logger().Infof("我是Repository GetIP")
    return repo.Worker().IrisContext().RemoteAddr()
}

```

#### 配置文件

|文件 | 作用 |
| ----- | :---: |
|server/conf/app.toml|服务配置|
|server/conf/db.toml|db配置|
|server/conf/redis.toml|缓存配置|
|server/conf/infra/.toml|组件相关配置|
