# Freedom 路由指南

## 简介

Freedom 框架集成了 Iris Web 框架的路由功能，并在此基础上提供了更符合 DDD 架构风格的路由注册和处理方式。本指南将介绍如何在 Freedom 项目中配置和使用路由功能。

## 路由注册

### 1. Controller 定义

```go
// adapter/controller/user_controller.go
package controller

import (
    "github.com/8treenet/freedom"
    "github.com/kataras/iris/v12"
)

func init() {
    // 注册 Controller
    freedom.Prepare(func(initiator freedom.Initiator) {
        initiator.BindController("/api/v1", &UserController{})
    })
}

// UserController 用户控制器
type UserController struct {
    Worker freedom.Worker
    // 依赖注入的 Service
    UserService *service.UserService
}

// Get 处理 GET 请求
// 路由: GET /api/v1/users/:id
func (c *UserController) Get(ctx iris.Context) {
    userID, _ := ctx.Params().GetInt("id")
    user, err := c.UserService.GetUserByID(userID)
    if err != nil {
        ctx.StatusCode(iris.StatusNotFound)
        return
    }
    ctx.JSON(user)
}

// Post 处理 POST 请求
// 路由: POST /api/v1/users
func (c *UserController) Post(ctx iris.Context) {
    var req struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := ctx.ReadJSON(&req); err != nil {
        ctx.StatusCode(iris.StatusBadRequest)
        return
    }
    
    err := c.UserService.CreateUser(req.Name, req.Email)
    if err != nil {
        ctx.StatusCode(iris.StatusInternalServerError)
        return
    }
    ctx.StatusCode(iris.StatusCreated)
}
```

### 2. 路由规则

Freedom 框架支持以下路由注册方式：

```go
// 1. RESTful 风格路由（自动根据方法名映射）
type UserController struct {
    // ...
}

func (c *UserController) Get()    {} // GET    /api/v1/users
func (c *UserController) Post()   {} // POST   /api/v1/users
func (c *UserController) Put()    {} // PUT    /api/v1/users
func (c *UserController) Delete() {} // DELETE /api/v1/users

// 2. 自定义方法路由
func (c *UserController) GetProfile() {}    // GET    /api/v1/users/profile
func (c *UserController) PostAvatar() {}    // POST   /api/v1/users/avatar
func (c *UserController) PutPassword() {}   // PUT    /api/v1/users/password
func (c *UserController) DeleteCache() {}   // DELETE /api/v1/users/cache

// 3. 带参数的路由
func (c *UserController) GetBy(id int) {}           // GET /api/v1/users/{id}
func (c *UserController) GetProfileBy(id int) {}    // GET /api/v1/users/{id}/profile
```

### 3. 中间件使用

```go
// middleware/auth.go
package middleware

import (
    "github.com/kataras/iris/v12"
)

func AuthMiddleware(ctx iris.Context) {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        ctx.StatusCode(iris.StatusUnauthorized)
        return
    }
    // 验证 token ...
    ctx.Next()
}

// 在 controller 中使用中间件
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 全局中间件
        initiator.BindMiddleware(AuthMiddleware)
        
        // 特定路由组中间件
        initiator.BindController("/api/v1", &UserController{}, AuthMiddleware)
    })
}
```

## 请求处理

### 1. 参数获取

```go
// 1. 路径参数
func (c *UserController) GetBy(id int) {
    // id 参数会自动解析并注入
    user, err := c.UserService.GetUserByID(id)
    // ...
}

// 2. 查询参数
func (c *UserController) Get(ctx iris.Context) {
    page := ctx.URLParamIntDefault("page", 1)
    size := ctx.URLParamIntDefault("size", 10)
    // ...
}

// 3. JSON 请求体
func (c *UserController) Post(ctx iris.Context) {
    var req struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := ctx.ReadJSON(&req); err != nil {
        ctx.StatusCode(iris.StatusBadRequest)
        return
    }
    // ...
}

// 4. 表单数据
func (c *UserController) PostForm(ctx iris.Context) {
    name := ctx.FormValue("name")
    file, _, err := ctx.FormFile("avatar")
    if err != nil {
        ctx.StatusCode(iris.StatusBadRequest)
        return
    }
    // ...
}
```

