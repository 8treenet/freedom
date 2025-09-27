# Freedom 路由开发指南

## 概述

Freedom 框架基于 Iris Web 框架构建，提供了强大的路由功能和符合 DDD 架构风格的路由管理方式。本指南将帮助你快速掌握 Freedom 的路由系统，构建高性能的 Web API。

## 快速开始

### 全局 API 前缀设置

在 `main.go` 中设置全局 API 前缀，这样所有控制器都会自动添加 `/api` 前缀：

```go
// main.go
package main

import (
    "github.com/8treenet/freedom"
    "github.com/kataras/iris/v12"
)

func main() {
    app := freedom.NewApplication()
    
    // 全局设置 API 前缀
    app.InstallParty("/api")
}
```

### 基础路由设置

```go
// adapter/controller/user_controller.go
package controller

import (
    "github.com/8treenet/freedom"
    "github.com/8treenet/freedom/example/base/infra"
    "path/filepath"
)

func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 注册用户控制器到 /v1/users 路径
        initiator.BindController("/v1/users", &UserController{})
    })
}

// 标准响应结构体
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// 用户响应结构体
type UserResponse struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// 用户列表响应结构体
type UserListResponse struct {
    Users []UserResponse `json:"users"`
    Total int            `json:"total"`
}

// UserController 用户控制器
type UserController struct {
    Worker      freedom.Worker
    Request     *infra.Request
    UserService *service.UserService
}
```

## 路由规则详解

### 自定义路由配置

使用 `BeforeActivation` 方法可以自定义更复杂的路由规则：

```go
type UserController struct {
    Worker freedom.Worker
}

// BeforeActivation 自定义路由规则
func (c *UserController) BeforeActivation(b freedom.BeforeActivation) {
    // 自定义路径参数类型
    b.Handle("GET", "/{id:int64}/orders/{orderId:int}", "GetUserOrder")
    
    // 商品搜索路由 - 展示复杂查询参数解析
    b.Handle("GET", "/products/search", "SearchProducts")
    
    b.Handle("GET", "/search", "SearchUsers")
    b.Handle("GET", "/profile", "GetProfile")
    b.Handle("PUT", "/password", "SetPassword")
    
    // 复杂路径参数
    b.Handle("GET", "/{id:int}/{action:string}", "UserAction")
    
    // 文件上传路由
    b.Handle("POST", "/{id:int}/upload", "UploadAvatar")
}

// 对应的处理方法实现

// GetUserOrder 获取用户订单
// 路由: GET /api/v1/users/{id}/orders/{orderId}
func (c *UserController) GetUserOrder(id int64, orderId int) freedom.Result {
    order, err := c.UserService.GetUserOrder(id, orderId)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "success",
        Data:    order,
    }}
}

// SearchUsers 搜索用户
// 路由: GET /api/v1/users/search
func (c *UserController) SearchUsers() freedom.Result {
    var query struct {
        Keyword string `url:"keyword" validate:"required"`
    }
    if err := c.Request.ReadQuery(&query, true); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    users, err := c.UserService.SearchUsers(query.Keyword)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "success",
        Data:    users,
    }}
}

// GetProfile 获取用户资料
// 路由: GET /api/v1/users/profile
func (c *UserController) GetProfile() freedom.Result {
    profile, err := c.UserService.GetProfile()
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "success",
        Data:    profile,
    }}
}

// SetPassword 设置密码
// 路由: PUT /api/v1/users/password
func (c *UserController) SetPassword() freedom.Result {
    var req struct {
        OldPassword string `json:"old_password" validate:"required"`
        NewPassword string `json:"new_password" validate:"required,min=6"`
    }
    
    if err := c.Request.ReadJSON(&req, true); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    err := c.UserService.SetPassword(req.OldPassword, req.NewPassword)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "密码设置成功",
    }}
}

// UserAction 用户操作
// 路由: GET /api/v1/users/{id}/{action}
func (c *UserController) UserAction(id int, action string) freedom.Result {
    result, err := c.UserService.PerformAction(id, action)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "操作成功",
        Data:    result,
    }}
}

// UploadAvatar 上传头像
// 路由: POST /api/v1/users/{id}/upload
func (c *UserController) UploadAvatar(id int) freedom.Result {
    ctx := c.Worker.IrisContext()
    file, header, err := ctx.FormFile("avatar")
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    defer file.Close()
    
    filename := generateUniqueFilename(header.Filename)
    filepath := filepath.Join("uploads", "avatars", filename)
    
    if err := c.UserService.SaveAvatar(file, filepath); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    return &infra.JSONResponse{
        Object: Response{
            Code:    200,
            Message: "头像上传成功",
            Data: struct {
                URL string `json:"url"`
            }{
                URL: "/uploads/avatars/" + filename,
            },
        },
    }
}
```

