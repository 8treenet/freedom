# FShop - DDD 架构教学示例

## 项目简介

FShop 是一个完整的电商系统示例，展示如何使用 Freedom Framework 构建基于 DDD（领域驱动设计）架构的企业级应用。本项目通过电商场景（商品、购物车、订单、发货）演示 DDD 的核心概念和最佳实践。

### 学习目标

通过本示例，你将学习到：
- DDD 分层架构的实际应用
- 聚合、实体、工厂的设计与实现
- CQS（命令查询分离）模式
- 领域事件驱动架构
- 资源库模式与依赖倒置
- 事务处理与数据一致性

---

## 项目结构

```
fshop/
├── adapter/              # 适配器层（外层）
│   ├── controller/      # HTTP 控制器
│   ├── repository/      # 资源库实现
│   └── consumer/       # 事件消费者
├── domain/              # 领域层（核心层）
│   ├── aggregate/       # 聚合（CQS 模式）
│   ├── entity/          # 实体
│   ├── po/              # 持久化对象
│   ├── vo/              # 值对象
│   ├── event/           # 领域事件
│   └── dependency/      # 依赖接口
└── infra/               # 基础设施层（内层）
    └── domainevent/     # 领域事件基础设施
```

### 架构分层说明

| 层级 | 职责 | 说明 |
|------|------|------|
| Adapter | 适配外部 | 处理 HTTP 请求、数据库访问、消息队列 |
| Domain | 核心业务 | 封装业务逻辑，不依赖外部实现 |
| Infra | 基础设施 | 提供通用能力（事件、缓存、事务） |

---

## 步骤一：数据库与 PO 生成

### 1.1 创建数据库表

```sql
-- 商品表
CREATE TABLE goods (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100),
    price INT,
    stock INT,
    tag VARCHAR(20),
    created_at DATETIME,
    updated_at DATETIME
);

-- 购物车表
CREATE TABLE cart (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT,
    goods_id INT,
    num INT,
    created_at DATETIME
);

-- 订单表
CREATE TABLE orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_no VARCHAR(50),
    user_id INT,
    total_price INT,
    status VARCHAR(20),
    created_at DATETIME
);
```

### 1.2 生成 PO 代码

使用 Freedom 提供的 `new-po` 工具自动生成持久化对象：

```bash
freedom new-po --dsn "root:password@tcp(127.0.0.1:3306)/fshop?charset=utf8"
```

生成的文件位于 `domain/po/` 目录：

```go
// domain/po/goods.go
type Goods struct {
    ID        int       `gorm:"primaryKey;column:id"`
    Name      string    `gorm:"column:name"`
    Price     int       `gorm:"column:price"`
    Stock     int       `gorm:"column:stock"`
    Tag       string    `gorm:"column:tag"`
    CreatedAt time.Time `gorm:"column:created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at"`
    changes   map[string]interface{}
}

// 自动生成的字段更新方法
func (obj *Goods) SetName(name string) {
    obj.Name = name
    obj.Update("name", name)
}

func (obj *Goods) AddStock(stock int) {
    obj.Stock += stock
    obj.Update("stock", gorm.Expr("stock + ?", stock))
}
```

> **详细说明**：参见 [PO 指南](../../doc/po-guide.md)

---

## 步骤二：创建实体

实体继承 `freedom.Entity` 和 PO 结构，封装业务行为。

```go
// domain/entity/goods.go
package entity

import (
    "strconv"
    "github.com/8treenet/freedom"
    "github.com/8treenet/freedom/example/fshop/domain/po"
)

// Goods 商品实体
type Goods struct {
    freedom.Entity    // 必须继承
    po.Goods         // 继承 PO 结构
}

// Identity 返回唯一标识
func (g *Goods) Identity() string {
    return strconv.Itoa(g.ID)
}

// MarkedTag 为商品打标签
func (g *Goods) MarkedTag(tag string) error {
    if tag != "HOT" && tag != "NEW" && tag != "NONE" {
        return errors.New("Tag doesn't exist")
    }
    g.SetTag(tag)
    return nil
}
```

```go
// domain/entity/cart.go
package entity

type Cart struct {
    freedom.Entity
    po.Cart
}

// Identity 返回唯一标识
func (c *Cart) Identity() string {
    return strconv.Itoa(c.ID)
}

// AddNum 增加商品数量
func (c *Cart) AddNum(num int) {
    c.Num += num
}
```

