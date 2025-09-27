# FShop 项目文档

## 1. 项目简介

FShop 是一个基于 Go 语言开发的电商系统示例，采用 DDD（领域驱动设计）架构模式实现。该项目展示了一个包含商品管理、购物车、订单处理等基本电商功能的系统。本项目主要用于展示如何在实际项目中应用 DDD 架构，以及如何使用 Freedom Framework 构建高可维护性的企业级应用。

### 1.1 核心特性

- 完整的电商业务流程实现
- DDD架构的最佳实践示例
- 领域事件驱动的业务处理
- 完善的测试用例
- 支持容器化部署
- 高性能的缓存设计
- 分布式追踪支持

### 1.2 项目依赖

```bash
# 必需组件
- Go 1.18+
- MySQL 5.7+
- Redis 6.0+
- Freedom Framework latest

# 可选组件
- Prometheus (用于监控)
- Docker (用于容器化部署)
```

## 2. 项目结构

```
fshop/
├── adapter/            # 适配器层
│   ├── controller/    # HTTP 控制器
│   │   ├── cart.go   # 购物车控制器
│   │   ├── goods.go  # 商品控制器
│   │   ├── order.go  # 订单控制器
│   │   └── delivery.go # 发货控制器
│   ├── repository/    # 数据仓储实现
│   │   ├── cart.go   # 购物车仓储
│   │   ├── goods.go  # 商品仓储
│   │   └── order.go  # 订单仓储
│   └── consumer/      # 消费者实现
│       └── consumer.go # 事件消费者
├── domain/            # 领域层
│   ├── aggregate/     # 聚合根
│   │   ├── cart_factory.go    # 购物车工厂
│   │   ├── shop_factory.go    # 购买工厂
│   │   └── order_factory.go   # 订单工厂
│   ├── entity/        # 实体
│   │   ├── cart.go   # 购物车实体
│   │   ├── goods.go  # 商品实体
│   │   └── order.go  # 订单实体
│   ├── vo/           # 值对象
│   │   ├── cart.go   # 购物车值对象
│   │   └── order.go  # 订单值对象
│   └── event/        # 领域事件
│       ├── shop.go   # 购买事件
│       └── order.go  # 订单事件
├── infra/            # 基础设施层
│   └── domainevent/  # 领域事件基础设施
├── config/           # 服务器配置和启动
├── api_test/         # API 测试用例
├── fshop.sql        # 数据库脚本
└── Dockerfile       # Docker 构建文件
```

## 3. 核心功能模块

### 3.1 商品模块 (GoodsService)

#### 3.1.1 主要功能
- 商品的创建和管理
- 商品库存管理
- 商品标签管理
- 商品列表查询
- 商品库存增减

#### 3.1.2 关键接口
```go
// 商品相关 API
GET    /goods/items      # 获取商品列表
POST   /goods           # 添加商品
PUT    /goods/stock/:id/:num  # 增加库存
PUT    /goods/tag       # 为商品打标签
POST   /goods/shop      # 直接购买商品
```

#### 3.1.3 示例请求
```json
// 添加商品请求
POST /goods
{
    "name": "示例商品",
    "price": 9900,
    "stock": 100,
    "tag": "热销"
}

// 商品列表响应
GET /goods/items
{
    "items": [
        {
            "id": 1,
            "name": "示例商品",
            "price": 9900,
            "stock": 100,
            "tag": "热销"
        }
    ],
    "total": 1
}
```

### 3.2 购物车模块 (CartService)

#### 3.2.1 主要功能
- 添加商品到购物车
- 查看购物车商品列表
- 清空购物车
- 购物车商品结算

#### 3.2.2 关键接口
```go
// 购物车相关 API
GET    /cart/items      # 获取购物车商品列表
POST   /cart           # 添加商品到购物车
DELETE /cart/all       # 清空购物车
POST   /cart/shop      # 购物车结算
```

#### 3.2.3 示例请求
```json
// 添加购物车
POST /cart
{
    "userId": 1,
    "goodsId": 1,
    "goodsNum": 2
}

// 购物车列表响应
GET /cart/items?userId=1
{
    "items": [
        {
            "goodsId": 1,
            "goodsName": "示例商品",
            "goodsNum": 2,
            "price": 9900
        }
    ]
}
```

### 3.3 订单模块 (OrderService)

#### 3.3.1 主要功能
- 订单创建
- 订单支付
- 订单发货
- 订单列表查询

#### 3.3.2 关键接口
```go
// 订单相关 API
GET    /order/items     # 获取订单列表
POST   /order/pay      # 订单支付
POST   /delivery       # 订单发货
```

