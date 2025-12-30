# DDD 设计指南

## 目录
- [实体（Entity）](#实体entity)
- [聚合（Aggregate）](#聚合aggregate)
- [工厂（Factory）](#工厂factory)

## 项目结构

```
domain/
├── entity/          # 实体定义
│   ├── cart.go     # 购物车实体
│   ├── goods.go    # 商品实体
│   └── user.go     # 用户实体
│
├── aggregate/       # 聚合定义（CQS 模式）
│   ├── cart_factory.go   # 购物车工厂
│   ├── cart_add_cmd.go   # 添加命令
│   └── cart_item_query.go # 查询
│
└── dependency/     # 领域依赖接口
    └── dependency.go
```

### 命名规范
- 实体文件：`实体名.go`
- 工厂文件：`聚合名_factory.go`
- 命令文件：`聚合名_操作_cmd.go`
- 查询文件：`聚合名_查询_query.go`

---

## 实体（Entity）

### 概念
实体是具有唯一标识和生命周期的领域对象，封装业务行为和规则。

### 基础结构

```go
type Cart struct {
    freedom.Entity    // 必须继承
    po.Cart          // 继承 PO 结构
}

func (c *Cart) Identity() string {
    return strconv.Itoa(c.ID)
}
```

### 核心要求

1. **必须继承 freedom.Entity**
2. **必须实现 Identity() 方法**，返回唯一标识
3. **封装业务行为**，避免贫血模型

### 业务行为示例

```go
// 添加商品到购物车
func (c *Cart) AddItem(item CartItem) error {
    if item.Quantity <= 0 {
        return errors.New("quantity must be positive")
    }
    // 业务逻辑...
    return nil
}

// 计算总价
func (c *Cart) TotalAmount() decimal.Decimal {
    total := decimal.Zero
    for _, item := range c.Items {
        total = total.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
    }
    return total
}
```

### 领域事件

```go
func (u *User) ChangePassword(newPassword, oldPassword string) error {
    if u.Password != oldPassword {
        return errors.New("Password error")
    }
    u.SetPassword(newPassword)

    // 发布领域事件
    u.AddPubEvent(&event.ChangePassword{
        UserID:      u.ID,
        NewPassword: u.Password,
        OldPassword: oldPassword,
    })
    return nil
}
```

### 最佳实践

| 原则 | 说明 |
|------|------|
| 继承顺序 | freedom.Entity 在第一位，PO 在第二位 |
| 业务封装 | 业务逻辑在实体内部，不在服务层 |
| 验证规则 | 在实体方法中实现，返回有意义错误 |
| 状态管理 | 实体管理自身状态转换 |
| 避免依赖 | 不直接依赖外部服务或基础设施 |

---

## 聚合（Aggregate）

### 概念
聚合基于 CQS（命令查询分离）原则设计，将写操作（Command）和读操作（Query）分离。聚合根**组合**实体而非继承 freedom.Entity。

### 命令（Command）

```go
// CartAddCmd 添加购物车命令
type CartAddCmd struct {
    entity.User          // 继承 User 实体
    goods    entity.Goods // 包含 Goods 实体
    cartRepo dependency.CartRepo
}

// Run 执行命令
func (cmd *CartAddCmd) Run(goodsNum int) error {
    // 库存验证
    if goodsNum > cmd.goods.Stock {
        return errors.New("inventory not enough")
    }

    // 查找购物车中是否已有该商品
    if cartEntity, err := cmd.cartRepo.FindByGoodsID(cmd.User.ID, cmd.goods.ID); err == nil {
        cartEntity.AddNum(goodsNum)
        return cmd.cartRepo.Save(cartEntity)
    }

    // 创建新购物车项
    _, e := cmd.cartRepo.New(cmd.User.ID, cmd.goods.ID, goodsNum)
    return e
}
```

### 查询（Query）

```go
// CartItemQuery 购物车查询
type CartItemQuery struct {
    entity.User
    allCart  []*entity.Cart
    goodsMap map[int]*entity.Goods
}

// VisitAllItem 遍历所有商品
func (query *CartItemQuery) VisitAllItem(f func(id, goodsId int, goodsName string, goodsNum, totalPrice int)) {
    for i := 0; i < len(query.allCart); i++ {
        goodsEntity := query.goodsMap[query.allCart[i].GoodsID]
        f(query.allCart[i].ID, goodsEntity.ID, goodsEntity.Name, 
          query.allCart[i].Num, query.allCart[i].Num*goodsEntity.Price)
    }
}

// AllItemTotalPrice 计算总价
func (query *CartItemQuery) AllItemTotalPrice() (totalPrice int) {
    for i := 0; i < len(query.allCart); i++ {
        goodsEntity := query.goodsMap[query.allCart[i].GoodsID]
        totalPrice += query.allCart[i].Num * goodsEntity.Price
    }
    return
}
```

### CQS 原则

| 类型 | Command（命令） | Query（查询） |
|------|----------------|---------------|
| 目的 | 改变系统状态 | 读取数据 |
| 返回值 | 通常返回 error | 返回数据 |
| 执行方法 | `Run()` | 自定义方法（如 `VisitAllItem`） |
| 副作用 | 有副作用 | 无副作用 |

### 注意事项

1. 聚合根继承实体，而非 freedom.Entity
2. 命令和查询严格分离
3. 命令执行保证原子性
4. 避免跨聚合直接引用
5. 明确定义聚合边界

---

## 工厂（Factory）

### 概念
工厂负责创建聚合根（Command/Query），封装创建逻辑并确保业务规则约束。通过框架自动注入依赖。

### 工厂注册

```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        initiator.BindFactory(func() *CartFactory {
            return &CartFactory{}
        })
    })
}
```

### 工厂结构

```go
type CartFactory struct {
    UserRepo  dependency.UserRepo  // 用户仓储接口
    CartRepo  dependency.CartRepo  // 购物车仓储接口
    GoodsRepo dependency.GoodsRepo // 商品仓储接口
    OrderRepo dependency.OrderRepo // 订单仓储接口
}
```

### 创建命令

```go
func (f *CartFactory) NewCartAddCmd(goodsID, userID int) (*CartAddCmd, error) {
    // 验证用户
    user, e := f.UserRepo.Get(userID)
    if e != nil {
        user.Worker().Logger().Error(e, "userId", userID)
        return nil, e
    }

    // 验证商品
    goods, e := f.GoodsRepo.Get(goodsID)
    if e != nil {
        goods.Worker().Logger().Error(e, "goodsId", goodsID)
        return nil, e
    }

    // 创建命令
    cmd := &CartAddCmd{
        cartRepo: f.CartRepo,
    }
    cmd.User = *user
    cmd.goods = *goods
    return cmd, nil
}
```

### 创建查询

```go
func (f *CartFactory) NewCartItemQuery(userID int) (*CartItemQuery, error) {
    user, e := f.UserRepo.Get(userID)
    if e != nil {
        user.Worker().Logger().Error(e, "userId", userID)
        return nil, e
    }

    query := &CartItemQuery{}
    query.User = *user
    query.goodsMap = make(map[int]*entity.Goods)

    // 加载购物车数据
    query.allCart, e = f.CartRepo.FindAll(query.User.ID)
    for i := 0; i < len(query.allCart); i++ {
        goodsEntity, e := f.GoodsRepo.Get(query.allCart[i].GoodsID)
        if e != nil {
            return nil, e
        }
        query.goodsMap[goodsEntity.ID] = goodsEntity
    }
    return query, nil
}
```

### 在服务层使用

```go
type CartService struct {
    CartFactory *CartFactory // 注入工厂
}

func (s *CartService) AddToCart(ctx context.Context, userID, goodsID string, quantity int) error {
    // 使用工厂创建命令
    cmd, err := s.CartFactory.NewCartAddCmd(goodsID, userID)
    if err != nil {
        return err
    }
    return cmd.Run(quantity)
}
```

### 注意事项

| 要求 | 说明 |
|------|------|
| 注册位置 | 必须在 `init()` 函数中注册 |
| 依赖字段 | 必须公开（首字母大写） |
| 依赖类型 | 必须是接口类型 |
| 职责范围 | 只负责创建，不包含业务逻辑 |
| 返回值 | 返回完整有效的聚合 |

### 组件关系

```
Service
   ↓ 依赖注入
Factory
   ↓ 依赖注入
Repository
   ↓ 数据访问
Entity (继承 freedom.Entity)
```
