# freedom
### base - 基础示例

#### 目录结构

- domain - 领域模型
    - aggregate - 聚合
    - entity - 实体
    - *.go - 领域服务

- adapter - 端口适配器
    - controller - 输入适配器
    - repository - 输出适配器
    - dto - 传输对象
    - po - 持久化对象

- server - 服务端程序入口
    - conf - 配置文件
    - main.go - 主函数

- infra - 基础设施
    - *go - 基础设施组件    


---
#### 接口介绍
```go
// main 应用安装接口
type Application interface {
    //安装gorm 
    InstallGorm(f func() (db *gorm.DB))
    //安装redis
    InstallRedis(f func() (client redis.Cmdable))
    //安装路由中间件
    InstallMiddleware(handler iris.Handler)
    //安装总线中间件,参考Http2 example
    InstallBusMiddleware(handle ...BusHandler)
    //安装全局Party http://domian/relativePath/controllerParty
    InstallParty(relativePath string)
    //创建http h2c
    CreateH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
    //创建http 
    CreateRunner(addr string, configurators ...host.Configurator) iris.Runner
    //返回iris应用
    Iris() *iris.Application
    //日志
    Logger() *golog.Logger
    //启动
    Run(serve iris.Runner, c iris.Configuration)
    //安装领域事件
    InstallDomainEventInfra(eventInfra DomainEventInfra)
    //安装其他, 如mongodb、es 等
    InstallOther(f func() interface{})
    //启动回调: Prepare之后，Run之前.
    Start(f func(starter Starter))
    //安装序列化:应用内的 领域事件对象、实体等序列化和反序列化。未安装默认使用官方json
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
    //注入到控制器
    InjectController(f interface{})
    //绑定Repo
    BindRepository(f interface{})
    //获取服务
    GetService(ctx iris.Context, service interface{})
    //绑定基础设施组件 如果组件是单例 com是对象， 如果组件是多例com是创建函数。
    BindInfra(single bool, com interface{})
    //获取基础设施组件, 只有控制器获取组件需要在Prepare内调用， service和repository可直接依赖注入
    GetInfra(ctx iris.Context, com interface{})
    //启动回调: Prepare之后，Run之前.
    Start(f func(starter Starter))
    Iris() *iris.Application
}

// Starter .
type Starter interface {
    Iris() *iris.Application
    //异步缓存预热
    AsyncCachePreheat(f func(repo *Repository))
    //同步缓存预热
    CachePreheat(f func(repo *Repository))
    //获取单例组件
    GetSingleInfra(com interface{})
}

```

---
#### 应用生命周期
|外部使用 | 内部调用 |
| ----- | :---: |
|Application.InstallMiddleware| 注册安装的全局中间件|
|Application.InstallGorm|触发回调|
|freedom.Prepare|触发回调|
|Initiator.Starter|触发回调|
|infra.Booting|触发组件方法|
|http.Run|开启监听服务|
|infra.Closeing|触发回调|
|Application.Close|程序关闭|

#### 请求生命周期
###### 每一个请求开始都会创建若干依赖对象，worker、controller、service、repository、infra等。每一个请求独立使用这些对象，不会多请求并发的读写共享对象。当然也无需担心效率问题，框架已经做了池。请求结束会回收这些对象。 如果过程中使用了go func(){//访问相关对象}，请在之前调用 **Worker.DelayReclaiming()**.

---
#### main
```go
func main() {
    app := freedom.NewApplication() //创建应用
    installDatabase(app)
    installRedis(app)
    installMiddleware(app)

    //创建http 监听
    addrRunner := app.CreateRunner(conf.Get().App.Other["listen_addr"].(string))
    //创建http2.0 h2c 监听
    addrRunner = app.CreateH2CRunner(conf.Get().App.Other["listen_addr"].(string))
    app.Run(addrRunner, *conf.Get().App)
}

func installMiddleware(app freedom.Application) {
    //安装全局中间件
    app.InstallMiddleware(middleware.NewRecover())
    app.InstallMiddleware(middleware.NewTrace("x-request-id"))
    app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))

    requests.InstallPrometheus(conf.Get().App.Other["service_name"].(string), freedom.Prometheus())
    app.InstallBusMiddleware(middleware.NewLimiter())
}

func installDatabase(app freedom.Application) {
    app.InstallGorm(func() (db *gorm.DB) {
        //安装db的回调函数
        conf := conf.Get().DB
        var e error
        db, e = gorm.Open("mysql", conf.Addr)
        if e != nil {
            freedom.Logger().Fatal(e.Error())
        }
        return
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
	c.Worker.Logger().Info(method, "CustomHello")
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
	DefRepo   *repository.Default   //注入仓库对象  DI方式
	DefRepoIF repository.DefaultRepoInterface  //也可以注入仓库接口 DIP方式
}

// RemoteInfo .
func (s *Default) RemoteInfo() (result struct {
	Ip string
	Ua string
}) {
        s.Worker.Logger().Infof("我是service")
        //调用仓库的方法
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
    //只有继承仓库后才有DB、Redis、NewHttp、Other 访问权限, 并且可以直接获取 Worker
    repo.DB()
    repo.Redis()
    repo.Other()
    repo.NewHttpRequest()
    repo.Worker.Logger().Infof("我是Repository GetIP")
    return repo.Worker.IrisContext().RemoteAddr()
}

```

#### 配置文件

|文件 | 作用 |
| ----- | :---: |
|server/conf/app.toml|服务配置|
|server/conf/db.toml|db配置|
|server/conf/redis.toml|缓存配置|
|server/conf/infra/.toml|组件相关配置|