### 2. 响应处理

```go
// 1. JSON 响应
func (c *UserController) Get(ctx iris.Context) {
    user, err := c.UserService.GetUser()
    if err != nil {
        ctx.JSON(iris.Map{
            "error": err.Error(),
        })
        return
    }
    ctx.JSON(user)
}

// 2. 自定义状态码
func (c *UserController) Post(ctx iris.Context) {
    err := c.UserService.CreateUser()
    if err != nil {
        ctx.StatusCode(iris.StatusBadRequest)
        ctx.JSON(iris.Map{
            "error": err.Error(),
        })
        return
    }
    ctx.StatusCode(iris.StatusCreated)
}

// 3. 文件下载
func (c *UserController) GetExport(ctx iris.Context) {
    data := c.UserService.ExportData()
    ctx.Header("Content-Disposition", "attachment; filename=export.xlsx")
    ctx.Binary(data)
}
```

## 最佳实践

### 1. 路由组织

```go
// 按业务领域组织路由
initiator.BindController("/api/v1/users", &UserController{})
initiator.BindController("/api/v1/orders", &OrderController{})
initiator.BindController("/api/v1/products", &ProductController{})

// 版本控制
initiator.BindController("/api/v1", &V1Controller{})
initiator.BindController("/api/v2", &V2Controller{})
```

### 2. 错误处理

```go
// common/errors.go
var (
    ErrNotFound = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
)

// controller/base.go
type BaseController struct {
    Worker freedom.Worker
}

func (c *BaseController) HandleError(ctx iris.Context, err error) {
    switch err {
    case ErrNotFound:
        ctx.StatusCode(iris.StatusNotFound)
    case ErrInvalidInput:
        ctx.StatusCode(iris.StatusBadRequest)
    default:
        ctx.StatusCode(iris.StatusInternalServerError)
    }
    ctx.JSON(iris.Map{
        "error": err.Error(),
    })
}
```

### 3. 参数验证

```go
// 使用结构体标签进行验证
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=0,lte=150"`
}

func (c *UserController) Post(ctx iris.Context) {
    var req CreateUserRequest
    if err := ctx.ReadJSON(&req); err != nil {
        ctx.StatusCode(iris.StatusBadRequest)
        return
    }
    
    if err := c.Worker.Validator().Struct(req); err != nil {
        ctx.StatusCode(iris.StatusBadRequest)
        ctx.JSON(iris.Map{
            "error": err.Error(),
        })
        return
    }
    // ...
}
```

## 高级特性

### 1. 自定义路由绑定

```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        app := initiator.Iris()
        
        // 自定义路由组
        api := app.Party("/api")
        {
            api.Use(AuthMiddleware)
            api.Get("/health", HealthCheck)
            api.Post("/webhook", WebhookHandler)
        }
    })
}
```

### 2. BeforeActivation 自定义路由

在 Controller 中实现 `BeforeActivation` 方法可以自定义更灵活的路由规则：

```go
// Controller 实现 BeforeActivation 接口
type DefaultController struct {
    Worker freedom.Worker
}

// BeforeActivation 自定义路由规则
func (c *DefaultController) BeforeActivation(b freedom.BeforeActivation) {
    // 自定义复杂路径参数
    b.Handle("GET", "/customPath/{id:int64}/{uid:int}/{username:string}", "CustomPath")
    
    // 支持任意 HTTP 方法的路由
    b.Handle("ANY", "/custom", "Custom")
    
    // 或者指定具体的 HTTP 方法
    b.Handle("GET", "/users/{id:int}", "GetUser")
    b.Handle("POST", "/users/batch", "BatchCreate")
    b.Handle("PUT", "/users/{id:int}/status", "UpdateStatus")
    b.Handle("DELETE", "/users/{id:int}", "DeleteUser")
}

// Custom handles the ANY: /custom route.
func (c *Default) Custom() freedom.Result {
	//curl http://127.0.0.1:8000/custom
	//curl -X PUT http://127.0.0.1:8000/custom
	//curl -X DELETE http://127.0.0.1:8000/custom
	//curl -X POST http://127.0.0.1:8000/custom
	method := c.Worker.IrisContext().Request().Method
	c.Worker.Logger().Info("CustomHello", freedom.LogFields{"method": method})
	return &infra.JSONResponse{Object: method + "/Custom"}
}


