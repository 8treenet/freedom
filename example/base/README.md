# Freedom Framework

Freedom 是一个基于 DDD (Domain-Driven Design) 设计理念的 Go 语言框架，提供了清晰的分层架构和强大的依赖注入能力。它集成了 Iris Web 框架，并提供了完整的微服务开发支持。

## 特性

- 🏗️ **DDD 架构**: 完整支持领域驱动设计，包括聚合、实体、值对象等概念
- 💉 **依赖注入**: 强大的依赖注入系统，支持构造器注入和属性注入
- 🔌 **适配器模式**: 采用六边形架构（端口和适配器），实现清晰的代码分层
- 🚀 **高性能**: 基于对象池的请求隔离，确保并发安全和性能
- 📊 **可观测性**: 集成 Prometheus 监控、分布式追踪和结构化日志
- 🔒 **安全性**: 内置安全中间件，支持 TLS/SSL
- 🎯 **易测试**: 依赖倒置原则使得单元测试和集成测试更容易

## 目录结构

```
.
├── domain          # 领域模型层
│   ├── aggregate   # 聚合 - 实体的组合，确保业务不变性
│   ├── entity      # 实体 - 具有唯一标识的领域对象
│   ├── event      # 领域事件 - 领域模型中的状态变化
│   ├── vo         # 值对象 - 描述事物特征的对象
│   ├── po         # 持久化对象 - 数据库映射对象
│   └── *.go       # 领域服务 - 无法归属于实体的领域逻辑
│
├── adapter         # 端口适配器层
│   ├── controller # 控制器 (输入适配器) - 处理外部请求
│   └── repository # 仓库 (输出适配器) - 持久化领域对象
│
├── config        # 配置文件 - 应用配置管理
│
└── infra         # 基础设施组件 - 技术支持层
```

## 核心接口

### Application 接口

主应用程序接口，负责框架的核心配置和启动：

```go
type Application interface {
    // 数据存储相关
    InstallDB(f func() interface{})                    // 安装数据库
    InstallRedis(f func() (client redis.Cmdable))      // 安装 Redis
    InstallCustom(f func() interface{})                // 安装其他存储 (如 MongoDB、ES 等)
    
    // HTTP 服务相关
    InstallMiddleware(handler iris.Handler)            // 安装路由中间件
    InstallBusMiddleware(handle ...BusHandler)         // 安装链路中间件
    InstallParty(relativePath string)                  // 安装全局路由组
    
    // 服务器配置
    NewRunner(addr string, configurators ...host.Configurator) iris.Runner           // HTTP 服务
    NewH2CRunner(addr string, configurators ...host.Configurator) iris.Runner        // HTTP/2 服务
    NewAutoTLSRunner(addr string, domain string, email string, configurators ...host.Configurator) iris.Runner  // 自动 HTTPS
    NewTLSRunner(addr string, certFile, keyFile string, configurators ...host.Configurator) iris.Runner        // 手动 HTTPS
    
    // 工具函数
    Iris() *iris.Application                          // 获取 Iris 实例
    Logger() *golog.Logger                           // 获取日志实例
    Run(serve iris.Runner, c iris.Configuration)      // 启动服务
    BindBooting(f func(bootManager freedom.BootManager))  // 启动前回调
    InstallSerializer(                                // 自定义序列化
        marshal func(v interface{}) ([]byte, error),
        unmarshal func(data []byte, v interface{}) error,
    )
}
```

### Worker 接口

请求运行时对象，每个请求创建一个实例，支持依赖注入：

```go
type Worker interface {
    IrisContext() freedom.Context           // 获取 Iris 上下文
    Logger() Logger                         // 获取请求日志实例
    SetLogger(Logger)                       // 设置请求日志实例
    Store() *memstore.Store                // 获取请求级缓存
    Bus() *Bus                             // 获取数据总线
    Context() stdContext.Context           // 获取标准上下文
    WithContext(stdContext.Context)        // 设置标准上下文
    StartTime() time.Time                  // 获取请求开始时间
    DelayReclaiming()                      // 延迟对象回收
}
```

### Initiator 接口

实例初始化接口，用于依赖注入和控制器绑定：

```go
type Initiator interface {
    // 控制器相关
    CreateParty(relativePath string, handlers ...context.Handler) iris.Party
    BindControllerWithParty(party iris.Party, controller interface{})
    BindController(relativePath string, controller interface{}, handlers ...context.Handler)
    
    // 依赖注入
    BindService(f interface{})              // 注入服务
    BindFactory(f interface{})              // 注入工厂
    BindRepository(f interface{})           // 注入仓库
    BindInfra(single bool, com interface{}) // 注入基础组件
    
    // 控制器注入
    InjectController(f interface{})
    FetchInfra(ctx iris.Context, com interface{})
    FetchService(ctx iris.Context, service interface{})
    
    // 事件与启动
    BindBooting(f func(bootManager BootManager))
    ListenEvent(topic string, controller string)
    Iris() *iris.Application
}
```

## 生命周期

### 应用生命周期

| 阶段 | API | 说明 |
|------|-----|------|
| 全局中间件注册 | `Application.InstallMiddleware` | 注册全局中间件 |
| 数据库安装 | `Application.InstallDB` | 配置数据库连接 |
| 组件初始化 | `infra.Booting` | 初始化单例组件 |
| 启动前回调 | `Initiator.BindBooting` | 执行注册的启动回调 |
| 局部初始化 | `freedom.Prepare` | 初始化局部组件 |
| 服务启动 | `http.Run` | 启动 HTTP 服务 |
| 关闭回调 | `infra.RegisterShutdown` | 执行注册的关闭回调 |
| 应用关闭 | `Application.Close` | 关闭应用程序 |