> **详细说明**：参见 [DDD 指南 - 实体](../../doc/ddd-guide.md#实体entity)

---

## 步骤三：创建聚合（CQS 模式）

聚合基于 CQS（命令查询分离）原则，将写操作（Command）和读操作（Query）分离。

### 3.1 定义依赖接口

```go
// domain/dependency/dependency.go
package dependency

type CartRepo interface {
    New(userID, goodsID, num int) (*entity.Cart, error)
    FindByGoodsID(userID, goodsID int) (*entity.Cart, error)
    Save(cart *entity.Cart) error
    FindAll(userID int) ([]*entity.Cart, error)
    DeleteAll(userID int) error
}

type GoodsRepo interface {
    Get(id int) (*entity.Goods, error)
    Save(goods *entity.Goods) error
}
```

### 3.2 创建工厂

工厂负责创建聚合根（Command/Query），框架自动注入依赖。

```go
// domain/aggregate/cart_factory.go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        initiator.BindFactory(func() *CartFactory {
            return &CartFactory{}
        })
    })
}

// CartFactory 购物车聚合工厂
type CartFactory struct {
    UserRepo  dependency.UserRepo
    CartRepo  dependency.CartRepo
    GoodsRepo dependency.GoodsRepo
}

// NewCartAddCmd 创建添加购物车命令
func (f *CartFactory) NewCartAddCmd(goodsID, userID int) (*CartAddCmd, error) {
    // 验证用户
    user, e := f.UserRepo.Get(userID)
    if e != nil {
        return nil, e
    }

    // 验证商品
    goods, e := f.GoodsRepo.Get(goodsID)
    if e != nil {
        return nil, e
    }

    cmd := &CartAddCmd{
        cartRepo: f.CartRepo,
    }
    cmd.User = *user
    cmd.goods = *goods
    return cmd, nil
}
```

### 3.3 创建命令（Command）

命令负责修改系统状态。

```go
// domain/aggregate/cart_add_cmd.go
// CartAddCmd 添加购物车聚合根
type CartAddCmd struct {
    entity.User          // 继承 User 实体
    goods    entity.Goods // 包含 Goods 实体
    cartRepo dependency.CartRepo
}

// Run 执行命令
func (cmd *CartAddCmd) Run(goodsNum int) error {
    // 库存验证
    if goodsNum > cmd.goods.Stock {
        return errors.New("库存不足")
    }

    // 查找购物车中是否已有该商品
    if cartEntity, err := cmd.cartRepo.FindByGoodsID(cmd.User.ID, cmd.goods.ID); err == nil {
        // 已存在，增加数量
        cartEntity.AddNum(goodsNum)
        return cmd.cartRepo.Save(cartEntity)
    }

    // 创建新购物车项
    _, e := cmd.cartRepo.New(cmd.User.ID, cmd.goods.ID, goodsNum)
    return e
}
```

### 3.4 创建查询（Query）

查询负责读取数据，不修改状态。

```go
// domain/aggregate/cart_item_query.go
// CartItemQuery 购物车查询聚合根
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

// MarshalJSON JSON 序列化
func (query *CartItemQuery) MarshalJSON() ([]byte, error) {
    var obj struct {
        TotalPrice int `json:"totalPrice"`
        Items      []struct {
            ID         int    `json:"id"`
            GoodsID    int    `json:"goodsId"`
            GoodsName  string `json:"goodsName"`
            GoodsNum   int    `json:"goodsNum"`
            TotalPrice int    `json:"totalPrice"`
        }
    }
    obj.TotalPrice = query.AllItemTotalPrice()
    for i := 0; i < len(query.allCart); i++ {
        goodsEntity := query.goodsMap[query.allCart[i].GoodsID]
        obj.Items = append(obj.Items, struct {
            ID         int
            GoodsID    int
            GoodsName  string
            GoodsNum   int
            TotalPrice int
        }{
            query.allCart[i].ID,
            goodsEntity.ID,
            goodsEntity.Name,
            query.allCart[i].Num,
            query.allCart[i].Num * goodsEntity.Price,
        })
    }
    return json.Marshal(obj)
}
```

> **详细说明**：参见 [DDD 指南 - 聚合](../../doc/ddd-guide.md#聚合aggregate)

---

## 步骤四：创建服务层

服务层协调领域对象，实现业务流程编排。

```go
// domain/cart.go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        initiator.BindService(func() *CartService {
            return &CartService{}
        })
        initiator.InjectController(func(ctx freedom.Context) (service *CartService) {
            initiator.FetchService(ctx, &service)
            return
        })
    })
}

// CartService 领域服务
type CartService struct {
    Worker      freedom.Worker
    CartRepo    dependency.CartRepo
    CartFactory *aggregate.CartFactory
}

// Add 添加商品到购物车
func (c *CartService) Add(userID, goodsID, goodsNum int) error {
    // 使用工厂创建命令
    cmd, e := c.CartFactory.NewCartAddCmd(goodsID, userID)
    if e != nil {
        return e
    }
    return cmd.Run(goodsNum)
}

// Items 查询购物车商品
func (c *CartService) Items(userID int) (json.Marshaler, error) {
    // 使用工厂创建查询
    query, e := c.CartFactory.NewCartItemQuery(userID)
    if e != nil {
        return nil, e
    }
    return query, nil
}
```

> **详细说明**：参见 [Service 指南](../../doc/service-guide.md)

---

## 步骤五：创建控制器

控制器处理 HTTP 请求，调用服务层。

```go
// adapter/controller/cart.go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        initiator.BindController("/cart", &CartController{})
    })
}

// CartController 购物车控制器
type CartController struct {
    Worker    freedom.Worker
    CartSev   *domain.CartService
    Request   *infra.Request
}

// Post 添加商品到购物车
func (c *CartController) Post() freedom.Result {
    var req vo.CartAddReq
    e := c.Request.ReadJSON(&req, true)
    if e != nil {
        return &infra.JSONResponse{Error: e}
    }

    e = c.CartSev.Add(req.UserID, req.GoodsID, req.GoodsNum)
    return &infra.JSONResponse{Error: e}
}

// GetItems 获取购物车商品列表
func (c *CartController) GetItems() freedom.Result {
    userID, err := c.Worker.IrisContext().URLParamInt("userId")
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }

    vo, err := c.CartSev.Items(userID)
    if err != nil {
        return &infra.JSONResponse{Error: err}
    }
    return &infra.JSONResponse{Object: vo}
}
```

> **详细说明**：参见 [路由指南](../../doc/route-guide.md)

---

## 步骤六：领域事件

### 6.1 定义事件

```go
// domain/event/shop_goods.go
type ShopGoods struct {
    UserID    int
    OrderNO   string
    GoodsID   int
    GoodsNum  int
    GoodsName string
}
```

### 6.2 实体发布事件

```go
// domain/entity/user.go
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

### 6.3 事件基础设施

```go
// infra/domainevent/event_transaction.go
type EventTransaction struct {
    txManager *transaction.Transaction
    eventBus  *eventManager
}

// Execute 执行事务并处理事件
func (e *EventTransaction) Execute(fn func() error) error {
    // 1. 执行事务
    err := e.txManager.Execute(fn)
    if err != nil {
        return err
    }

    // 2. 事务成功后推送事件
    return e.eventBus.push()
}
```

---

## 步骤七：事务处理

事务组件保证多个数据库操作的原子性。

```go
// domain/aggregate/cart_shop_cmd.go
type CartShopCmd struct {
    orderRepo   dependency.OrderRepo
    goodsRepo   dependency.GoodsRepo
    cartRepo    dependency.CartRepo
    tx          *domainevent.EventTransaction
}

// Shop 购物车结算
func (cmd *CartShopCmd) Shop() error {
    return cmd.tx.Execute(func() error {
        // 1. 修改商品库存
        for _, cart := range cmd.allCartEntity {
            goods := cmd.goodsEntityMap[cart.GoodsID]
            goods.AddStock(-cart.Num)
            if err := cmd.goodsRepo.Save(goods); err != nil {
                return err
            }
        }

        // 2. 创建订单
        order, err := cmd.orderRepo.New()
        if err != nil {
            return err
        }

        // 3. 清空购物车
        if err := cmd.cartRepo.DeleteAll(cmd.userEntity.ID); err != nil {
            return err
        }

        return nil
    })
}
```

---

## 步骤八：资源库实现

资源库实现依赖接口，负责数据访问。

```go
// adapter/repository/cart.go
type CartRepository struct {
    freedom.Repository
}

func (repo *CartRepository) New(userID, goodsID, num int) (*entity.Cart, error) {
    cart := &entity.Cart{}
    cart.UserID = userID
    cart.GoodsID = goodsID
    cart.Num = num
    _, err := createCart(repo, cart.PO())
    return cart, err
}

func (repo *CartRepository) FindByGoodsID(userID, goodsID int) (*entity.Cart, error) {
    cart, err := findCartByMap(repo, map[string]interface{}{
        "user_id":  userID,
        "goods_id": goodsID,
    })
    if err != nil {
        return nil, err
    }
    return &entity.Cart{Cart: cart}, nil
}
```

---

## 学习路径建议

1. **基础理解**：阅读 [PO 指南](../../doc/po-guide.md)，了解数据持久化
2. **领域建模**：阅读 [DDD 指南](../../doc/ddd-guide.md)，理解实体、聚合、工厂
3. **服务层**：阅读 [Service 指南](../../doc/service-guide.md)，学习业务编排
4. **路由开发**：阅读 [路由指南](../../doc/route-guide.md)，掌握 HTTP 接口
5. **实践项目**：参考本项目代码，动手实现一个简单的 DDD 应用

---

## 核心概念总结

| 概念 | 说明 | 示例 |
|------|------|------|
| 实体（Entity） | 有唯一标识的对象，封装业务行为 | User、Goods、Cart |
| 聚合（Aggregate） | 一组相关实体，由聚合根管理 | CartAddCmd、CartItemQuery |
| 工厂（Factory） | 创建聚合根，封装创建逻辑 | CartFactory、ShopFactory |
| 服务（Service） | 协调领域对象，实现业务流程 | CartService、OrderService |
| 资源库（Repository） | 数据访问抽象，依赖倒置 | CartRepo、GoodsRepo |
| 领域事件（Event） | 重要状态变化的通知 | ShopGoods、OrderPay |
