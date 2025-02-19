# Freedom HTTP/2 示例项目

## 目录
- [1. 项目简介](#1-项目简介)
  - [1.1 HTTP/2 与 H2C 说明](#11-http2-与-h2c-说明)
- [2. 快速开始](#2-快速开始)
- [3. HTTP/H2C 服务器](#3-httph2c-服务器)
- [4. HTTP/H2C 客户端](#4-httph2c-客户端)
- [5. 中间件系统](#5-中间件系统)
- [6. 依赖注入](#6-依赖注入)
- [7. 最佳实践](#7-最佳实践)
- [8. 相关文档](#8-相关文档)

## 1. 项目简介

本项目展示了 Freedom 框架对 HTTP/2 和 H2C (HTTP/2 Cleartext) 的完整支持。项目实现了一个简单的购物系统示例，演示了：

- HTTP/2 和 H2C 服务器的配置和使用
- HTTP/2 和 H2C 客户端的实现
- 中间件系统的应用
- 依赖注入和接口设计
- 分布式追踪和日志记录

> 💡 完整的 HTTP 客户端使用指南请参考：[HTTP 客户端完整指南](../doc/http-client-guide.md)

### 1.1 HTTP/2 与 H2C 说明

HTTP/2 有两种工作模式：

1. **HTTP/2 (基于 TLS)**
   - 需要 TLS 加密
   - 默认用于外网通信
   - 提供更好的安全性
   - 需要证书管理
   - 适用场景：面向公网的服务

2. **HTTP/2 Cleartext (H2C)**
   - 无需 TLS 加密
   - 更适合内网服务间通信
   - 性能开销更小
   - 无需证书管理
   - 适用场景：内网微服务通信

**为什么内网服务通信推荐使用 H2C？**

1. **性能优势**
   - 省去 TLS 握手开销
   - 减少加解密运算
   - 降低延迟
   - 更少的 CPU 消耗

2. **运维便利**
   - 无需管理和更新证书
   - 配置更简单
   - 部署更方便
   - 降低运维成本

3. **HTTP/2 特性**
   - 多路复用
   - 头部压缩
   - 服务器推送
   - 二进制分帧

本示例项目默认使用 H2C 模式，特别适合于微服务架构中的内部服务间通信。

## 2. 快速开始

### 2.1 运行服务

```bash
$ go run server/main.go
```

### 2.2 访问服务

浏览器访问：`http://127.0.0.1:8000/shop/{id}`  
其中 `{id}` 为商品ID，可用范围：1-4

## 3. HTTP/H2C 服务器

Freedom 框架支持多种 HTTP 服务器模式：

```go
// 创建应用实例
app := freedom.NewApplication()

// 1. H2C 服务器 (HTTP/2 Cleartext)
h2cRunner := app.NewH2CRunner(":8000")

// 2. HTTP/2 TLS 服务器
tlsRunner := app.NewTLSRunner(":443", "certFile", "keyFile")

// 3. HTTP/2 AutoTLS 服务器 (自动获取 Let's Encrypt 证书)
autoTLSRunner := app.NewAutoTLSRunner(":443", "domain.com", "email@example.com")

// 运行服务器
app.Run(h2cRunner, conf.Get().App)
```

## 4. HTTP/H2C 客户端

框架提供了两种 HTTP 客户端：

```go
// 1. H2C 客户端请求
repo.NewH2CRequest(addr).Get().ToJSON(&result)

// 2. 普通 HTTP 客户端请求
repo.NewHTTPRequest(addr, false).Get().ToJSON(&result)
```

客户端特性：
- 支持请求追踪
- 自动连接池管理
- 超时控制
- 并发安全

## 5. 中间件系统

示例项目包含了以下中间件：

```go
// 安装中间件
func installMiddleware(app freedom.Application) {
    // Recover 中间件：处理 panic
    app.InstallMiddleware(middleware.NewRecover())
    
    // Trace 中间件：分布式追踪
    app.InstallMiddleware(middleware.NewTrace("x-request-id"))
    
    // 请求日志中间件
    app.InstallMiddleware(middleware.NewRequestLogger("x-request-id"))
    
    // Prometheus 监控中间件
    middle := middleware.NewClientPrometheus(serviceName, freedom.Prometheus())
    requests.InstallMiddleware(middle)
    
    // Bus 中间件：处理服务间通信
    app.InstallBusMiddleware(middleware.NewBusFilter())
}
```

## 6. 依赖注入

项目展示了基于接口的依赖注入模式：

```go
// 1. 定义接口
type GoodsInterface interface {
    GetGoods(goodsID int) vo.GoodsModel
}

// 2. 注入服务
type ShopService struct {
    Worker freedom.Worker
    Goods  repository.GoodsInterface  // 注入接口
}

// 3. 实现接口
type GoodsRepository struct {
    freedom.Repository
}

func (repo *GoodsRepository) GetGoods(goodsID int) vo.GoodsModel {
    // 实现细节
}
```

## 7. 最佳实践

### 7.1 请求追踪

```go
// 启用请求追踪
req := repo.NewH2CRequest(addr)
response := req.Get().ToJSON(&result)

// 获取追踪信息
traceInfo := response.TraceInfo()
```

### 7.2 并发处理

```go
// 使用 DeferRecycle 处理并发请求
repo.Worker().DeferRecycle()
go func() {
    var model vo.GoodsModel
    repo.NewH2CRequest(addr).Get().ToJSON(&model)
}()
```

### 7.3 错误处理

```go
// 使用中间件处理错误
app.InstallMiddleware(middleware.NewRecover())

// 请求错误处理
response := req.Get().ToJSON(&result)
if response.Error != nil {
    // 错误处理逻辑
}
```
