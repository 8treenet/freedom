# Worker 运行时指南

## 1. 运行时概述

Freedom 框架的 Worker 运行时是一个核心机制，它为每个请求提供了一个独立的运行时环境。通过 Worker，框架实现了请求级别的依赖注入和生命周期管理，使得开发者无需手动传递 context，就能在整个请求链路中共享请求上下文。

### 1.1 运行时特点

1. 请求隔离：每个请求都有独立的 Worker 实例
2. 自动注入：框架自动将 Worker 注入到各个组件中
3. 生命周期管理：Worker 生命周期与请求一致
4. 无侵入传递：避免了手动传递 context
5. 请求级别存储：提供了请求级别的存储空间

### 1.2 Worker 接口定义

```go
type Worker interface {
    // 获取请求上下文
    IrisContext() Context
    
    // 获取日志组件
    Logger() *golog.Logger
    
    // 获取请求存储空间
    Store() Store
    
    // 获取请求总线
    Bus() Bus
    
    // 开启延迟回收
    DeferRecycle()
    
    // 获取请求上下文
    Context() context.Context
}
```

## 2. 核心功能

### 2.1 请求上下文

Worker 提供了对请求上下文的访问，可以方便地获取请求参数和处理请求：

```go
func (s *UserService) HandleRequest() {
    // 获取原始请求上下文
    ctx := s.Worker.IrisContext()
    
    // 获取请求参数
    userID := ctx.URLParam("user_id")
}
```

### 2.2 日志记录

#### 2.2.1 基本使用

Worker 内置了结构化日志支持，并且会自动关联请求追踪ID：

```go
func (s *UserService) LogExample() {
    // 普通日志 - 会自动带上 X-Request-ID
    // 输出示例: [2024-01-01 10:00:00] [INFO] [X-Request-ID=trace-123] 处理请求
    s.Worker.Logger().Info("处理请求")
    
    // 带字段的结构化日志
    s.Worker.Logger().Infof("用户操作", freedom.LogFields{
        "user_id": 123,
        "action": "login",
        "ip": s.Worker.IrisContext().RemoteAddr(),
    })
    
    // 错误日志
    s.Worker.Logger().Error("发生错误", freedom.LogFields{
        "error": err.Error(),
        "stack": debug.Stack(),
    })
}
```

#### 2.2.2 日志级别使用

```go
// 调试信息
s.Worker.Logger().Debug("详细处理过程", freedom.LogFields{
    "step": "validate",
    "data": data,
})

// 普通信息
s.Worker.Logger().Info("处理成功")

// 警告信息
s.Worker.Logger().Warn("重试操作", freedom.LogFields{
    "retry_count": count,
})

// 错误信息
s.Worker.Logger().Error("处理失败", freedom.LogFields{
    "error": err.Error(),
    "stack": debug.Stack(),
})
```

#### 2.2.3 日志最佳实践

1. 使用结构化字段记录关键信息
2. 合理使用日志级别
3. 在分布式系统中追踪请求
4. 记录必要的上下文信息
5. 避免记录敏感信息

### 2.3 请求存储空间

Worker 提供了请求级别的键值存储，可用于在请求生命周期内共享数据：

```go
func (s *UserService) StoreExample() {
    store := s.Worker.Store()
    
    // 存储数据
    store.Set("user_id", 123)
    store.Set("request_time", time.Now())
    
    // 获取数据
    userID := store.Get("user_id")
    
    // 删除数据
    store.Remove("user_id")
    
    // 检查键是否存在
    if store.Contains("user_id") {
        // 处理逻辑
    }
}
```

### 2.4 请求总线

Worker 的 Bus 用于服务间通信和请求追踪：

```go
func (s *UserService) BusExample() {
    bus := s.Worker.Bus()
    
    // 设置业务数据
    bus.Set("userId", "123")  // 服务间通信传递给上游的服务
}
```

## 3. 链路追踪

### 3.1 配置追踪中间件

```go
func main() {
    app := freedom.NewApplication()
    
    // 注册 Trace 中间件，指定使用 x-request-id 作为追踪ID
    app.InstallMiddleware(middleware.NewTrace("x-request-id"))
    
    // 注册请求日志中间件，记录追踪ID
    loggerConfig := middleware.DefaultLoggerConfig()
    app.InstallMiddleware(middleware.NewRequestLogger("x-request-id", loggerConfig))
}
```

### 3.2 服务间调用