## 请求处理

### 1. 参数获取

#### 路径参数
```go
// 自动注入路径参数 - 在 BeforeActivation 中定义的路由会自动注入参数
func (c *UserController) GetUserOrder(id int64, orderId int) freedom.Result {
    // id 和 orderId 参数会自动从 URL 中解析并注入
    order, err := c.UserService.GetUserOrder(id, orderId)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "success",
        Data:    order,
    }}
}
```

#### 查询参数 - 结构体解析 URL
```go
// GET 请求使用结构体解析 URL 查询参数 - 商品搜索示例
func (c *UserController) SearchProducts() freedom.Result {
    // 定义商品搜索查询参数结构体，静态语言因该更遵循类型，推荐使用。
    var query struct {
        Keyword    string   `url:"keyword" validate:"required"`
        Category   string   `url:"category"`
        MinPrice   float64  `url:"min_price"`
        MaxPrice   float64  `url:"max_price"`
        Tags       []string `url:"tags"` // 支持标签数组
        SortBy     string   `url:"sort_by"`
        Page       int      `url:"page" validate:"min=1"`
        PageSize   int      `url:"page_size" validate:"min=1,max=100"`
    }
    
    // 自动解析和验证查询参数
    if err := c.Request.ReadQuery(&query, true); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    // 处理商品搜索业务逻辑
    products, total, err := c.ProductService.SearchProducts(query)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "success",
        Data: struct {
            Products []Product `json:"products"`
            Total    int64     `json:"total"`
            Page     int       `json:"page"`
            PageSize int       `json:"page_size"`
        }{
            Products: products,
            Total:    total,
            Page:     query.Page,
            PageSize: query.PageSize,
        },
    }}
}
```


#### 请求体参数
```go
// JSON 请求体 - 使用 c.Request.ReadJSON()
func (c *UserController) SetPassword() freedom.Result {
    var req struct {
        OldPassword string `json:"old_password" validate:"required"`
        NewPassword string `json:"new_password" validate:"required,min=6"`
    }
    
    if err := c.Request.ReadJSON(&req, true); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    err := c.UserService.SetPassword(req.OldPassword, req.NewPassword)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "密码设置成功",
    }}
}
```

### 2. 响应处理

#### 响应格式说明

在 Freedom 框架中，控制器方法使用 `infra.JSONResponse` 和结构体响应，提供类型安全和一致性。

#### JSON 响应
```go
// 成功响应
func (c *UserController) GetProfile() freedom.Result {
    profile, err := c.UserService.GetProfile()
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    return &infra.JSONResponse{
        Object: Response{
            Code:    200,
            Message: "success",
            Data:    profile,
        },
    }
}

// 错误响应
func (c *UserController) GetUserOrder(id int64, orderId int) freedom.Result {
    order, err := c.UserService.GetUserOrder(id, orderId)
    if err != nil {
        return &infra.JSONResponse{
            Error: err,
            Object: Response{
                Code:    404,
                Message: "订单不存在",
            },
        }
    }
    
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "success",
        Data:    order,
    }}
}
```

#### 自定义状态码
```go
func (c *UserController) SetPassword() freedom.Result {
    var req struct {
        OldPassword string `json:"old_password" validate:"required"`
        NewPassword string `json:"new_password" validate:"required,min=6"`
    }
    
    if err := c.Request.ReadJSON(&req, true); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    err := c.UserService.SetPassword(req.OldPassword, req.NewPassword)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    // 返回 200 状态码
    return &infra.JSONResponse{
        Object: Response{
            Code:    200,
            Message: "密码设置成功",
        },
    }
}
```

## 中间件使用

```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 为特定路由组应用中间件
        initiator.BindController("/admin", &AdminController{}, func(ctx freedom.Context) {
			worker := freedom.ToWorker(ctx)
            //认证处理
			worker.Logger().Info("Hello middleware begin")
			ctx.Next()
			worker.Logger().Info("Hello middleware end")
		})
    })
}
```