// CustomPath handles the GET: /customPath/{id:int64}/{uid:int}/{username:string} route.
func (c *DefaultController) CustomPath(id int64, uid int, username string) {
    id, _ := ctx.Params().GetInt64("id")
    uid, _ := ctx.Params().GetInt("uid")
    username := ctx.Params().Get("username")
    
    ctx.JSON(iris.Map{
        "id": id,
        "uid": uid,
        "username": username,
    })
}
```

路径参数支持的类型：
- `{name:string}` - 字符串参数
- `{id:int}` - 整数参数
- `{id:int64}` - 64位整数参数
- `{id:uint}` - 无符号整数参数
- `{id:uint64}` - 64位无符号整数参数
- `{name:alphabetical}` - 仅字母参数
- `{name:file}` - 文件路径参数
- `{name:path}` - 路径参数

### 3. WebSocket 支持

```go
func (c *ChatController) GetWs(ctx iris.Context) {
    ws := websocket.New(websocket.Config{})
    ws.OnConnection(func(c websocket.Connection) {
        // 处理 WebSocket 连接
        c.OnMessage(func(data []byte) {
            // 处理消息
        })
        
        c.OnDisconnect(func() {
            // 处理断开连接
        })
    })
    ws.Upgrade(ctx)
}
```

### 2. 消息队列集成

#### Kafka 配置和初始化

```go
// server/main.go
func main() {
    app := freedom.NewApplication()
    
    // Kafka 配置
    kafkaConf := sarama.NewConfig()
    kafkaConf.Version = sarama.V0_11_0_2

    // 启动消费者
    kafka.GetConsumer().Start(
        []string{":9092"},      // Kafka 服务器地址
        "freedom1",             // 消费者组
        kafkaConf,             // Kafka 配置
        "http://127.0.0.1:8000", // 回调地址,当前服务端口号，建议runner使用app.NewH2CRunner
        true,                  // 是否为H2C代理
    )

    // ... 其他配置
    app.Run(addrRunner, conf.Get().App)
}
```

#### 事件监听和处理

```go
// adapter/controller/goods_controller.go
package controller

import (
    "github.com/8treenet/freedom"
    "github.com/8treenet/freedom/example/infra-example/domain/event"
)

func init() {
    // 注册 Controller
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 绑定控制器
        initiator.BindController("/api/v1", &GoodsController{})
        
        // 监听 Kafka 事件
        // 参数1: Topic名称
        // 参数2: 处理方法名 "控制器名称.方法名"
        initiator.ListenEvent((&event.ShopGoods{}).Topic(), "GoodsController.PostShop")
    })
}

// GoodsController 商品控制器
type GoodsController struct {
    Worker freedom.Worker              //依赖倒置
    GoodsService *service.GoodsService //依赖注入
}

// PostShop 处理商店商品事件
// 注意：方法名必须以 HTTP 方法开头（Post/Get/Put/Delete）
func (c *GoodsController) PostShop(shopGoods *event.ShopGoods) error {
    // 处理接收到的商品事件
    c.Worker.Logger().Infof("接收到商品事件: %+v", shopGoods)
    
    // 调用服务处理业务逻辑
    return c.GoodsService.HandleShopGoods(shopGoods)
}
```

#### 事件定义

```go
// domain/event/shop_goods.go
package event

// ShopGoods 商店商品事件
type ShopGoods struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    Price   float64 `json:"price"`
    ShopID  int     `json:"shop_id"`
}

// Topic 返回事件主题
func (s *ShopGoods) Topic() string {
    return "shop.goods"
}
```

## 注意事项

1. Controller 方法命名要符合规范，以 HTTP 方法开头
2. 路由参数类型要匹配，否则会导致路由注册失败
3. 合理使用中间件，避免性能问题
4. 请求参数验证要完整
5. 统一错误处理方式
6. 合理组织路由结构，避免路由冲突

## 更多资源

- [Iris 框架文档](https://www.iris-go.com/)
- [Freedom 示例项目](https://github.com/8treenet/freedom/tree/master/example)
- [API 最佳实践](https://github.com/8treenet/freedom/blob/master/example/base) 