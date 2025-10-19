# Freedom HTTP 客户端使用指南

Freedom 框架内置了一个功能强大且易于使用的 HTTP 客户端，支持 HTTP/1.1、HTTP/2（H2C）协议，提供了丰富的中间件系统和性能优化特性。

## 快速开始

```go
import "github.com/8treenet/freedom/infra/requests"

// GET 请求
resp := requests.NewHTTPRequest("https://api.example.com/users/123").
    Get().
    ToJSON(&userData)

// POST 请求
result := requests.NewHTTPRequest("https://api.example.com/users").
    SetJSONBody(userInfo).
    Post().
    ToJSON(&response)
```

## 创建请求

Freedom 提供了灵活的请求创建方式，既可以直接使用，也可以在 Repository 中集成。

### 直接创建

```go
// 标准 HTTP 请求
req := requests.NewHTTPRequest("https://api.example.com/users")

// HTTP/2 Cleartext (H2C) 请求，适合内网服务通信
req := requests.NewH2CRequest("http://internal.service:8080/api")
```

### Repository 中创建

```go
type UserService struct {
    freedom.Repository
}

func (s *UserService) GetUser(id string) {
    // 自动传递链路追踪信息
    req := s.NewHTTPRequest("https://api.example.com/users/"+id)
    req := s.NewH2CRequest("http://internal.service:8080/users/"+id)
}
```

## 请求构建

### HTTP 方法

```go
req.Get()      // 获取资源
req.Post()     // 创建资源
req.Put()      // 更新资源
req.Delete()   // 删除资源
req.Head()     // 获取头部信息
req.Options()  // 获取支持的方法
```

### 请求参数

```go
// 查询参数 - 自动处理 URL 编码
req.SetQueryParam("page", 1)
req.SetQueryParam("size", 20)
req.SetQueryParam("tags", []string{"go", "microservice"}) // 支持数组

// JSON 请求体
user := User{Name: "张三", Age: 25}
req.SetJSONBody(user)

// 表单数据
form := url.Values{}
form.Set("username", "admin")
form.Set("password", "secret")
req.SetFormBody(form)

// 文件上传
req.SetFile("avatar", "/path/to/avatar.jpg")

// 原始数据
req.SetBody([]byte(`{"raw": "data"}`))
```

### 请求头

```go
// 单个请求头
req.AddHeader("Authorization", "Bearer token123")

// 批量设置
headers := http.Header{}
headers.Set("Content-Type", "application/json")
headers.Set("X-API-Version", "v1")
req.SetHeader(headers)
```

## 响应处理

### 多种响应格式

```go
// JSON 响应
var user User
response := req.Get().ToJSON(&user)

// 字符串响应
content, response := req.Get().ToString()

// 字节数组
data, response := req.Get().ToBytes()

// XML 响应
var config Config
response := req.Get().ToXML(&config)
```

### 错误处理

```go
response := req.Get().ToJSON(&result)
if response.Error != nil {
    // 网络错误、超时、JSON 解析错误等
    log.Printf("请求失败: %v", response.Error)
    return
}

if response.StatusCode >= 400 {
    log.Printf("服务器返回错误: %d", response.StatusCode)
    return
}
```

## 中间件系统

Freedom 的中间件系统提供了强大的请求处理管道能力，支持日志、监控、重试等功能。

### 中间件接口

```go
type Middleware interface {
    Next()                      // 继续执行下一个中间件
    Stop(err ...error)         // 停止执行并返回错误
    GetRequest() *http.Request // 获取当前请求
    GetRespone() *Response     // 获取响应对象
    Context() context.Context  // 获取请求上下文
    // ... 更多方法
}
```

### 安装中间件

```go
// 全局安装，影响所有请求
requests.InstallMiddleware(
    NewLogMiddleware(),
    NewTraceMiddleware(),
    NewRetryMiddleware(3),
)
```

### 实用中间件示例

#### 请求日志中间件

