# freedom
### base - 基础示例

#### 目录结构

- domain - 领域
    - service - 服务目录
    - repositorys - 数据仓库目录
    - entity - 实体.

- cmd - 启动入口
    - conf - toml配置文件
    - main.go - 主函数
- controllers - 控制器目录
- infra - 基础设施
    - config - 配置组件

- objects - 值对象

---
#### 入口main.go
```go

//创建应用
app := freedom.NewApplication()
//安装中间件
installMiddleware(app)

//应用启动 传入绑定地址和应用配置;
app.Run(addrRunner, *config.Get().App)

func installMiddleware(app freedom.Application) {
    //安装TRACE中间件，框架已默认安装普罗米修斯
    app.InstallMiddleware(middleware.NewTrace("TRACE-ID"))
    app.InstallMiddleware(middleware.NewLogger("TRACE-ID", true))
    app.InstallMiddleware(middleware.NewRuntimeLogger("TRACE-ID"))
}
```

#### 接口介绍
```go
 /*
    每一个请求都会串行的创建一系列对象(controller,service,repository)，不用太担心内存问题，因为底层有对象池。
    Runtime 运行时对象导出接口，一个请求创建一个运行时对象，在不同层里使用也是同一实例。
 */
type Runtime interface {
    //获取iris的上下文
    Ctx() iris.Context
    //获取带上下文的日志实例。
    Logger() Logger
    //获取一级缓存实例，请求结束，该缓存生命周期结束。
    Store() *memstore.Store
    //获取普罗米修斯插件
    Prometheus() *Prometheus
}

// Initiator 实例初始化接口，在Booting使用。
type Initiator interface {
    //创建 iris.Party 中间件
    CreateParty(relativePath string, handlers ...context.Handler) iris.Party
    //绑定控制器到 iris.Party
    BindControllerByParty(party iris.Party, controller interface{}, service ...interface{})
    //直接绑定控制器到路径
    BindController(relativePath string, controller interface{}, service ...interface{})
    //绑定服务
    BindService(f interface{})
    //注入到控制器
    InjectController(f interface{})
    //绑定Repo
    BindRepository(f interface{})
    //获取服务
    GetService(ctx iris.Context, service interface{})
    //异步缓存预热
    AsyncCachePreheat(f func(repo *Repository))
    //同步缓存预热
    CachePreheat(f func(repo *Repository))
    //绑定基础设施组件 如果组件是单例 com是对象， 如果组件是多例com是创建函数。
    BindInfra(single bool, com interface{})
    //获取基础设施组件, 只有控制器获取组件需要在booting内调用， service和repository可直接依赖注入
    GetInfra(ctx iris.Context, com interface{})
}
```

#### controllers/default.go
```go
func init() {
    freedom.Booting(func(initiator freedom.Initiator) {
        //绑定Default控制器到路径 /
        initiator.BindController("/", &DefaultController{})
    })
}

/*
    控制器 DefaultController
    Sev : 使用服务对象定义成员变量，与init里的依赖注入相关。可以声明接口成员变量接收。
*/
type DefaultController struct {
    Sev     *services.DefaultService //被注入的变量名，必须是大写开头。go规范小写变量名，反射是无法注入的。
    Runtime freedom.Runtime //注入请求运行时
}

/* 
    Get handles the GET: / route.
    返回json数据 result到客户端
*/
func (c *DefaultController) Get() (result struct {
    IP string
    UA string
}, e error) {
    
    //打印日志，并且调用服务的 RemoteInfo 方法。
    c.Runtime.Logger().Infof("我是控制器")
    remote := c.Sev.RemoteInfo()

    result.IP = remote.IP
    result.UA = remote.UA
    return
}
```

#### service/default.go
```go
func init() {
    freedom.Booting(func(initiator freedom.Initiator) {
        //绑定服务
        initiator.BindService(func() *DefaultService {
            return &DefaultService{}
        })

        //控制器中使用依赖注入, 需要注入到控制器.
        initiator.InjectController(func(ctx freedom.Context) (ds *DefaultService) {
            //从对象池中获取service
            initiator.GetService(ctx, &ds)
            return
        })
    })
}

//  服务 DefaultService
type DefaultService struct {
    Runtime freedom.Runtime
    DefRepo   *repositorys.DefaultRepository
}

// RemoteInfo .
func (s *DefaultService) RemoteInfo() (result struct {
	IP string
	UA string
}) {
    s.Runtime.Logger().Infof("我是service")

    //调用repo获取数据
    result.IP = s.DefRepo.GetIP()
    result.UA = s.DefRepo.GetUA()
    return
}
```

#### repositorys/default.go
```go
func init() {
    freedom.Booting(func(initiator freedom.Initiator) {
        //绑定 Repository
        initiator.BindRepository(func() *DefaultRepository {
            return &DefaultRepository{}
        })
    })
}

/* 
    数据仓库 DefaultRepository 
    数据仓库主要用于 db、缓存、http...的数据组织和抽象，示例直接返回上下文里的ip数据。
*/
type DefaultRepository struct {
    freedom.Repository
}

// GetIP .
func (repo *DefaultRepository) GetIP() string {
    repo.Runtime.Logger().Infof("我是Repository GetIP")
    return repo.Runtime.Ctx().RemoteAddr()
}

// GetUA - implment DefaultRepoInterface interface
func (repo *DefaultRepository) GetUA() string {
    //该方法的接口声明在service/default.go, 推荐创建接口目录来声明一揽子接口。
    repo.Runtime.Logger().Infof("我是Repository GetUA")
    return repo.Runtime.Ctx().Request().UserAgent()
}  
```

#### 配置文件

|文件 | 作用 |
| ----- | :---: |
|cmd/conf/app.toml|服务配置|
|cmd/conf/db.toml|db配置|
|cmd/conf/redis.toml|缓存配置|

#### 生命周期
###### 每一个请求接入都会创建若干依赖对象，从controller、service、repository、infra。简单的讲每一个请求都是独立的创建和使用这一系列对象，不会导致并发的问题。当然也无需担心效率问题，框架已经做了池。可以参见 initiator 里的各种Bind方法。