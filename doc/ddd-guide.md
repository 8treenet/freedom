# DDD 实体、聚合、工厂设计指南

## 目录
- [实体（Entity）](#实体entity)
- [聚合（Aggregate）](#聚合aggregate)
- [工厂（Factory）](#工厂factory)

## 项目结构
在 DDD 项目中，实体、聚合和工厂通常按照以下目录结构组织：

```
domain/
├── entity/             # 实体定义
│   ├── cart.go        # 购物车实体
│   ├── goods.go       # 商品实体
│   ├── delivery.go    # 配送实体
│   └── admin.go       # 管理员实体
│
├── aggregate/          # 聚合定义
│   ├── cart_factory.go    # 购物车聚合工厂
│   └── cart_add_cmd.go    # 购物车命令
│
└── dependency/         # 领域依赖定义
    └── dependency.go   # 依赖接口
```

### 目录说明
- **entity/**: 存放所有领域实体，每个实体一个文件
- **aggregate/**: 包含聚合根和相关命令的定义
- **dependency/**: 定义领域层所需的依赖接口

### 命名规范
- 实体文件：`实体名.go`
- 聚合工厂：`聚合名_factory.go`
- 聚合命令：`聚合名_命令.go`

## 实体（Entity）

### 概念
实体是领域驱动设计中的核心概念，代表了业务领域中具有唯一标识和生命周期的对象。在实现上，实体通常继承自持久化对象（PO），但会添加业务行为和规则。

### 实现方式

1. **基础结构**
   - 所有实体必须继承 freedom.Entity
   - 同时继承对应的 PO 结构
   ```go
   // 标准实体结构示例
   type Goods struct {
       freedom.Entity    // 必须继承基础实体
       po.Goods         // 继承对应的PO结构
   }

   // 实现 Identity 方法
   func (g *Goods) Identity() string {
       return strconv.Itoa(g.ID)
   }
   ```

2. **实体标准接口**
   - 实体必须实现 Identity() 方法返回唯一标识
   - Entity 基类提供了领域事件等基础能力
   ```go
   type Cart struct {
       freedom.Entity
       po.Cart
       Items []CartItem
   }

   func (c *Cart) Identity() string {
       return c.ID
   }
   ```

3. **业务行为封装**
   - 实体应该封装其业务行为，而不是将业务逻辑暴露在服务层
   - 方法名应该体现业务含义
   ```go
   func (c *Cart) AddItem(item CartItem) error {
       // 业务规则验证
       if item.Quantity <= 0 {
           return errors.New("quantity must be positive")
       }
       
       // 业务逻辑实现
       for i, existingItem := range c.Items {
           if existingItem.ProductID == item.ProductID {
               c.Items[i].Quantity += item.Quantity
               return nil
           }
       }
       
       c.Items = append(c.Items, item)
       return nil
   }
   ```

4. **状态管理**
   - 实体应该管理自己的状态转换
   - 提供清晰的状态查询方法
   ```go
   func (c *Cart) IsEmpty() bool {
       return len(c.Items) == 0
   }
   
   func (c *Cart) TotalAmount() decimal.Decimal {
       total := decimal.Zero
       for _, item := range c.Items {
           total = total.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
       }
       return total
   }
   ```

### 最佳实践

1. **继承规范**
   - freedom.Entity 必须放在结构体的第一位
   - PO 继承放在第二位
   - 其他字段放在后面
   ```go
   type Admin struct {
       freedom.Entity     // 第一位：基础实体
       po.Admin          // 第二位：PO结构
       permissions []string  // 其他业务字段
   }
   ```

2. **验证规则**
   - 在实体方法中实现业务验证规则
   - 返回有意义的错误信息
   ```go
   func (a *Admin) GrantPermission(permission string) error {
       if permission == "" {
           return errors.New("permission cannot be empty")
       }
       
       for _, p := range a.permissions {
           if p == permission {
               return errors.New("permission already granted")
           }
       }
       
       a.permissions = append(a.permissions, permission)
       return nil
   }
   ```

3. **不可变性**
   - 关键属性设置后不应更改
   - 使用私有字段和公共方法控制访问
   ```go
   type Delivery struct {
       po.Delivery
       address    string  // 私有字段
       createdAt time.Time
   }
   
   func (d *Delivery) Address() string {
       return d.address
   }
   ```

4. **领域事件**
   - 实体的重要状态变化应该触发领域事件
   - 使用事件来通知其他部分系统
   ```go
   func (c *Cart) Checkout() error {
       if c.IsEmpty() {
           return errors.New("cannot checkout empty cart")
       }
       
       // 状态变更
       c.Status = "Checked_Out"
       
       // 触发领域事件
       c.AddDomainEvent(&CartCheckedOutEvent{
           CartID: c.ID,
           Time:   time.Now(),
       })
       
       return nil
   }
   ```

5. **与服务层的协作**
   - 实体提供业务操作接口
   - 服务层负责协调多个实体的操作
   ```go
   // 在服务层中使用实体
   func (s *CartService) AddItemToCart(cartID string, item CartItem) error {
       cart, err := s.repo.Find(cartID)
       if err != nil {
           return err
       }
       
       if err := cart.AddItem(item); err != nil {
           return err
       }
       
       return s.repo.Save(cart)
   }
   ```

### 注意事项

1. 实体必须继承 freedom.Entity 并实现 Identity() 方法
2. 实体应该专注于自身的业务逻辑，不要包含其他实体的业务规则
3. 避免在实体中直接依赖外部服务或基础设施
4. 实体的方法应该是幂等的，相同的输入应该产生相同的结果
5. 实体的验证规则应该在实体内部实现，而不是依赖外部验证


## 工厂（Factory）
### 概念
工厂在 DDD 中主要用于创建复杂的聚合根和实体。它封装了对象的创建逻辑，确保聚合在创建时满足所有业务规则和约束。在 Freedom 框架中，工厂通过依赖注入的方式自动管理依赖关系。

### 实现方式

1. **工厂注册**
   ```go
   func init() {
       freedom.Prepare(func(initiator freedom.Initiator) {
           // 绑定创建工厂函数到框架，框架会自动处理依赖注入
           initiator.BindFactory(func() *CartFactory {
               return &CartFactory{} // 创建CartFactory
           })
       })
   }
   ```

2. **工厂结构定义**
   ```go
   // CartFactory 购物车聚合根工厂
   type CartFactory struct {
       UserRepo  dependency.UserRepo   // 依赖倒置用户资源库
       CartRepo  dependency.CartRepo   // 依赖倒置购物车资源库
       GoodsRepo dependency.GoodsRepo  // 依赖倒置商品资源库
       OrderRepo dependency.OrderRepo  // 依赖倒置订单资源库
   }
   ```

### 最佳实践

1. **依赖注入**
   - 工厂的依赖通过框架自动注入
   - 所有依赖都应该定义为接口类型
   ```go
   type CartFactory struct {
       UserRepo  dependency.UserRepo    // 用户仓储接口
       CartRepo  dependency.CartRepo    // 购物车仓储接口
       EventBus  dependency.EventBus    // 事件总线接口
   }
   ```

2. **聚合创建**
   - 工厂负责协调多个仓储
   - 确保创建过程中的数据一致性
   ```go
   func (f *CartFactory) Create(userID string) (*CartAddCmd, error) {
       // 验证用户是否存在
       if _, err := f.UserRepo.Find(ctx, userID); err != nil {
           return nil, err
       }
       ...
       return
   }
   ```

### 注意事项

1. 工厂必须在 init() 函数中注册到框架
2. 工厂的依赖字段必须是公开的（首字母大写），以便框架注入
3. 工厂应该只负责聚合的创建，不应包含业务逻辑
4. 所有依赖都应该通过接口定义，而不是具体实现
5. 工厂方法应该返回完整且有效的聚合

### 与其他组件的关系

1. **与框架的关系**
   - 工厂通过 freedom.Prepare 注册到框架
   - 框架负责依赖的自动注入和管理

2. **与仓储的关系**
   - 工厂通过依赖注入获得仓储接口
   - 仓储负责具体的数据访问操作

3. **与命令的关系**
   - 命令定义在聚合目录下
   - 工厂根据命令创建或操作聚合

4. **与 Service 的关系**
   - Service 通过依赖注入获得工厂实例
   - Service 调用工厂方法创建聚合
   - Service 负责业务流程编排，工厂负责聚合创建
   ```go
   type CartService struct {
       CartFactory *CartFactory  // 注入购物车工厂
   }

   func (s *CartService) AddToCart(ctx context.Context, userID, goodsID string, quantity int) error {
       // 使用工厂创建命令
       cmd, err := s.CartFactory.Create(userID)
       if err != nil {
           return err
       }
       cmd.Run()
   }
   ```

通过这种方式实现的工厂，可以很好地配合框架的依赖注入机制，使代码更加清晰和易于维护。同时，通过接口依赖的方式，也保证了代码的可测试性和灵活性。




## 聚合（Aggregate）

### 概念
聚合是一组相关实体的集合，由聚合根统一管理。在 Freedom 框架中，聚合基于命令查询职责分离(CQS)原则设计，将命令(Command)和查询(Query)分开处理。

### 实现方式

1. **聚合根定义**
   - 聚合根必须继承 freedom.Entity
   - 包含其他相关实体
   ```go
    // CartAddCmd 添加购物车聚合根
    type CartAddCmd struct {
        entity.User           //继承User实体
        goods    entity.Goods //包含Goods实体
        cartRepo dependency.CartRepo
    }
   ```

2. **命令定义**
   - 命令文件命名格式：`聚合名_操作_cmd.go`
   - 命令结构包含操作所需参数
   ```go
   // cart_add_cmd.go
   type CartAddCmd struct {
   }

   func (c *CartAddCmd) Run() error {
       // 执行添加商品到购物车的业务逻辑
       return
   }
   ```

3. **查询定义**
   - 查询文件命名格式：`聚合名_操作_query.go`
   - 查询结构包含查询条件
   ```go
   // cart_item_query.go
   type CartItemQuery struct {
        entity.User           //继承User实体
        cartRepo dependency.CartRepo
   }

   func (q *CartItemQuery) Result() ([]CartItem, error) {
        // 执行查询的业务逻辑
        return
   }
   ```

### 注意事项

1. 聚合根必须继承 freedom.Entity
2. 命令和查询要严格分离
3. 命令执行要保证原子性
4. 聚合边界要明确定义
5. 避免跨聚合的直接引用

### CQS 原则说明

1. **命令(Command)**
   - 改变系统状态
   - 不返回数据
   - 通过 Run() 方法执行

2. **查询(Query)**
   - 不改变系统状态
   - 返回数据
   - 通过 Result() 方法获取结果

通过这种方式实现的聚合，可以使业务逻辑更加清晰，同时通过 CQS 原则的应用，使系统更容易维护和扩展。命令和查询的分离也使得系统的行为更加可预测。