```go
func NewLogMiddleware() requests.Handler {
    return func(middle requests.Middleware) {
        start := time.Now()
        req := middle.GetRequest()

        fmt.Printf("➡️  %s %s\n", req.Method, req.URL.String())

        middle.Next() // 执行后续中间件和请求

        resp := middle.GetRespone()
        duration := time.Since(start)
        fmt.Printf("⬅️  %s %s %d (%v)\n",
            req.Method, req.URL, resp.StatusCode, duration)
    }
}
```

#### 监控中间件

```go
func NewMetricsMiddleware() requests.Handler {
    return func(middle requests.Middleware) {
        start := time.Now()
        middle.Next()

        req := middle.GetRequest()
        resp := middle.GetRespone()
        duration := time.Since(start)

        // 记录监控指标
        metrics.RecordHTTPRequest(
            req.URL.Host,
            req.Method,
            resp.StatusCode,
            duration,
        )
    }
}
```

#### 重试中间件

```go
func NewRetryMiddleware(maxRetries int) requests.Handler {
    return func(middle requests.Middleware) {
        for attempt := 0; attempt < maxRetries; attempt++ {
            middle.Next()
            resp := middle.GetRespone()

            // 成功或客户端错误，不重试
            if resp.Error == nil && resp.StatusCode < 500 {
                return
            }

            if attempt < maxRetries-1 {
                // 指数退避
                backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
                time.Sleep(backoff)
            }
        }

        middle.Stop(fmt.Errorf("重试 %d 次后仍然失败", maxRetries))
    }
}
```

## 客户端配置

Freedom 允许自定义 HTTP 客户端配置，以满足不同场景的需求。

### 快速配置

```go
// 初始化客户端，设置超时时间
func InitClients() {
    // HTTP 客户端：读写超时 10 秒，连接超时 2 秒
    requests.InitHTTPClient(10*time.Second, 2*time.Second)

    // H2C 客户端：读写超时 5 秒，连接超时 1 秒
    requests.InitH2CClient(5*time.Second, 1*time.Second)
}
```

### 自定义 HTTP 客户端

```go
type CustomClient struct {
    *requests.ClientImpl
}

func NewProductionHTTPClient() *CustomClient {
    transport := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        DialContext: (&net.Dialer{
            Timeout:   3 * time.Second,  // 连接超时
            KeepAlive: 30 * time.Second, // 保持连接
            DualStack: true,            // IPv4/IPv6 双栈
        }).DialContext,

        // 连接池配置
        MaxIdleConns:        1000,             // 全局最大空闲连接
        MaxIdleConnsPerHost: 100,              // 每个主机最大空闲连接
        IdleConnTimeout:     90 * time.Second, // 空闲连接超时

        // TLS 配置
        TLSHandshakeTimeout: 5 * time.Second,
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false, // 生产环境建议设为 false
        },

        // 性能优化
        ForceAttemptHTTP2:     true,              // 尝试使用 HTTP/2
        ExpectContinueTimeout: 1 * time.Second,   // 100-continue 超时
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   15 * time.Second, // 总请求超时
    }

    return &CustomClient{
        ClientImpl: &requests.ClientImpl{Client: client},
    }
}

// 设置全局客户端
func init() {
    customClient := NewProductionHTTPClient()
    requests.SetHTTPClient(customClient)
}
```

### 自定义 H2C 客户端

H2C（HTTP/2 Cleartext）适合内网服务间通信，提供更好的性能。

```go
func NewInternalH2CClient() *CustomClient {
    transport := &http2.Transport{
        AllowHTTP: true, // 允许明文 HTTP/2
        DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
            // 自定义连接逻辑
            dialer := &net.Dialer{Timeout: 2 * time.Second}
            return dialer.Dial(network, addr)
        },
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   8 * time.Second,
    }

    return &CustomClient{
        ClientImpl: &requests.ClientImpl{Client: client},
    }
}

func init() {
    h2cClient := NewInternalH2CClient()
    requests.SetH2CClient(h2cClient)
}
```

### 针对单个请求使用自定义客户端

