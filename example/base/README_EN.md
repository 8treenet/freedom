# freedom
### base - base example

#### Directory Structure

- domain - Domain Model
    - aggregate
    - entity
    - dto - Data Transfer Object
    - po - Persistent Object
    - *.go - Domain Service

- adapter - Port Adapter
    - controller - In Adapter
    - repository - Out Adapter

- server - Server Program Entrance
    - conf - Config File
    - main.go - Main

- infra - Infrastructure
    - *go - Component   

---
#### Interface Introduction
```go
// main Application Install Interface 
type Application interface {
    // Install Gorm 
    InstallGorm(f func() (db *gorm.DB))
    // Install Redis
    InstallRedis(f func() (client redis.Cmdable))
    // Install Router Middleware
    InstallMiddleware(handler iris.Handler)
    //Install Bus Middleware，using by http2 example
    InstallBusMiddleware(handle ...BusHandler)
    //Install Global Party http://domian/relativePath/controllerParty
    InstallParty(relativePath string)
    // Create H2C Runner
    CreateH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
    // Create HTTP Runner 
    CreateRunner(addr string, configurators ...host.Configurator) iris.Runner
    // Return Iris Application
    Iris() *iris.Application
    // Logger
    Logger() *golog.Logger
    // Start Running
    Run(serve iris.Runner, c iris.Configuration)
    // Install Domain Events
    InstallDomainEventInfra(eventInfra DomainEventInfra)
    // Install Others, such as mongodb、es
    InstallOther(f func() interface{})
    // Start Callback Function: After Prepare And Before Run
    Start(f func(starter Starter))
    // Install Serializer: Used for serialization and deserialization, such as domain event objects or entity.
  	// Used official json by default
    InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
}

 /*
    Worker is a requesting runtime object, a request corresponds to an object. It can be injected into controller、service or repository.
 */
type Worker interface {
    // Get Iris Context
    IrisContext() freedom.Context
    // Logger with some context information
    Logger() Logger
    // Get the first level cache instance. When the request ends, the life cycle of the cache ends.
    Store() *memstore.Store
    // The bus can get the information transmitted from top to bottom
    Bus() *Bus
    // Common Context
    Context() stdContext.Context
    // With Common Context
    WithContext(stdContext.Context)
    // Start Time
    StartTime() time.Time
    // Delay Reclaiming
    DelayReclaiming()
}

// Initiator is the instance initialization interface, used in Prepare.
type Initiator interface {
    // Create Iris.Party. You can specify middleware.
    CreateParty(relativePath string, handlers ...context.Handler) iris.Party
    // Bind Controller By Iris.Party
    BindControllerByParty(party iris.Party, controller interface{})
    // Binding controller via relativePath. You can specify middleware.
    BindController(relativePath string, controller interface{}, handlers ...context.Handler)
    // Bind Service
    BindService(f interface{})
    // Get Service
    GetService(ctx iris.Context, service interface{})
    // Inject Somethings to Controller
    InjectController(f interface{})
    // Bind Factory
    BindFactory(f interface{})
    // Bind Repository
    BindRepository(f interface{})
    // Bind Infrastructure. If is a singleton, com is an object. if is multiton, com is a function
    BindInfra(single bool, com interface{})
    // Get Infrastructure. 
  	// Only the controller obtains the component that needs to be called in Prepare, 
  	// and service and repository can be directly dependency injected
    GetInfra(ctx iris.Context, com interface{})
    // Start Callback Function: After Prepare And Before Run
    Start(f func(starter Starter))
    Iris() *iris.Application
}

// Starter .
type Starter interface {
    Iris() *iris.Application
    // Asynchronous Cache Preheat
    AsyncCachePreheat(f func(repo *Repository))
    // Synchronous Cache Preheat
    CachePreheat(f func(repo *Repository))
    // Get Single Infrastructure
    GetSingleInfra(com interface{})
}

```

---
#### Application Cycle
|External Use | Internal Call |
| ----- | :---: |
|Application.InstallMiddleware| Register Install's global middleware |
|Application.InstallGorm|Trigger callback|
|freedom.Prepare|Trigger callback|
|Initiator.Starter|Trigger callback|
|infra.Booting|触发组件方法|
|http.Run|开启监听服务|
|infra.Closeing|Trigger callback|
|Application.Close|程序关闭|

