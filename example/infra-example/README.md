# Freedom 基础设施示例

## 目录
- [概述](#概述)
- [自定义基础设施组件](#自定义基础设施组件)
   - [组件类型与生命周期](#组件类型与生命周期)
   - [组件注册](#组件注册)
   - [单例组件示例](#单例组件示例)
   - [多例组件示例](#多例组件示例)
   - [最佳实践](#最佳实践)
- [Kafka消息](#kafka消息)
   - [Kafka消息概述](#kafka消息概述)
   - [消息发送](#消息发送)
   - [消息消费](#消息消费)
   - [与领域事件的区别](#与领域事件的区别)
- [事务组件](#事务组件)
   - [核心概念](#核心概念)
   - [使用示例](#使用示例)
- [领域事件](#领域事件)
	- [领域事件概述](#领域事件概述)
	- [事件实现](#事件实现)
	- [事件持久化](#事件持久化)
	- [订阅流程](#订阅流程)
	- [EventManager组件](#EventManager组件)
	- [最佳实践](#最佳实践)
	- [调试与监控](#调试与监控)


## 概述

本指南演示了Freedom框架的三个核心基础设施：
1. 自定义组件：扩展框架功能
2. 事务管理：确保数据一致性
3. 领域事件：基于Kafka的事件驱动架构


## 自定义基础设施组件

Freedom 框架支持自定义基础设施组件（Infrastructure Components），用于扩展框架功能和复用基础设施代码。基础设施组件是一种特殊的依赖注入对象，可以被注入到控制器、服务或其他组件中。

### 组件类型与生命周期

Freedom 框架支持两种类型的基础设施组件：

1. **单例组件（Singleton）**：
   - 整个应用生命周期内只有一个实例
   - 通过 `Booting` 方法在应用启动时初始化
   - 适用于需要全局共享的资源，如配置管理、连接池、缓存等
   - 所有请求共享同一个实例，需要注意并发安全

2. **多例组件（Prototype）**：
   - 每个请求都会创建新的实例
   - 通过 `BeginRequest` 方法在请求开始时初始化
   - 适用于请求级别的处理，如请求解析、参数验证等
   - 实例之间相互独立，不需要考虑并发问题

### 组件注册

组件注册通常在 `init()` 函数中完成：

```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 1. 注册单例组件
        initiator.BindInfra(true, &Single{})
        
        // 2. 注册多例组件
        initiator.BindInfra(false, func() *RequestValidator {
            return &RequestValidator{}
        })
        
        // 3. 注入到控制器（可选），Service、Factory、Repository无需注册，可以直接注入使用
        initiator.InjectController(func(ctx freedom.Context) (com *Single) {
            initiator.FetchInfra(ctx, &com)
            return
        })
    })
}
```

### 单例组件示例

以下是一个完整的单例组件示例，实现了配置管理功能：

```go
// example/infra-example/infra/demo/single.go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        // 绑定为单例组件，全局只创建1个对象
        initiator.BindInfra(true, &Single{})
        
        // 注入到控制器
        initiator.InjectController(func(ctx freedom.Context) (com *Single) {
            initiator.FetchInfra(ctx, &com)
            return
        })
    })
}

// Single 单例组件示例
type Single struct {
    life int    // 生命值
}

// Booting 启动时初始化
func (s *Single) Booting(boot freedom.BootManager) {
    freedom.Logger().Info("Single.Booting")
    s.life = rand.Intn(100)    // 初始化生命值
}

// GetLife 获取生命值
func (s *Single) GetLife() int {
    return s.life    // 所有请求获取的都是相同的值
}

// GoodsController .
type GoodsController struct {
	Single       *infra.Single
}

// GetBy handles the GET: /goods/:id route 获取指定商品.
func (goods *GoodsController) GetBy(goodsID int) freedom.Result {
	goods.Single.GetLife() //返回100
}
```

### 多例组件示例

以下是一个处理请求验证的多例组件示例：

```go
// example/infra-example/infra/request.go
func init() {
	validate = validator.New()
	freedom.Prepare(func(initiator freedom.Initiator) {
		// 绑定为多例组件，每个请求会创建一个对象
		initiator.BindInfra(false, func() *Request {
			return &Request{}
		})
		// 注入到控制器
		initiator.InjectController(func(ctx freedom.Context) (com *Request) {
			initiator.FetchInfra(ctx, &com)
			return
		})
	})
}

// Request .
type Request struct {
	freedom.Infra
}

// BeginRequest .
func (req *Request) BeginRequest(worker freedom.Worker) {
	req.Infra.BeginRequest(worker)
}

// ReadJSON .
func (req *Request) ReadJSON(obj interface{}, validates ...bool) error {
	rawData, err := ioutil.ReadAll(req.Worker().IrisContext().Request().Body)
	//每个请求的数据都不同
}
```

### 最佳实践

1. **组件设计原则**：
   - 单一职责：每个组件专注于解决特定的基础设施问题
   - 接口隔离：提供清晰的公共接口，隐藏实现细节
   - 可配置性：支持通过配置文件或启动参数调整行为

2. **并发处理**：
   - 单例组件需要保证并发安全
   - 使用适当的锁机制保护共享资源
   - 多例组件不需要考虑并发问题

3. **错误处理**：
   - 在初始化阶段失败时及时报错
   - 提供详细的错误信息和日志
   - 实现优雅降级和故障恢复机制


## Kafka消息

### Kafka消息概述

Kafka消息是Freedom框架提供的普通消息队列功能，用于实现异步处理、系统解耦和事件驱动架构。与领域事件不同，普通Kafka消息不依赖事务组件，可以独立发送和消费。

### 消息发送

Kafka消息通过依赖注入的方式使用，只需定义成员变量即可开箱即用。Producer是单例组件，全局共享。

以下是一个完整的消息发送示例：

```go
// OrderRepository 订单资源库
type OrderRepository struct {
	freedom.Repository
	Producer kafka.Producer // Kafka生产者组件（单例）
}

// Pay 订单支付并发送消息
func (repo *OrderRepository) Pay(orderID, userID int) error {
	// 1. 更新订单状态
	orderPO, err := OrderToPoint(findOrderByWhere(repo, "id = ? and user_id = ?", []interface{}{orderID, userID}))
	if err != nil {
		return err
	}
	if orderPO.Status == "PAID" {
		return errors.New("订单已支付")
	}
	orderPO.SetStatus("PAID")
	if _, err := saveOrder(repo, orderPO); err != nil {
		return err
	}

	// 2. 发送支付消息到 Kafka
	payEvent := event.OrderPay{
		OrderID: orderID,
		UserID:  userID,
	}
	data, _ := json.Marshal(payEvent)
	
	// 获取链路追踪ID并设置到消息头
	traceHead := repo.Worker().Bus().Header.Clone()
	return repo.Producer.NewMsg("OrderPay", data).
		SetHeader(map[string]interface{}{"x-request-id": traceHead.Get("x-request-id")}).
		Publish()
}
```

### 消息消费

Kafka消息消费通过Controller实现，使用`ListenEvent`注册Topic与处理方法的映射关系。

```go
// 注册事件监听
func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/order", &OrderController{})
		// 注册OrderPay topic的消费处理
		initiator.ListenEvent("OrderPay", "OrderController.OrderPayEvent", false)
	})
}

// OrderController 订单控制器
type OrderController struct {
	Worker   freedom.Worker
	Request  *infra.Request
}

// BeforeActivation 自定义路由
func (controller *OrderController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("POST", "/orderPayEvent", "OrderPayEvent")
}

// OrderPayEvent 处理订单支付消息
func (c *OrderController) OrderPayEvent() freedom.Result {
	var payEvent event.OrderPay
	if e := c.Request.ReadJSON(&payEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	// 处理支付消息，这里可以添加其他业务逻辑
	c.Worker.Logger().Infof("收到订单支付消息: OrderID=%d, UserID=%d", payEvent.OrderID, payEvent.UserID)
	
	// 业务处理逻辑...
	
	return &infra.JSONResponse{}
}
```

### 与领域事件的区别

| 特性 | Kafka消息 | 领域事件 |
|------|-----------|-----------|
| **事务保证** | 无事务保证，独立发送 | 依赖事务组件，保证原子性 |
| **消息可靠性** | 发送即忘，无重试机制 | 本地消息表+重试机制，保证可靠投递 |
| **消息持久化** | 不持久化 | 持久化到本地消息表 |
| **适用场景** | 非关键业务、可容忍消息丢失 | 关键业务、必须保证消息可靠 |


## 事务组件

### 核心概念

- 事务原子性：事务内的所有操作要么全部成功，要么全部失败
- 自动回滚：任何操作失败都会触发回滚
- 事件集成：成功执行后自动触发领域事件发布

### 使用示例

事务组件通过依赖注入的方式使用，只需定义成员变量即可开箱即用。事务组件确保了多个操作的原子性，要么全部成功，要么全部回滚。

以下是一个完整的示例：

```go
// OrderService 订单服务
type OrderService struct {
	Worker           freedom.Worker                   // 请求运行时对象
	GoodsRepo        *repository.GoodsRepository     // 商品资源库
	OrderRepo        *repository.OrderRepository     // 订单资源库
    Transaction 	 transaction.Transaction 		// 依赖注入事务组件
}

// Shop 购物服务，展示了事务和领域事件的结合使用
func (srv *OrderService) Shop(goodsID, num, userID int) (e error) {
	// 1. 获取商品实体
	goodsEntity, e := srv.GoodsRepo.Get(goodsID)
	if e != nil {
		return
	}

	// 2. 验证库存
	if goodsEntity.Stock < num {
		return errors.New("库存不足")
	}

	// 3. 更新库存
	goodsEntity.AddStock(-num)

	// 4. 执行事务
	return srv.Transaction.Execute(func() error {
		// 保存商品信息
		if err := srv.GoodsRepo.Save(goodsEntity); err != nil {
			return err
		}
		// 创建订单
		return srv.OrderRepo.Create(goodsEntity.ID, num, userID)
	})
}
```

## 领域事件与KafkaMQ

### 领域事件概述

领域事件（Domain Event）是对系统中已发生的、对其他部分有意义的事实的记录。它代表了一个领域中发生的事情，这个事情已经发生了，并且其他部分可能会对此感兴趣。

#### 为什么使用领域事件？

1. **业务解耦**：
   - 避免不同领域之间的直接调用
   - 减少系统间的直接依赖
   - 使系统更容易扩展和维护

2. **数据一致性**：
   - 最终一致性的实现方式
   - 分布式事务的异步解决方案
   - 跨服务数据同步

3. **业务可追溯**：
   - 记录重要的业务变更
   - 支持事件回溯
   - 便于问题排查

### 事件实现

Freedom框架使用事务性消息表配合Kafka来实现可靠的事件处理：

```go
// domain/event/shop_goods.go
// ShopGoods 购买事件
type ShopGoods struct {
	ID         string `json:"identity"`
	UserID     int    `json:"userID"`
	GoodsID    int    `json:"goodsID"`
	GoodsNum   int    `json:"goodsNum"`
	GoodsName  string `json:"goodsName"`
}
// Topic .
func (shop *ShopGoods) Topic() string {
	return "ShopGoods"
}


// OrderService 订单服务
type OrderService struct {
	Worker           freedom.Worker                   // 请求运行时对象
	GoodsRepo        *repository.GoodsRepository     // 商品资源库
	OrderRepo        *repository.OrderRepository     // 订单资源库
	EventTransaction *domainevent.EventTransaction   // 事务组件
}

// Shop 购物服务，展示了事务和领域事件的结合使用
func (srv *OrderService) Shop(goodsID, num, userID int) (e error) {
	// 1. 获取商品实体
	goodsEntity, e := srv.GoodsRepo.Get(goodsID)
	if e != nil {
		return
	}

	// 2. 验证库存
	if goodsEntity.Stock < num {
		return errors.New("库存不足")
	}

	// 3. 更新库存
	goodsEntity.AddStock(-num)

	// 4. 添加购物事件（重要：在业务逻辑中添加事件）
	goodsEntity.AddPubEvent(&event.ShopGoods{
		UserID:    userID,
		GoodsID:   goodsID,
		GoodsNum:  num,
		GoodsName: goodsEntity.Name,
	})

	// 5. 执行事务（事务会自动处理事件的发布）
	return srv.EventTransaction.Execute(func() error {
		if err := srv.GoodsRepo.Save(goodsEntity); err != nil {
			return err
		}
		return srv.OrderRepo.Create(goodsEntity.ID, num, userID)
	})
}
```

### 事件持久化

Freedom使用事务性消息表来确保事件的可靠性：

```go

// GoodsRepository 商品资源库
type GoodsRepository struct {
	freedom.Repository
	EventManager *domainevent.EventManager //领域事件组件
}

// Save 保存商品并处理事件
func (repo *GoodsRepository) Save(goods *entity.Goods) (e error) {
	// 这个方法内是由事务组件调用的，所以它们会一起Commit/Rollback
	_, e = saveGoods(repo, &goods.Goods)
	if e != nil {
		return
	}
	// 持久化实体身上的Pub/Sub事件
	return repo.EventManager.Save(&repo.Repository, goods)
}
```

### 订阅流程
```go
// 注册订阅者
func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/goods", &GoodsController{})
		initiator.ListenEvent((&event.ShopGoods{}).Topic(), "GoodsController.PostShop")
	})
}

// GoodsController 订单创建事件订阅者
type GoodsController struct {
	Worker       freedom.Worker
	GoodsSev     *domain.GoodsService
	Request      *infra.Request
	EventManager *domainevent.EventManager //领域事件组件
}

// PostShop handles the POST: /goods/shop route 商品购买事件.
func (goods *GoodsController) PostShop() freedom.Result {
	shopEvent := &event.ShopGoods{}
	//使用自定义的json组件
	if e := goods.Request.ReadJSON(shopEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}
	goods.Worker.Logger().Infof("领域事件消费 Topic:%s, %+v", shopEvent.Topic(), shopEvent)

	//先插入到事件表
	if e := goods.EventManager.InsertSubEvent(shopEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	if e := goods.GoodsSev.ShopEvent(shopEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}
	return &infra.JSONResponse{}
}

```

### EventManager组件

EventManager 是个自定义组件，它通过事务性消息表（也称为消息队列表）的设计模式，解决了本地事务与消息队列之间的数据一致性问题。

#### 1. 组件结构
```go
// EventManager 领域事件管理器
type EventManager struct {
    freedom.Infra
    eventRetry    *eventRetry         // 重试处理器
    kafkaProducer *kafka.ProducerImpl // Kafka生产者组件
}

// 发布事件表
type pubEventObject struct {
    Identity string    `gorm:"primary_key"` // 事件ID
    Topic    string    // 事件主题
    Content  string    // 事件内容
    Created  time.Time // 创建时间
    Updated  time.Time // 更新时间
}

// 订阅事件表
type subEventObject struct {
    Identity string    `gorm:"primary_key"` // 事件ID
    Topic    string    // 事件主题
    Content  string    // 事件内容
    Created  time.Time // 创建时间
    Updated  time.Time // 更新时间
}
```

#### 2. 工作原理

EventManager 采用了本地消息表的方案来保证事务和消息的一致性，其工作流程如下：

1. **事务阶段**：
```go
// Save 在事务中保存事件
func (manager *EventManager) Save(repo *freedom.Repository, entity freedom.Entity) (e error) {
    txDB := getTxDB(repo) //获取事务db
    defer entity.RemoveAllPubEvent() //清理实体事件

    // 保存发布事件到事务表
    for _, domainEvent := range entity.GetPubEvents() {
        if !manager.eventRetry.pubExist(domainEvent.Topic()) {
            continue //未注册重试,无需存储
        }

        content, err := domainEvent.Marshal()
        if err != nil {
            return err
        }

        // 将事件保存到本地消息表
        model := pubEventObject{
            Identity: domainEvent.Identity(),
            Topic:    domainEvent.Topic(),
            Content:  string(content),
            Created:  time.Now(),
            Updated:  time.Now(),
        }
        if e = txDB.Create(&model).Error; e != nil {
            return
        }
    }
    
    // 将事件添加到Worker，等待事务提交后发送
    manager.addPubToWorker(repo.Worker(), entity.GetPubEvents())
    return
}
```

2. **消息发送阶段**：
```go
// push 事务提交后异步推送到Kafka
func (manager *EventManager) push(event freedom.DomainEvent) {
    freedom.Logger().Debugf("Publish Event Topic:%s, %+v", event.Topic(), event)
    identity := event.Identity()
    go func() {
        // 创建Kafka消息
        msg := manager.kafkaProducer.NewMsg(event.Topic(), event.Marshal())
        msg.SetMessageKey(fmt.Sprint(identity))

        // 发送消息到Kafka
        if err := msg.Publish(); err != nil {
            freedom.Logger().Error(err)
            return
        }

        // 发送成功后删除本地消息表中的记录
        if err := manager.db().Where("identity = ?", identity).Delete(&pubEventObject{}).Error; err != nil {
            freedom.Logger().Error(err)
        }
    }()
}
```

3. **消息重试机制**：
```go
// SetRetryPolicy 设置重试策略
func (manager *EventManager) SetRetryPolicy(delay time.Duration, retries int) {
    manager.eventRetry.setRetryPolicy(delay, retries)
}

// Retry 执行重试
func (manager *EventManager) Retry() {
    manager.eventRetry.retry()
}
```

#### 3. 一致性保证

EventManager 通过以下机制保证消息的最终一致性：

1. **本地事务保证**：
   - 业务数据变更和事件记录在同一个事务中
   - 事务失败时自动回滚，保证原子性
   - 事务成功后事件一定被记录在本地消息表

2. **消息可靠投递**：
   - 异步发送消息到Kafka
   - 发送成功后删除本地记录
   - 发送失败时保留记录，通过重试机制处理

3. **消息重试机制**：
   - 定时扫描未处理的消息
   - 根据配置的重试策略进行重试
   - 超过重试次数后进入告警流程

#### 4. 使用示例

```go
// 1. 注册事件重试
func init() {
    manager := domainevent.GetEventManager()
    // 注册发布事件重试
    manager.BindRetryPubEvent(&event.ShopGoods{})
    // 设置重试策略：5分钟间隔，最多重试3次
    manager.SetRetryPolicy(time.Minute*5, 3)
}

// 2. 在业务代码中使用
type OrderService struct {
    EventTransaction *domainevent.EventTransaction
    GoodsRepo        *repository.GoodsRepository
}

func (srv *OrderService) Shop(goodsID, num, userID int) error {
    // 创建购物事件
    shopEvent := &event.ShopGoods{
        UserID:    userID,
        GoodsID:   goodsID,
        GoodsNum:  num,
    }
    
    // 添加事件到实体
    goodsEntity.AddPubEvent(shopEvent)

    // 执行事务，事务成功后自动发送事件
    return srv.EventTransaction.Execute(func() error {
        return srv.GoodsRepo.Save(goodsEntity)
    })
}
```

#### 5. 最佳实践

1. **事件设计**：
   - 事件应该包含足够的上下文信息
   - 使用唯一的事件ID标识
   - 合理设计事件主题

2. **错误处理**：
   - 记录详细的错误日志
   - 设置合理的重试策略
   - 实现监控和告警机制

3. **性能优化**：
   - 使用批量处理提高性能
   - 合理设置Kafka分区
   - 定期清理已处理的消息

### 最佳实践

1. **事件设计原则**：
   - 使用过去时态命名事件（如：OrderCreated, PaymentCompleted）
   - 事件应该是不可变的
   - 事件应包含必要的上下文信息
   - 避免在事件中包含敏感信息

2. **事务一致性**：
   - 业务数据变更和事件发布在同一个事务中
   - 使用事务表作为消息队列的可靠发布保证
   - 通过定时任务重试失败的事件

3. **消费端处理**：
   - 实现幂等性处理
   - 合理设置消费者组
   - 正确处理消费失败的情况

4. **运维建议**：
   - 监控事件处理延迟
   - 设置合理的重试策略
   - 实现死信队列
   - 做好日志记录

### 调试与监控

```go
// 事件发布日志
freedom.Logger().Info("事件发布",
	"topic", event.Topic(),
	"eventId", event.Identity(),
	"content", string(event.Marshal()))

// 事件处理日志
freedom.Logger().Info("事件处理开始", 
	"topic", msg.Topic(),
	"eventId", eventID)
```





更多示例和详细文档，请参考Freedom框架文档。
