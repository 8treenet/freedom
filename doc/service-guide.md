# Service 使用指南

## 基本定义

Service 是领域服务层，用于实现业务逻辑和协调领域对象。

### Service 结构

```go
type CartService struct {
    Worker      freedom.Worker         // 请求运行时
    CartRepo    dependency.CartRepo    // 资源库接口
    CartFactory *aggregate.CartFactory // 聚合工厂
}
```

### 注册 Service

```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 绑定 Service
        initiator.BindService(func() *CartService {
            return &CartService{}
        })

        // 注入到 Controller
        initiator.InjectController(func(ctx freedom.Context) (service *CartService) {
            initiator.FetchService(ctx, &service)
            return
        })
    })
}
```

---

## 依赖注入

Freedom 框架通过依赖注入自动管理 Service 的依赖。

### 可注入的依赖类型

| 类型 | 说明 | 示例 |
|------|------|------|
| Worker | 请求运行时，一个请求绑定一个 | `Worker freedom.Worker` |
| Repository | 资源库接口或实现 | `CartRepo dependency.CartRepo` |
| Factory | 聚合工厂 | `CartFactory *aggregate.CartFactory` |
| Infrastructure | 基础设施组件 | `EventTransaction *domainevent.EventTransaction` |

### Worker 使用

```go
func (s *CartService) SomeMethod() {
    // 日志记录
    s.Worker.Logger().Info("log message")

    // 获取请求上下文
    ctx := s.Worker.IrisContext()
}
```

### Repository 使用

```go
type OrderService struct {
    OrderRepo dependency.OrderRepo
    GoodsRepo dependency.GoodsRepo
}
```

### Factory 使用

```go
func (c *CartService) Add(userID, goodsID, goodsNum int) error {
    // 创建命令
    cmd, err := c.CartFactory.NewCartAddCmd(goodsID, userID)
    if err != nil {
        return err
    }
    return cmd.Run(goodsNum)
}
```

---

## 事务使用

事务组件保证多个数据库操作的原子性。

### 事务注入

```go
type OrderService struct {
    EventTransaction *domainevent.EventTransaction // 事务组件
}
```

### 基本事务

```go
func (srv *OrderService) CreateOrder() error {
    return srv.EventTransaction.Execute(func() error {
        // 闭包内所有操作在同一事务中
        if err := srv.OrderRepo.Create(); err != nil {
            return err // 返回错误触发回滚
        }
        return nil // 返回 nil 提交事务
    })
}
```

### 复杂业务事务

```go
func (srv *OrderService) Shop(goodsID, num, userID int) error {
    goodsEntity, err := srv.GoodsRepo.Get(goodsID)
    if err != nil {
        return err
    }

    if goodsEntity.Stock < num {
        return errors.New("库存不足")
    }

    goodsEntity.AddStock(-num)

    // 添加领域事件
    goodsEntity.AddPubEvent(&event.ShopGoods{
        UserID:    userID,
        GoodsID:   goodsID,
        GoodsNum:  num,
        GoodsName: goodsEntity.Name,
    })

    // 事务保证：修改库存 + 创建订单 + 记录事件
    return srv.EventTransaction.Execute(func() error {
        if err := srv.GoodsRepo.Save(goodsEntity); err != nil {
            return err
        }
        return srv.OrderRepo.Create(goodsEntity.ID, num, userID)
    })
}
```

### 事务特性

| 特性 | 说明 |
|------|------|
| 自动回滚 | Execute 闭包返回 error 时自动回滚 |
| 自动提交 | 返回 nil 时自动提交 |
| 同一连接 | 所有操作使用同一事务连接 |
| 嵌套支持 | 支持事务嵌套使用 |
| 事件集成 | 事务成功后自动推送领域事件 |

---

## 与工厂配合使用

Service 通过工厂创建聚合，实现业务逻辑编排。

### 命令模式（写操作）

```go
func (c *CartService) Add(userID, goodsID, goodsNum int) error {
    // 创建命令
    cmd, err := c.CartFactory.NewCartAddCmd(goodsID, userID)
    if err != nil {
        return err
    }
    // 执行命令
    return cmd.Run(goodsNum)
}
```

### 查询模式（读操作）

```go
func (c *CartService) Items(userID int) (json.Marshaler, error) {
    // 创建查询
    query, err := c.CartFactory.NewCartItemQuery(userID)
    if err != nil {
        return nil, err
    }
    return query, nil
}
```

---

## 最佳实践

| 原则 | 说明 |
|------|------|
| 职责单一 | Service 只实现特定领域的业务逻辑 |
| 依赖倒置 | 通过接口解耦各个组件 |
| 合理事务 | 使用事务保证数据一致性 |
| Worker 使用 | 使用 Worker 处理日志、上下文等 |
| 公开字段 | 依赖字段必须公开（首字母大写） |
| DDD 原则 | 合理划分领域边界，遵循 CQS 模式 |