#### Request Cycle
###### Each request starts with the creation of a number of dependency objects, such as worker, controller, service, repository, infra  and so on. Each request to use these objects independently, not more than one request concurrent read and write shared objects. Of course, there is no need to worry about efficiency, the framework has made the pool. The end of the request will recycle these objects. 

###### If the process uses go func(){/access related objects}, please call Worker.DelayReclaiming() before.

---
#### main
```go
func main() {
    app := freedom.NewApplication() // Create Application
    installDatabase(app)
    installRedis(app)
    installMiddleware(app)

    // Create HTTP Listener
    addrRunner := app.CreateRunner(conf.Get().App.Other["listen_addr"].(string))
    // Create HTTP2.0 H2C Listener
    addrRunner = app.CreateH2CRunner(conf.Get().App.Other["listen_addr"].(string))
    app.Run(addrRunner, *conf.Get().App)
}

func installMiddleware(app freedom.Application) {
    //Install Global Middleware
    app.InstallMiddleware(middleware.NewRecover())
    app.InstallMiddleware(middleware.NewTrace("x-request-id"))
    app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", true))

    requests.InstallPrometheus(conf.Get().App.Other["service_name"].(string), freedom.Prometheus())
    app.InstallBusMiddleware(middleware.NewBusFilter())
}

func installDatabase(app freedom.Application) {
    app.InstallGorm(func() (db *gorm.DB) {
        //Install DB CallBack Function
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
##### [Iris Routing Document]](https://github.com/kataras/iris/wiki/MVC)
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
		   Common binding, default controller to path '/'.
		   initiator.BindController("/", &DefaultController{})
		*/

		// Middleware binding. Valid only for this controller.
		// If you need global middleware, please add in main.
		initiator.BindController("/", &Default{}, func(ctx freedom.Context) {
			worker := freedom.ToWorker(ctx)
			worker.Logger().Info("Hello middleware begin")
			ctx.Next()
			worker.Logger().Info("Hello middleware end")
		})
	})
}

type Default struct {
  Sev    *domain.Default // Injected domain services. (Default)
  Worker freedom.Worker // Injected requesting runtime object. (Worker)
}

// Get handles the GET: / route.
func (c *Default) Get() freedom.Result {
    c.Worker.Logger().Infof("I'm Controller")
  	// Call services function
    remote := c.Sev.RemoteInfo() 
    // Return Json object
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
	// b.Handle("GET", "/custom", "CustomHello")
	// b.Handle("PUT", "/custom", "CustomHello")
	// b.Handle("POST", "/custom", "CustomHello")
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

#### Domain Service (domain/default.go)
```go
package domain

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/adapter/repository"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
            // Bind Default Service
            initiator.BindService(func() *Default {
                return &Default{}
            })
            initiator.InjectController(func(ctx freedom.Context) (service *Default) {
                // Injected Default to Controller
                initiator.GetService(ctx, &service)
                return
            })
	})
}

// Default .
type Default struct {
	Worker    freedom.Worker    // Injected requesting runtime object. (Worker)
	DefRepo   *repository.Default   // Injected repository object (DI)
  DefRepoIF repository.DefaultRepoInterface  // Injected repository interface (DIP)
}

// RemoteInfo .
func (s *Default) RemoteInfo() (result struct {
	Ip string
	Ua string
}) {
        s.Worker.Logger().Infof("I'm Service")
        // Call repository function
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
        // Bind default repository
		initiator.BindRepository(func() *Default {
			return &Default{}
		})
	})
}

// Default .
type Default struct {
    // Implement freedom.Repository
	freedom.Repository
}

// GetIP .
func (repo *Default) GetIP() string {
    // It have access to DB, Redis, NewHttp and Other only when 
  	// it inherit the repository, and you can get the worker directly.
    repo.DB()
    repo.Redis()
    repo.Other()
    repo.NewHttpRequest()
    repo.Worker.Logger().Infof("I'm Repository GetIP")
    return repo.Worker.IrisContext().RemoteAddr()
}

```

#### Configuration

|File | Effect |
| ----- | :---: |
|server/conf/app.toml|Service Configuration|
|server/conf/db.toml|DB Configuration|
|server/conf/redis.toml|Cache Configuration|
|server/conf/infra/.toml|Component Related Configuration|

