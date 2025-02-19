# Service 使用指南

## 1. 基本使用

### 1.1 Service 定义

Service 是领域服务层,用于实现业务逻辑。在 Freedom 框架中,Service 的定义和使用非常简单:

```go
// 定义 Service
type DefaultService struct {
    Worker    freedom.Worker    // 依赖注入请求运行时
    DefRepo   *repository.Default   // 依赖注入资源库对象
    DefRepoIF repository.DefaultRepoInterface  // 也可以注入资源库接口
}

// 实现业务方法
func (s *Default) RemoteInfo() (result struct {
    IP string
    Ua string
}) {
    s.Worker.Logger().Info("I'm service") 
    result.IP = s.DefRepo.GetIP()
    result.Ua = s.DefRepoIF.GetUA()
    return
}
```

### 1.2 Service 注册

Service 需要在 init() 函数中注册到框架:

```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 绑定创建服务函数到框架
        initiator.BindService(func() *DefaultService {
            return &DefaultService{} 
        })
        
        // 注入到控制器
        initiator.InjectController(func(ctx freedom.Context) (service *DefaultService) {
            initiator.FetchService(ctx, &service)
            return
        })
    })
}
```

## 2. 依赖注入

Freedom 框架支持多种依赖注入方式:

### 2.1 Worker 注入

每个 Service 可以注入 Worker 对象,用于获取请求上下文、日志等:

```go
type UserService struct {
    Worker freedom.Worker // 运行时,一个请求绑定一个运行时
}

func (s *UserService) SomeMethod() {
    // 使用 Logger
    s.Worker.Logger().Info("some log")
    
    // 获取上下文
    ctx := s.Worker.IrisContext()
}
```

### 2.2 Repository 注入 

Service 可以注入 Repository 对象或接口:

```go
type OrderService struct {
    OrderRepo    dependency.OrderRepo    // 依赖倒置订单资源库
    GoodsRepo    dependency.GoodsRepo    // 依赖倒置商品资源库
}
```

### 2.3 Factory 注入

可以注入 Factory 用于创建聚合根等:

```go
type GoodsService struct {
    ShopFactory *aggregate.ShopFactory  // 依赖注入购买聚合根工厂
}
```

### 2.4 Infrastructure 注入

可以注入基础设施组件:

```go
type UserService struct {
    Transaction transaction.Transaction // 依赖注入事务组件
    CartFactory *aggregate.CartFactory //依赖注入购物车聚合根工厂
}
```

## 3. 事务使用

Freedom 提供了事务组件用于保证数据一致性:

### 3.1 事务组件注入

```go
type OrderService struct {
    Transaction transaction.Transaction //事务组件
}
```

### 3.2 事务使用示例

```go
// 简单事务示例
func (srv *OrderService) CreateOrder() error {
    return srv.Transaction.Execute(func() error {
        // 在这个闭包函数中的所有数据库操作都在同一个事务中
        if err := srv.OrderRepo.Create(); err != nil {
            return err // 返回错误会触发回滚
        }
        return nil // 返回 nil 会提交事务
    })
}

// 复杂业务事务示例
func (srv *OrderService) Shop(goodsID, num, userID int) error {
    goodsEntity, err := srv.GoodsRepo.Get(goodsID)
    if err != nil {
        return err
    }
    
    if goodsEntity.Stock < num {
        return errors.New("库存不足")
    }
    
    goodsEntity.AddStock(-num) // 扣减库存
    
    // 使用事务保证一致性:
    // 1. 修改商品库存
    // 2. 创建订单
    return srv.Transaction.Execute(func() error {
        if err := srv.GoodsRepo.Save(goodsEntity); err != nil {
            return err
        }
        
        return srv.OrderRepo.Create(goodsEntity.ID, num, userID)
    })
}
```

### 3.3 事务使用示例
通过工厂和聚合的合理使用,可以使 Service 层的代码更加清晰和易于维护。同时遵循 CQS 原则,也使系统行为更加可预测。[DDD 指南 - 工厂](ddd-guide.md#工厂factory)
```go
// 购买商品使用工厂的示例
func (s *CartService) Shop(ctx context.Context, cartID string) error {
    // 使用工厂创建购买命令
    cmd, err := s.CartFactory.CreateShopCmd(cartID)
    if err != nil {
        return err
    }

    // 执行购买操作
    return cmd.Run()
}

// 查询商品使用工厂的示例
func (s *CartService) GetCartDetail(ctx context.Context, cartID string) (*dto.CartDetail, error) {
    // 创建查询
    query := s.CartFactory.CreateDetailQuery(cartID)
    
    // 执行查询并返回结果
    return query.Result()
}
```

### 3.4 事务注意事项

1. 事务会自动处理回滚,在 Execute 闭包中返回 error 会触发回滚
2. 事务中的所有数据库操作都会使用同一个事务连接
3. 事务结束后会自动提交或回滚
4. 事务支持嵌套使用
5. 事务中可以结合领域事件使用

## 4. 最佳实践

1. Service 职责单一,只实现特定领域的业务逻辑
2. 通过依赖注入解耦各个组件
3. 合理使用事务保证数据一致性
4. 使用 Worker 处理日志、上下文等
5. 遵循 DDD 设计原则,合理划分领域边界
6. Service 依赖字段必须是公开的（首字母大写），以便框架注入