### 请求生命周期

Freedom 框架为每个请求创建独立的运行时对象集合，包括：
- Worker: 请求上下文管理
- Controller: 请求处理控制器
- Service: 业务逻辑服务
- Factory: 对象工厂
- Repository: 数据访问层
- Infra 组件: 基础设施支持

这些对象都是请求隔离的，不会发生并发读写。框架使用对象池来确保性能。

> **注意**: 如果在请求处理过程中使用 goroutine 访问这些对象，请在启动 goroutine 前调用 `Worker.DelayReclaiming()` 以延迟对象回收。

## 最佳实践

### 依赖注入

推荐使用构造器注入方式：

```go
// 服务注册
freedom.Prepare(func(initiator freedom.Initiator) {
    initiator.BindService(func() *UserService {
        return &UserService{}
    })
})

// 控制器中使用
type UserController struct {
    UserSrv *UserService  // 自动注入
    Worker  freedom.Worker
}
```

### 事务管理

使用 Repository 模式管理事务：

```go
type UserRepository struct {
    freedom.Repository
}

func (r *UserRepository) CreateUser(user *entity.User) error {
    tx := r.DB().Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    if err := tx.Create(user).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```

### 错误处理

统一错误处理方式：

```go
// 定义错误码
const (
    ErrCodeNotFound = 404
    ErrCodeInternal = 500
)

// 错误响应结构
type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// 错误处理中间件
func ErrorHandler(ctx freedom.Context) {
    defer func() {
        if err := recover(); err != nil {
            ctx.JSON(500, ErrorResponse{
                Code:    ErrCodeInternal,
                Message: "Internal Server Error",
            })
        }
    }()
    ctx.Next()
}
```

## 性能优化

### 对象池

Freedom 框架使用对象池来提高性能：

```go
// 自定义对象池
type MyObject struct {
    // fields
}

var myObjectPool = sync.Pool{
    New: func() interface{} {
        return &MyObject{}
    },
}

// 获取对象
obj := myObjectPool.Get().(*MyObject)
defer myObjectPool.Put(obj)
```

### 缓存策略

使用多级缓存提高性能：

```go
func (s *UserService) GetUser(id string) (*entity.User, error) {
    // 1. 检查请求级缓存
    if user := s.Worker.Store().Get(id); user != nil {
        return user.(*entity.User), nil
    }

    // 2. 检查 Redis 缓存
    if user, err := s.Redis().Get(id).Result(); err == nil {
        return unmarshalUser(user), nil
    }

    // 3. 从数据库获取
    user, err := s.UserRepo.Find(id)
    if err != nil {
        return nil, err
    }

    // 4. 更新缓存
    s.Redis().Set(id, marshalUser(user), time.Hour)
    return user, nil
}
```

## 监控和日志

### Prometheus 监控

```go
// 注册 Prometheus 中间件
promMiddleware := middleware.NewClientPrometheus(
    "service_name",
    freedom.Prometheus(),
)
requests.InstallMiddleware(promMiddleware)
```

### 结构化日志

```go
// 请求日志中间件
app.InstallMiddleware(func(ctx freedom.Context) {
    worker := freedom.ToWorker(ctx)
    logger := worker.Logger().With(
        "request_id", ctx.GetHeader("X-Request-ID"),
        "method", ctx.Method(),
        "path", ctx.Path(),
    )
    worker.SetLogger(logger)
    
    start := time.Now()
    ctx.Next()
    
    logger.Info("request completed",
        "status", ctx.GetStatusCode(),
        "duration", time.Since(start),
    )
})
```

## 测试

### 单元测试

```go
func TestUserService_CreateUser(t *testing.T) {
    // 准备测试环境
    app := freedom.NewTestApplication()
    
    // 注入 mock 依赖
    app.InstallDB(func() interface{} {
        return &mockDB{}
    })
    
    // 创建测试实例
    srv := &UserService{}
    app.Inject(srv)
    
    // 执行测试
    user := &entity.User{Name: "test"}
    err := srv.CreateUser(user)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### 集成测试

```go
func TestUserAPI(t *testing.T) {
    app := freedom.NewTestApplication()
    e := httptest.New(t, app.Iris())
    
    // 测试创建用户
    e.POST("/users").
        WithJSON(map[string]string{"name": "test"}).
        Expect().
        Status(http.StatusOK).
        JSON().Object().
        HasValue("id", gomega.Not(gomega.BeEmpty()))
}
```

## 部署

### Docker 支持

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o main ./main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 8080
CMD ["./main"]
```

### Kubernetes 配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: freedom-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: freedom-app
  template:
    metadata:
      labels:
        app: freedom-app
    spec:
      containers:
      - name: freedom-app
        image: freedom-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: db_host
```

## 更多资源

- [Iris 路由文档](https://github.com/kataras/iris/wiki/MVC)
- [完整示例代码](https://github.com/8treenet/freedom/tree/master/example)
- [API 文档](https://pkg.go.dev/github.com/8treenet/freedom)
- [性能基准测试](https://github.com/8treenet/freedom/tree/master/benchmark)
- [常见问题解答](https://github.com/8treenet/freedom/wiki/FAQ)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 提交问题和建议
- 改进文档
- 提交代码修复
- 添加新功能

请参阅我们的[贡献指南](CONTRIBUTING.md)了解更多信息。

## License

MIT License

Copyright (c) 2023 Freedom Framework

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
