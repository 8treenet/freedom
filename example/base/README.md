#freedom
###base - 基础示例

> 目录结构

- business - 业务逻辑
    - controllers - 控制器目录
    - service - 服务目录
    - repositorys - 数据仓库目录

- cmd - 启动入口
    - conf - toml配置文件
    - main.go - 主函数

- components - 组件
    - config - 配置组件目录 *go

- models - 模型

---
> main.go
```go
//应用启动 传入绑定地址和应用配置;
freedom.Run(addrRunner, config.Get().App)

//使用全局中间件;
freedom.UseMiddleware(middleware.NewTrace("TRACE-ID"))

//返回要使用的 gorm 和 gcache 句柄, 如不使用可返回nil;
freedom.InstallGorm(func() (db *gorm.DB, cache gcache.Plugin) {
    //gcache.Plugin 是 gorm的缓存中间件
})

//返回redis 连接句柄
freedom.InstallRedis(func() (client *redis.Client) {
})

//安装第三方日志 logrus
freedom.Logger().Install(logrus.StandardLogger())
```

> 接口介绍
```go
 /*
    每一个请求都会串行的创建一系列对象(controller,service,repository)，不用太担心内存问题，因为底层有对象池。
    Runtime 运行时对象导出接口，一个请求创建一个运行时对象，在不同层里使用也是同一实例。
 */
type Runtime interface {
    //获取iris的上下文
    Ctx() iris.Context
    //获取带上下文的日志实例，可以注入上下文的trace。
    Logger() Logger
    //获取一级缓存实例，请求结束，该缓存生命周期结束。
    Store() *memstore.Store
    //获取普罗米修斯插件
    Prometheus() *Prometheus
}

// Initiator 初始化接口，在Booting使用。
type Initiator interface {
    //创建 iris.Party 中间件
    CreateParty(relativePath string, handlers ...context.Handler) iris.Party
    //绑定控制器到 iris.Party
    BindControllerByParty(party iris.Party, controller interface{}, service ...interface{})
    //直接绑定控制器到路径
    BindController(relativePath string, controller interface{}, service ...interface{})
    //绑定服务
    BindService(f interface{})
    //绑定Repo
    BindRepository(f interface{})
    //获取服务，freedom内部有 服务的对象池可复用。
    GetService(ctx iris.Context, service interface{})
    //异步缓存预热
    AsyncCachePreheat(f func(repo *Repository))
    //同步缓存预热
    CachePreheat(f func(repo *Repository))
    //绑定组件 如果组件是单例 com是对象， 如果组件是多例com是创建方法。
    BindComponent(single bool, com interface{})
    //获取组件, 只有控制器获取组件需要在booting内调用， service和repository可直接依赖注入
    GetComponent(ctx iris.Context, com interface{})
}
```

> controllers/default.go
```go
func init() {
    freedom.Booting(func(initiator freedom.Initiator) {
        //iris 依赖注入 返回要注入的服务对象，service/default.go。
        serFunc := func(ctx iris.Context) (m *services.DefaultService) {
            //从服务池里取出服务对象
            initiator.GetService(ctx, &m)
            return
        }
        //绑定Default控制器到路径 /
        initiator.BindController("/", &DefaultController{}, serFunc)
    })
}

/*
    控制器 DefaultController
    Sev : 使用服务对象定义成员变量，与init里的依赖注入相关。可以声明接口成员变量接收。
*/
type DefaultController struct {
    Sev     *services.DefaultService
    Runtime freedom.Runtime
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

> service/default.go
```go
func init() {
    freedom.Booting(func(initiator freedom.Initiator) {
        //绑定服务
        initiator.BindService(func() *DefaultService {
            return &DefaultService{}
        })
    })
}

//  服务 DefaultService
type DefaultService struct {
    Runtime freedom.Runtime

    //以下是实体调用和接口调用的2种成员变量，推荐使用接口隔离
    DefRepo   *repositorys.DefaultRepository // repository/default.go
    DefRepoIF DefaultRepoInterface
}

// RemoteInfo .
func (s *DefaultService) RemoteInfo() (result struct {
	IP string
	UA string
}) {
    s.Runtime.Logger().Infof("我是service")

    //调用repo获取数据
    result.IP = s.DefRepo.GetIP()
    result.UA = s.DefRepoIF.GetUA()
    return
}

// 接口声明
type DefaultRepoInterface interface {
    GetUA() string
}
```

> repositorys/default.go
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
    数据仓库主要用于 db、缓存、http、消息队列...的数据组织和抽象，示例直接返回上下文里的数据。
*/
type DefaultRepository struct {
    freedom.Repository
}

// GetIP .
func (repo *DefaultRepository) GetIP() string {
    //repo.DB().Find()
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