```go
// 为特定请求使用自定义客户端
customClient := NewProductionHTTPClient()

response := requests.NewHTTPRequest("https://api.example.com").
    SetClient(customClient).
    Get().
    ToJSON(&result)
```

## 高级特性

### 请求追踪

Freedom 内置了详细的 HTTP 连接追踪功能，帮助你分析网络性能。

```go
// 启用追踪
response := requests.NewHTTPRequest("https://api.example.com/data").
    EnableTrace().
    Get().
    ToJSON(&result)

// 获取详细的追踪信息
trace := response.TraceInfo()
fmt.Printf("总耗时: %v\n", trace.TotalTime)
fmt.Printf("DNS 查询: %v\n", trace.DNSLookup)
fmt.Printf("TCP 连接: %v\n", trace.TCPConnTime)
fmt.Printf("TLS 握手: %v\n", trace.TLSHandshake)
fmt.Printf("服务器处理: %v\n", trace.ServerTime)
fmt.Printf("连接复用: %v\n", trace.IsConnReused)

// 追踪信息结构
type HTTPTraceInfo struct {
    DNSLookup    time.Duration // DNS 解析时间
    ConnTime     time.Duration // 连接建立时间
    TCPConnTime  time.Duration // TCP 连接时间
    TLSHandshake time.Duration // TLS 握手时间
    ServerTime   time.Duration // 服务器处理时间
    ResponseTime time.Duration // 响应接收时间
    TotalTime    time.Duration // 总耗时
    IsConnReused bool          // 是否复用连接
}
```

### 并发控制

#### Singleflight 机制

在高并发场景下，多个相同的请求可能会同时发起。Freedom 的 Singleflight 机制可以防止重复请求，提高效率。

```go
// 场景：100个用户同时请求同一个用户信息
func GetUserInfo(userID string) (*UserInfo, error) {
    var result struct {
        Code int       `json:"code"`
        Data UserInfo  `json:"data"`
        Msg  string    `json:"msg"`
    }

    // 使用 Singleflight，相同的 key 只会发起一次实际请求
    // 其他并发请求会等待并共享结果
    response := requests.NewHTTPRequest("https://api.example.com/users/"+userID).
        Get().
        Singleflight("user_info", userID). // 缓存键："user_info" + userID
        ToJSON(&result)

    if response.Error != nil {
        return nil, response.Error
    }

    if result.Code != 0 {
        return nil, fmt.Errorf(result.Msg)
    }

    return &result.Data, nil
}

// 并发测试
func TestSingleflight(t *testing.T) {
    var wg sync.WaitGroup

    // 100个并发请求
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            user, err := GetUserInfo("123")
            if err != nil {
                t.Error(err)
            }
            t.Logf("获取用户: %+v", user)
        }()
    }

    wg.Wait()
    // 实际只会发起一次 HTTP 请求
}
```

#### 适用场景

- **缓存穿透保护**：防止热点数据的重复请求
- **API 限流**：减少对下游服务的压力
- **数据一致性**：确保同一时间窗口内的请求获得相同结果

### 上下文传递

Freedom 支持标准的 Go context，便于请求取消和超时控制。

```go
// 设置超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response := requests.NewHTTPRequest("https://api.example.com/slow-api").
    WithContext(ctx).
    Get().
    ToJSON(&result)

// 手动取消
ctx, cancel = context.WithCancel(context.Background())
go func() {
    // 3秒后取消请求
    time.Sleep(3 * time.Second)
    cancel()
}()

response = requests.NewHTTPRequest("https://api.example.com/long-task").
    WithContext(ctx).
    Get().
    ToJSON(&result)

if response.Error == context.Canceled {
    fmt.Println("请求被取消")
}
```


## 实践建议

### Repository 模式

在 Freedom 中，推荐将 HTTP 请求封装在 Repository 层，这样可以：

- 统一管理外部服务调用
- 便于单元测试和 Mock
- 实现更好的关注点分离

