# Freedom HTTP 客户端完整指南

## 目录
- [1. 基础用法](#1-基础用法)
  - [1.1 创建请求](#11-创建请求)
  - [1.2 HTTP 方法](#12-http-方法)
  - [1.3 请求参数设置](#13-请求参数设置)
  - [1.4 响应处理](#14-响应处理)
- [2. 中间件](#2-中间件)
  - [2.1 中间件定义](#21-中间件定义)
  - [2.2 中间件安装](#22-中间件安装)
  - [2.3 常用中间件示例](#23-常用中间件示例)
- [3. 自定义客户端](#3-自定义客户端)
  - [3.1 HTTP 客户端](#31-http-客户端)
  - [3.2 H2C 客户端](#32-h2c-客户端)
  - [3.3 超时设置](#33-超时设置)
- [4. 高级特性](#4-高级特性)
  - [4.1 请求追踪](#41-请求追踪)
  - [4.2 并发控制](#42-并发控制)
  - [4.3 上下文控制](#43-上下文控制)
- [5. 最佳实践](#5-最佳实践)
  - [5.1 Repository 模式](#51-repository-模式)
  - [5.2 错误处理](#52-错误处理)
  - [5.3 性能优化](#53-性能优化)

## 1. 基础用法

### 1.1 创建请求

Freedom 框架提供了两种创建 HTTP 客户端请求的方式：

1. 直接使用 requests 包：

```go
import "github.com/8treenet/freedom/infra/requests"

// 创建普通 HTTP 请求
req := requests.NewHTTPRequest("http://example.com")

// 创建 HTTP/2 Cleartext (H2C) 请求
req := requests.NewH2CRequest("http://example.com") 
```

2. 通过 Repository 创建：

```go
// 在 Repository 中创建请求
type MyRepository struct {
    freedom.Repository
}

func (repo *MyRepository) DoSomething() {
    // 创建普通 HTTP 请求，第二个参数控制是否传递上下文信息，默认为 true
    req := repo.NewHTTPRequest("http://example.com", true)
    
    // 创建 H2C 请求
    req := repo.NewH2CRequest("http://example.com", true)
}
```

### 1.2 HTTP 方法

支持所有标准的 HTTP 方法：

```go
req.Get()      // GET 请求
req.Post()     // POST 请求
req.Put()      // PUT 请求
req.Delete()   // DELETE 请求
req.Head()     // HEAD 请求
req.Options()  // OPTIONS 请求
```

### 1.3 请求参数设置

1. 查询参数：

```go
// 设置单个查询参数
req.SetQueryParam("key", "value")

// 设置多个查询参数
params := map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
    "ids": []int{1, 2, 3}, // 支持切片类型
}
req.SetQueryParams(params)
```

2. 请求体：

```go
// JSON 请求体
data := struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}{
    Name: "freedom",
    Age:  1,
}
req.SetJSONBody(data)

// Form 表单
formData := url.Values{}
formData.Set("username", "freedom")
formData.Set("password", "123456")
req.SetFormBody(formData)

// 文件上传
req.SetFile("file", "/path/to/file.jpg")

// 原始数据
req.SetBody([]byte("raw data"))
```

3. 请求头：

```go
// 设置单个 header
req.AddHeader("Content-Type", "application/json")

// 设置多个 headers
headers := http.Header{}
headers.Set("Authorization", "Bearer token")
headers.Set("X-Custom-Header", "value")
req.SetHeader(headers)
```

### 1.4 响应处理

1. JSON 响应：

```go
var result struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data interface{} `json:"data"`
}
response := req.Get().ToJSON(&result)
if response.Error != nil {
    // 处理错误
}
```

2. 其他响应格式：

```go
// 字符串响应
str, response := req.Get().ToString()

// 字节数组响应
bytes, response := req.Get().ToBytes()

// XML 响应
var xmlData interface{}
response := req.Get().ToXML(&xmlData)
```

## 2. 中间件

### 2.1 中间件定义

中间件是一个实现了 `requests.Handler` 的函数：

```go
type Handler func(Middleware)

type Middleware interface {
    Next()
    Stop(...error)
    GetRequest() *http.Request
    GetRespone() *Response
    GetResponeBody() []byte
    IsStopped() bool
    IsH2C() bool
    Context() context.Context
    WithContextFromMiddleware(context.Context)
    EnableTraceFromMiddleware()
    SetClientFromMiddleware(Client)
}
```

### 2.2 中间件安装

```go
// 安装全局中间件
requests.InstallMiddleware(middleware1, middleware2)
```

### 2.3 常用中间件示例

1. 日志中间件：

```go
func NewLogMiddleware() requests.Handler {
    return func(middle requests.Middleware) {
        // 请求前
        req := middle.GetRequest()
        startTime := time.Now()
        fmt.Printf("[HTTP] >>> %s %s\n", req.Method, req.URL)
        
        middle.Next()
        
        // 请求后
        resp := middle.GetRespone()
        duration := time.Since(startTime)
        fmt.Printf("[HTTP] <<< %s %s %d %v\n", 
            req.Method, req.URL, resp.StatusCode, duration)
    }
}
```

2. 追踪中间件：

```go
func NewTraceMiddleware() requests.Handler {
    return func(middle requests.Middleware) {
        middle.EnableTraceFromMiddleware()
        
        // 添加追踪 ID
        traceID := uuid.New().String()
        middle.GetRequest().Header.Set("X-Trace-ID", traceID)
        
        middle.Next()
        
        // 记录追踪信息
        traceInfo := middle.GetRespone().TraceInfo()
        fmt.Printf("Trace[%s] Total=%v DNS=%v Connect=%v\n",
            traceID,
            traceInfo.TotalTime,
            traceInfo.DNSLookup,
            traceInfo.ConnTime)
    }
}
```

3. 重试中间件：

```go
func NewRetryMiddleware(maxRetries int) requests.Handler {
    return func(middle requests.Middleware) {
        var lastErr error
        for i := 0; i < maxRetries; i++ {
            middle.Next()
            
            resp := middle.GetRespone()
            if resp.Error == nil && resp.StatusCode < 500 {
                return
            }
            
            lastErr = resp.Error
            if i < maxRetries-1 {
                time.Sleep(time.Second * time.Duration(i+1))
            }
        }
        
        middle.Stop(fmt.Errorf("重试 %d 次后失败: %v", maxRetries, lastErr))
    }
}
```

## 3. 自定义客户端

### 3.1 HTTP 客户端

```go
// 自定义客户端实现
type CustomClient struct {
    *requests.ClientImpl
}

func NewCustomHTTPClient() *CustomClient {
    transport := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,
            KeepAlive: 30 * time.Second,
            DualStack: true,
        }).DialContext,
        MaxIdleConns:          1000,
        MaxIdleConnsPerHost:   100,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   5 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   10 * time.Second,
    }

    return &CustomClient{
        ClientImpl: &requests.ClientImpl{Client: client},
    }
}

// 设置全局客户端
func init() {
    customClient := NewCustomHTTPClient()
    requests.SetHTTPClient(customClient)
}
```

### 3.2 H2C 客户端

```go
func NewCustomH2CClient() *CustomClient {
    transport := &http2.Transport{
        AllowHTTP: true,
        DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
            return net.DialTimeout(network, addr, 5*time.Second)
        },
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   10 * time.Second,
    }

    return &CustomClient{
        ClientImpl: &requests.ClientImpl{Client: client},
    }
}

// 设置全局 H2C 客户端
func init() {
    customH2CClient := NewCustomH2CClient()
    requests.SetH2CClient(customH2CClient)
}
```

### 3.3 超时设置

```go
// 初始化带有自定义超时的客户端
func InitCustomTimeoutClients() {
    // HTTP 客户端 (读写超时10秒，连接超时2秒)
    requests.InitHTTPClient(10*time.Second, 2*time.Second)
    
    // H2C 客户端 (读写超时5秒，连接超时1秒)
    requests.InitH2CClient(5*time.Second, 1*time.Second)
}
```

## 4. 高级特性

### 4.1 HTTP连接请求追踪

```go
// 启用请求追踪
req := requests.NewHTTPRequest("http://example.com")
req.EnableTrace() //开启HTTP连接请求追踪

response := req.Get().ToJSON(&result)

// 获取追踪信息
traceInfo := response.TraceInfo()
fmt.Printf("DNS查询: %v\n", traceInfo.DNSLookup)
fmt.Printf("连接时间: %v\n", traceInfo.ConnTime)
fmt.Printf("总耗时: %v\n", traceInfo.TotalTime)
fmt.Printf("是否复用连接: %v\n", traceInfo.IsConnReused)
```

### 4.2 并发控制

1. Singleflight（请求去重）：

```go
func ConcurrentRequests() {
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            var result interface{}
            // 相同的 key 在并发请求时只会发起一次请求
            response := requests.NewHTTPRequest("http://api.example.com/data")
                .Get()
                .Singleflight("cache_key")
                .ToJSON(&result)
                
            if response.Error != nil {
                log.Printf("Error: %v", response.Error)
            }
        }()
    }
    
    wg.Wait()
}
```

### 4.3 上下文控制

```go
// 超时控制
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req := requests.NewHTTPRequest("http://example.com")
req.WithContext(ctx)

// 带有自定义值的上下文
ctx = context.WithValue(context.Background(), "key", "value")
req.WithContext(ctx)
```

## 5. 最佳实践

### 5.1 Repository 模式

在 Freedom 框架中，推荐在 Repository 层中封装 HTTP 请求：

```go
type UserRepository struct {
    freedom.Repository
}

func (repo *UserRepository) GetUserInfo(userID string) (*UserInfo, error) {
    var result struct {
        Code int       `json:"code"`
        Data UserInfo  `json:"data"`
        Msg  string    `json:"msg"`
    }
    
    response := repo.NewHTTPRequest("http://api.example.com/users/"+userID)
        .Get()
        .EnableTrace()//开启HTTP连接请求追踪
        .ToJSON(&result)
        
    if response.Error != nil {
        return nil, response.Error
    }
    
    if result.Code != 0 {
        return nil, errors.New(result.Msg)
    }
    
    return &result.Data, nil
}
```

### 5.2 错误处理

1. 请求错误处理：

```go
response := req.Get().ToJSON(&result)
if response.Error != nil {
    // 网络错误、超时等
    return fmt.Errorf("请求失败: %w", response.Error)
}

// HTTP 状态码检查
if response.StatusCode != http.StatusOK {
    return fmt.Errorf("服务器错误: %d", response.StatusCode)
}

// 业务逻辑错误
if result.Code != 0 {
    return fmt.Errorf("业务错误: %s", result.Msg)
}
```

2. 中间件错误处理：

```go
func ErrorHandlingMiddleware() requests.Handler {
    return func(middle requests.Middleware) {
        middle.Next()
        
        resp := middle.GetRespone()
        if resp.Error != nil {
            // 记录错误日志
            log.Printf("请求错误: %v", resp.Error)
            return
        }
        
        if resp.StatusCode >= 500 {
            middle.Stop(fmt.Errorf("服务器错误: %d", resp.StatusCode))
            return
        }
    }
}
```

### 5.3 性能优化

1. 连接池配置：

```go
transport := &http.Transport{
    MaxIdleConns:        100,              // 最大空闲连接数
    MaxIdleConnsPerHost: 10,               // 每个主机的最大空闲连接数
    IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
}
```

2. 使用 H2C：
- 对于内部服务间通信，推荐使用 H2C
- H2C 支持多路复用，可以减少连接建立的开销

3. 合理使用 Singleflight：
- 对于相同的请求，使用 Singleflight 避免重复请求
- 适用于高并发场景下的缓存穿透问题

4. 超时控制：
- 设置合适的连接超时和读写超时
- 使用 context 控制请求生命周期

5. 启用 Keep-Alive：
- 默认启用，可以复用连接
- 通过 `IdleConnTimeout` 控制空闲连接的生命周期

## 总结

Freedom 框架的 HTTP 客户端提供了丰富的功能和灵活的扩展机制：

1. 支持 HTTP/1.1 和 HTTP/2 (H2C)
2. 强大的中间件系统
3. 可自定义的客户端配置
4. 内置的追踪和监控功能
5. 并发控制和性能优化

通过合理使用这些特性，可以构建出高性能、可靠的 HTTP 客户端应用。在实际使用中，建议：

- 在 Repository 层封装 HTTP 请求
- 使用中间件处理通用逻辑
- 正确处理错误和超时
- 根据实际需求优化性能配置