## 参数验证

### 结构体验证

```go
// 定义验证规则
type SetPasswordRequest struct {
    OldPassword string `json:"old_password" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=6"`
}

func (c *UserController) SetPassword() freedom.Result {
    var req SetPasswordRequest
    if err := c.Request.ReadJSON(&req, true); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    // 处理业务逻辑
    err := c.UserService.SetPassword(req.OldPassword, req.NewPassword)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    return &infra.JSONResponse{Object: Response{
        Code:    200,
        Message: "密码设置成功",
    }}
}
```


## 高级特性

### 文件上传处理

```go
import (
    "errors"
    "path/filepath"
)

// PostAvatar 处理头像上传
// 路由: POST /api/v1/users/avatar
func (c *UserController) PostAvatar() freedom.Result {
    ctx := c.Worker.IrisContext()
    // 获取上传的文件
    file, header, err := ctx.FormFile("avatar")
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    defer file.Close()
    
    // 保存文件
    filename := generateUniqueFilename(header.Filename)
    filepath := filepath.Join("uploads", "avatars", filename)
    
    if err := c.UserService.SaveAvatar(file, filepath); err != nil {
        return &infra.JSONResponse{Error: err}
    }
    
    return &infra.JSONResponse{
        Object: Response{
            Code:    200,
            Message: "头像上传成功",
            Data: struct {
                URL string `json:"url"`
            }{
                URL: "/uploads/avatars/" + filename,
            },
        },
    }
}
```


## 最佳实践

### 1. 路由组织

```go
// 按业务模块组织路由，使用 BeforeActivation 自定义路由
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 用户相关路由
        initiator.BindController("/v1/users", &UserController{})
        
        // 订单相关路由
        initiator.BindController("/v1/orders", &OrderController{})
        
        // 商品相关路由
        initiator.BindController("/v1/products", &ProductController{})
        
        // 管理后台路由（需要认证）
        initiator.BindController("/v1/admin", &AdminController{}, func(ctx freedom.Context) {
            worker := freedom.ToWorker(ctx)
            // 认证处理
            worker.Logger().Info("Admin middleware begin")
            ctx.Next()
            worker.Logger().Info("Admin middleware end")
        })
    })
}
```

### 2. 版本控制

```go
// API 版本控制
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // v1 版本
        initiator.BindController("/v1/users", &V1UserController{})
        
        // v2 版本
        initiator.BindController("/v2/users", &V2UserController{})
    })
}
```





## 调试和监控

### 1. 请求日志

```go
// 请求日志中间件，仅示例中间件的用法，框架已集成.
func RequestLogMiddleware() freedom.Handler {
    return func(ctx freedom.Context) {
        start := time.Now()
        method := ctx.Method()
        path := ctx.Path()
        ip := ctx.RemoteAddr()
        
        ctx.Next()
        
        duration := time.Since(start)
        status := ctx.GetStatusCode()
        
        // 结构化日志
        logger.Info("HTTP请求", freedom.LogFields{
            "method": method,
            "path": path,
            "ip": ip,
            "status": status,
            "duration": duration.String(),
            "user_agent": ctx.GetHeader("User-Agent"),
        })
    }
}
```

### 2. 性能监控

```go
// 性能监控中间件，仅示例中间件的用法，框架已集成.
func PerformanceMiddleware() freedom.Handler {
    return func(ctx freedom.Context) {
        start := time.Now()
        
        ctx.Next()
        
        duration := time.Since(start)
        
        // 记录慢请求
        if duration > 1*time.Second {
            logger.Warn("慢请求", freedom.LogFields{
                "path": ctx.Path(),
                "method": ctx.Method(),
                "duration": duration.String(),
            })
        }
        
        // 发送到监控系统
        metrics.RecordRequestDuration(ctx.Path(), ctx.Method(), duration)
    }
}
```

## 总结

Freedom 框架的路由系统提供了强大而灵活的功能：

1. **简单易用**: 基于方法名的自动路由映射
2. **高度可定制**: 支持自定义路由规则和中间件
3. **性能优化**: 内置对象池和缓存支持
4. **安全可靠**: 完善的错误处理和参数验证
5. **易于扩展**: 支持 WebSocket、文件上传等高级特性

通过合理使用这些特性，你可以构建出高性能、可维护的 Web API 服务。记住始终遵循 RESTful 设计原则，保持 API 的一致性和可预测性。