```go
type UserRepository struct {
    freedom.Repository
}

func (r *UserRepository) GetUser(id string) (*User, error) {
    var resp struct {
        Code int    `json:"code"`
        Data User   `json:"data"`
        Msg  string `json:"msg"`
    }

    // Repository 自动传递链路追踪信息
    response := r.NewHTTPRequest("https://api.example.com/users/"+id).
        Get().
        EnableTrace().
        ToJSON(&resp)

    if response.Error != nil {
        return nil, fmt.Errorf("获取用户失败: %w", response.Error)
    }

    if resp.Code != 0 {
        return nil, fmt.Errorf("业务错误: %s", resp.Msg)
    }

    return &resp.Data, nil
}

func (r *UserRepository) CreateUser(user User) error {
    response := r.NewHTTPRequest("https://api.example.com/users").
        SetJSONBody(user).
        Post().
        ToJSON(&struct{}{})

    return response.Error
}
```

### 错误处理策略

#### 分层错误处理

```go
// 1. 网络层错误
if response.Error != nil {
    // 记录网络错误，可能需要重试或降级
    log.Printf("网络请求失败: %v", response.Error)
    return fmt.Errorf("服务暂时不可用")
}

// 2. HTTP 状态码错误
switch response.StatusCode {
case 400:
    return fmt.Errorf("请求参数错误")
case 401:
    return fmt.Errorf("未授权访问")
case 404:
    return fmt.Errorf("资源不存在")
case 500:
    return fmt.Errorf("服务器内部错误")
default:
    if response.StatusCode >= 400 {
        return fmt.Errorf("请求失败: %d", response.StatusCode)
    }
}

// 3. 业务逻辑错误
if result.Code != 0 {
    switch result.Code {
    case 1001:
        return fmt.Errorf("用户不存在")
    case 1002:
        return fmt.Errorf("权限不足")
    default:
        return fmt.Errorf("业务错误: %s", result.Msg)
    }
}
```

#### 中间件中的错误处理

```go
func NewErrorHandlingMiddleware() requests.Handler {
    return func(middle requests.Middleware) {
        middle.Next()

        resp := middle.GetRespone()
        req := middle.GetRequest()

        if resp.Error != nil {
            // 记录请求失败的详细信息
            log.Printf("请求失败 %s %s: %v",
                req.Method, req.URL.String(), resp.Error)
            return
        }

        // 监控错误率
        if resp.StatusCode >= 400 {
            metrics.RecordHTTPError(req.URL.Host, resp.StatusCode)
        }
    }
}
```

### 性能优化建议

#### 1. 连接池优化

```go
// 高并发场景的连接池配置
func NewHighPerformanceClient() *http.Client {
    transport := &http.Transport{
        // 连接池大小
        MaxIdleConns:        1000,             // 全局最大空闲连接
        MaxIdleConnsPerHost: 200,              // 每个主机最大空闲连接
        MaxConnsPerHost:     500,              // 每个主机最大连接数

        // 超时设置
        IdleConnTimeout:     90 * time.Second, // 空闲连接超时
        DialContext: (&net.Dialer{
            Timeout:   3 * time.Second,  // 连接超时
            KeepAlive: 30 * time.Second, // TCP KeepAlive
        }).DialContext,

        // TLS 优化
        TLSHandshakeTimeout: 3 * time.Second,

        // HTTP/2 支持
        ForceAttemptHTTP2: true,
    }

    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second, // 总超时
    }
}
```

#### 2. 选择合适的协议

- **外部 API**：使用标准 HTTPS（HTTP/1.1 或 HTTP/2）
- **内网服务**：使用 H2C（HTTP/2 Cleartext）
- **文件下载**：使用 HTTP/1.1，避免内存问题

#### 3. 监控和追踪