```go
func (repo *UserRepository) CallUserService() {
    // 创建 HTTP 请求，第二个参数为 true 表示启用总线传递
    // 这将自动传递 X-Request-ID 和其他 Bus 中的数据
    request := repo.NewHTTPRequest("http://other-service/api", true)
    
    // 发起请求
    var response struct {
        Code int    `json:"code"`
        Data string `json:"data"`
    }
    err := request.Get().ToJSON(&response).Error
    
    // 日志中会自动包含同一个 X-Request-ID
    repo.Worker.Logger().Info("调用用户服务", freedom.LogFields{
        "url": "http://user-service/api",
        "response": response,
        "error": err,
    })
}
```

### 3.3 追踪ID传递机制

1. 自动获取：框架自动从请求头获取 X-Request-ID
2. 自动生成：请求头无 X-Request-ID 时自动生成
3. 自动注入：日志自动包含追踪ID
4. 链路传递：服务间调用自动传递

### 3.4 链路追踪特点

1. 自动传递：
   - 服务间调用自动传递 X-Request-ID
   - Bus 中的其他数据也会自动传递
   - 异步处理中保持同一个追踪ID

2. 日志关联：
   - 所有服务的日志都带有相同的追踪ID
   - 可以通过追踪ID查看完整调用链
   - 支持分布式系统的问题排查

3. 使用建议：
   - 总是使用 `NewHTTPRequest(url, true)` 启用总线传递
   - 在日志中记录关键业务节点
   - 合理设置要传递的 Bus 数据

## 4. 延迟回收机制

### 4.1 基本概念

Worker 默认在请求结束时立即回收，但在异步场景下，我们需要延长 Worker 的生命周期。使用 DeferRecycle 可以实现这一点。

### 4.2 使用方法

```go
func (s *UserService) AsyncProcess() {
    // 1. 开启延迟回收
    s.Worker.DeferRecycle()
    
    // 2. 启动异步处理
    go func() {
        defer func() {
            if err := recover(); err != nil {
                s.Worker.Logger().Error("异步处理异常:", err)
            }
        }()
        
        // 3. 异步处理逻辑
        s.Worker.Logger().Info("开始异步处理")
        s.processAsync()
        s.Worker.Logger().Info("异步处理完成")
    }()
}
```

### 4.3 使用场景

#### 4.3.1 异步日志记录

```go
func (s *LogService) AsyncLog() {
    s.Worker.DeferRecycle()
    go func() {
        s.Worker.Logger().Info("异步记录日志")
        s.logRepo.Save(logData)
    }()
}
```

#### 4.3.2 消息队列处理

```go
func (repo *MQRepository) PublishMessage() {
    repo.Worker.DeferRecycle()
    go func() {
        msg := createMessage()
        repo.Worker.Logger().Info("发送消息到MQ")
        repo.mqClient.Publish(msg)
    }()
}
```

#### 4.3.3 异步通知

```go
func (repo *NotifyRepository) SendNotification() {
    repo.Worker.DeferRecycle()
    go func() {
        repo.Worker.Logger().Info("发送通知")
        repo.notifyRepo.SendEmail()
    }()
}
```

### 4.4 注意事项

1. 必须在启动 goroutine 前调用 DeferRecycle
2. 做好错误处理和恢复机制
3. 避免过长时间的异步处理
4. 及时清理资源
5. 合理使用日志记录异步处理状态

## 5. 最佳实践

### 5.1 日志使用

1. 使用结构化日志记录业务信息
2. 记录关键业务节点和状态变更
3. 合理使用不同日志级别
4. 包含必要的上下文信息
5. 避免记录敏感数据

### 5.2 链路追踪

1. 启用总线传递功能
2. 保持追踪ID一致性
3. 记录完整调用链
4. 合理使用 Bus 传递业务数据
5. 监控服务调用链路

### 5.3 异步处理

1. 正确使用 DeferRecycle
2. 实现完善的错误处理
3. 控制异步执行时间
4. 避免资源泄露
5. 保持日志的连续性

### 5.4 调试技巧

1. 追踪 Worker 生命周期
2. 监控异步处理状态
3. 使用追踪ID排查问题
4. 检查日志完整性
5. 验证数据传递正确性

### 5.5 性能优化

1. 控制异步任务数量
2. 及时释放资源
3. 避免过度日志记录
4. 优化服务间调用
5. 合理使用存储空间
