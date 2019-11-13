#freedom
###base - 基础示例

> 目录结构

- business - 业务逻辑
    - controllers - 控制器目录
    - service - 服务目录
    - repositorys - 数据源目录

- cmd - 启动入口
    - conf - toml配置文件
    - main.go - 启动

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
> freedom.Booting
```go
type Initiator interface {
    freedom.Booting(func(initiator freedom.Initiator) {
    })
    
    CreateParty(relativePath string, handlers ...context.Handler) iris.Party
    BindController(relativePath string, controller interface{}, service ...interface{})
    BindControllerByParty(party iris.Party, controller interface{}, service ...interface{})
    BindService(f interface{})
    BindRepository(f interface{})
    GetService(ctx iris.Context, service interface{})
    AsyncCachePreheat(f func(repo *Repository))
    CachePreheat(f func(repo *Repository))
    //BindComponent 如果是单例 com是对象， 如果是多例，com是函数
    BindComponent(single bool, com interface{})
    GetComponent(ctx iris.Context, com interface{})
}
```
> controllers/default.go
```go
func init() {
    freedom.Booting(func(initiator freedom.Initiator) {
        serFunc := func(ctx iris.Context) (m *services.DefaultService) {
            initiator.GetService(ctx, &m)
            return
        }
        initiator.BindController("/", &DefaultController{}, serFunc)
    })
}
```