```go
// 全局监控中间件
func NewMonitoringMiddleware(serviceName string) requests.Handler {
    return func(middle requests.Middleware) {
        start := time.Now()
        middle.Next()

        req := middle.GetRequest()
        resp := middle.GetRespone()
        duration := time.Since(start)

        // Prometheus 指标
        metrics.HTTPRequestTotal.WithLabelValues(
            serviceName,
            req.URL.Host,
            req.Method,
            strconv.Itoa(resp.StatusCode),
        ).Inc()

        metrics.HTTPRequestDuration.WithLabelValues(
            serviceName,
            req.URL.Host,
        ).Observe(duration.Seconds())

        // 慢请求告警
        if duration > 2*time.Second {
            log.Printf("慢请求警告: %s %s 耗时 %v",
                req.Method, req.URL.String(), duration)
        }
    }
}
```

### 调试技巧

#### 1. 开启详细日志

```go
func NewDebugMiddleware() requests.Handler {
    return func(middle requests.Middleware) {
        req := middle.GetRequest()

        // 记录请求详情
        log.Printf("➡️  %s %s", req.Method, req.URL.String())
        if req.Body != nil {
            body, _ := io.ReadAll(req.Body)
            log.Printf("请求体: %s", string(body))
            req.Body = io.NopCloser(bytes.NewReader(body))
        }

        middle.Next()

        resp := middle.GetRespone()
        log.Printf("⬅️  %s %s %d", req.Method, req.URL, resp.StatusCode)

        // 记录追踪信息
        if trace := resp.TraceInfo(); trace != nil {
            log.Printf("追踪信息: 总耗时=%v DNS=%v 连接=%v",
                trace.TotalTime, trace.DNSLookup, trace.ConnTime)
        }
    }
}
```

#### 2. 测试技巧

```go
func TestHTTPClient(t *testing.T) {
    // 使用 httptest 创建测试服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"code": 0, "data": {"name": "test"}}`))
    }))
    defer server.Close()

    // 测试请求
    var result struct {
        Code int `json:"code"`
        Data struct {
            Name string `json:"name"`
        } `json:"data"`
    }

    response := requests.NewHTTPRequest(server.URL).
        Get().
        ToJSON(&result)

    assert.NoError(t, response.Error)
    assert.Equal(t, 0, result.Code)
    assert.Equal(t, "test", result.Data.Name)
}
```

## 常见问题

### Q: 如何处理大文件下载？

```go
// 流式处理，避免内存占用过大
func DownloadFile(url, filepath string) error {
    response := requests.NewHTTPRequest(url).
        Get()

    if response.Error != nil {
        return response.Error
    }

    // 创建文件
    file, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer file.Close()

    // 流式写入
    _, err = io.Copy(file, response.Body)
    return err
}
```

### Q: 如何实现断点续传？

```go
func DownloadWithResume(url, filepath string) error {
    // 检查已下载的文件大小
    var startByte int64
    if stat, err := os.Stat(filepath); err == nil {
        startByte = stat.Size()
    }

    response := requests.NewHTTPRequest(url).
        AddHeader("Range", fmt.Sprintf("bytes=%d-", startByte)).
        Get()

    // 追加写入文件
    file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(file, response.Body)
    return err
}
```

### Q: 如何处理 Cookie？

```go
// 创建带 Cookie 的请求
jar, _ := cookiejar.New(nil)
client := &http.Client{
    Jar: jar,
}

requests.SetHTTPClient(&requests.ClientImpl{Client: client})

// Cookie 会自动处理
response := requests.NewHTTPRequest("https://example.com/login").
    SetJSONBody(LoginData{Username: "user", Password: "pass"}).
    Post()

// 后续请求会自动携带 Cookie
response = requests.NewHTTPRequest("https://example.com/profile").
    Get().
    ToJSON(&profile)
```

## 总结

Freedom HTTP 客户端提供了完整的企业级功能：

- ✅ **多协议支持**：HTTP/1.1、HTTP/2、H2C
- ✅ **中间件系统**：日志、监控、重试、认证等
- ✅ **性能优化**：连接池、Singleflight、追踪
- ✅ **易用性**：简洁的 API、Repository 集成
- ✅ **可靠性**：完善的错误处理和超时控制

合理使用这些特性，可以构建出高性能、可维护的 HTTP 客户端应用。