#### 3.3.3 示例请求
```json
// 订单支付
POST /order/pay
{
    "orderNo": "202312200001",
    "userId": 1
}

// 订单列表响应
GET /order/items?userId=1
{
    "orders": [
        {
            "orderNo": "202312200001",
            "totalPrice": 19800,
            "status": "待发货",
            "items": [
                {
                    "goodsId": 1,
                    "goodsName": "示例商品",
                    "num": 2
                }
            ]
        }
    ]
}
```

## 4. 领域驱动设计实现

### 4.1 聚合根设计

#### 4.1.1 购物车聚合根
```go
// CartFactory 示例
type CartFactory struct {
    UserRepo  dependency.UserRepo
    CartRepo  dependency.CartRepo
    GoodsRepo dependency.GoodsRepo
}

// 创建购物车聚合根
func (factory *CartFactory) NewCartAddCmd(goodsID, userID int) (*CartAddCmd, error) {
    // 实现细节...
}
```

#### 4.1.2 订单聚合根
```go
// OrderFactory 示例
type OrderFactory struct {
    OrderRepo dependency.OrderRepo
    UserRepo  dependency.UserRepo
}

// 创建订单聚合根
func (factory *OrderFactory) NewOrderPayCmd(orderNo string, userID int) (*OrderPayCmd, error) {
    // 实现细节...
}
```

### 4.2 领域事件

#### 4.2.1 事件定义
```go
// 购买商品事件
type ShopGoods struct {
    UserID    int
    OrderNO   string
    GoodsID   int
    GoodsNum  int
    GoodsName string
}

// 订单支付事件
type OrderPay struct {
    OrderNo string
    UserID  int
    Amount  int
}
```

#### 4.2.2 事件处理
```go
// 事件处理示例
func (g *GoodsService) ShopEvent(event *event.ShopGoods) error {
    // 处理购买事件...
    return nil
}
```

### 4.3 资源库模式

#### 4.3.1 资源库接口
```go
// 商品资源库接口
type GoodsRepo interface {
    Get(id int) (*entity.Goods, error)
    Save(goods *entity.Goods) error
    FindsByPage(page, pagesize int, tag string) ([]*entity.Goods, error)
}
```

#### 4.3.2 实现示例
```go
// 商品资源库实现
type GoodsRepository struct {
    freedom.Repository
}

func (repo *GoodsRepository) Get(id int) (*entity.Goods, error) {
    // 实现细节...
}
```

## 5. 技术特性

### 5.1 配置管理

#### 5.1.1 TOML 配置示例
```toml
[db]
addr = "root:password@tcp(localhost:3306)/fshop?charset=utf8mb4"
max_open_conns = 16
max_idle_conns = 8

[redis]
addr = "localhost:6379"
pool_size = 32
```

#### 5.1.2 YAML 配置示例
```yaml
db:
  addr: "root:password@tcp(localhost:3306)/fshop?charset=utf8mb4"
  max_open_conns: 16
  max_idle_conns: 8

redis:
  addr: "localhost:6379"
  pool_size: 32
```

### 5.2 中间件集成

#### 5.2.1 日志中间件
```go
// 日志中间件示例
func LoggerMiddleware(ctx freedom.Context) {
    // 实现细节...
}
```

#### 5.2.2 追踪中间件
```go
// 追踪中间件示例
func TracingMiddleware(ctx freedom.Context) {
    // 实现细节...
}
```

### 5.3 事务处理

#### 5.3.1 事务示例
```go
// 购物车结算事务
func (cmd *CartShopCmd) Shop() error {
    return cmd.tx.Execute(func() error {
        // 1. 修改商品库存
        // 2. 清空购物车
        // 3. 创建订单
        // 4. 记录事件
        return nil
    })
}
```

## 6. 最佳实践

### 6.1 代码规范
- 遵循 Go 官方代码规范
- DDD 分层架构严格划分
- 统一的错误处理机制
- 完善的日志记录

### 6.2 监控告警
- 接口响应时间监控
- 错误率监控
- 资源使用监控
- 业务指标监控

## 7. 常见问题

### 7.1 安装部署
Q: 如何处理依赖问题？
A: 使用 go mod tidy 更新依赖

Q: 配置文件找不到？
A: 检查 FSHOP-CONFIG 环境变量

### 7.2 开发相关
Q: 如何添加新的 API？
A: 在 controller 中添加新的处理方法，并在 init 中注册路由

Q: 如何处理并发问题？
A: 使用事务和锁机制确保数据一致性

## 8. 总结

FShop 项目是一个基于 DDD 架构的电商系统示例，展示了如何在 Go 语言中实现领域驱动设计。项目结构清晰，代码组织合理，实现了电商系统的核心功能，并且提供了完善的测试用例和部署支持。通过这个项目，开发者可以学习到：

1. DDD 架构在实际项目中的应用
2. Go 语言企业级应用开发最佳实践
3. 微服务设计和实现方法
4. 高并发场景下的性能优化
5. 分布式系统开发经验
