# freedom

### 事务组件和自定义组件

---

#### 本篇概要

- 事务组件
- 基于 KafkaMQ 的领域事件
- 如何自定义 Infra 组件

#### 事务组件

###### 事务组件也是以依赖注入的方式使用，简单的来说定义成员变量后开箱即用。

```go
import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/adapter/repository"
	"github.com/8treenet/freedom/example/infra-example/domain/event"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//初始化绑定该服务实例
		initiator.BindService(func() *OrderService {
			return &OrderService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *OrderService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// OrderService .
type OrderService struct {
	Worker           freedom.Worker	 	           //请求运行时对象
	GoodsRepo        *repository.GoodsRepository   //商品资源库
	OrderRepo        *repository.OrderRepository   //订单资源库
	EventTransaction *domainevent.EventTransaction //事务组件
}

// Shop 这不是一个正确的示例，只是为展示领域事件和Kafka的结合, 请参考fshop的聚合根.
func (srv *OrderService) Shop(goodsID, num, userID int) (e error) {
	//通过商品资源库获取商品实体
	goodsEntity, e := srv.GoodsRepo.Get(goodsID)
	if e != nil {
		return
	}
	if goodsEntity.Stock < num {
		e = errors.New("库存不足")
		return
	}
	goodsEntity.AddStock(-num) //商品实体扣库存

	//给商品实体增加购买Pub事件
	goodsEntity.AddPubEvent(&event.ShopGoods{
		UserID:    userID,
		GoodsID:   goodsID,
		GoodsNum:  num,
		GoodsName: goodsEntity.Name,
	})

	//使用事务组件保证一致性 1.修改商品库存, 2.创建订单, 3.事件表增加记录
	//Execute 如果返回错误 会触发回滚。成功会调用infra/domainevent/EventManager.push(event)
	e = srv.EventTransaction.Execute(func() error {
		if err := srv.GoodsRepo.Save(goodsEntity); err != nil {
			return err
		}

		return srv.OrderRepo.Create(goodsEntity.ID, num, userID)
	})
	return
}
```

#### 基于 KafkaMQ 的领域事件

###### 领域事件的组件的 持久化和 Pub/Sub 的处理

```go
package repository
import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/entity"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
)

// GoodsRepository .
type GoodsRepository struct {
	freedom.Repository
	EventManager *domainevent.EventManager //领域事件组件
}

// Save .
func (repo *GoodsRepository) Save(goods *entity.Goods) (e error) {
	//这个方法内是由事务组件调用的，所以它们会一起Commit/Rollback
	_, e = saveGoods(repo, &goods.Goods)//持久化化Goods实体到Mysql.
	if e != nil {
		return
	}
 	//持久化实体身上的Pub/Sub事件
	return repo.EventManager.Save(&repo.Repository, goods)
}
```

```go
package domainevent
// EventManager .
type EventManager struct {
	freedom.Infra
	kafkaProducer *kafka.ProducerImpl //Kafka Producer组件
}

//只介绍Pub的处理, Sub事件参考代码.
func (manager *EventManager) Save(repo *freedom.Repository, entity freedom.Entity) (e error) {
	txDB := getTxDB(repo) //获取事务db
	defer entity.RemoveAllPubEvent() //删除实体里的全部事件

	//获取实体的发布事件，插入到发布事件表
	for _, domainEvent := range entity.GetPubEvent() {
		model := domainEventPublish{
			Topic:   domainEvent.Topic(),
			Content: string(domainEvent.Marshal()),
			Created: time.Now(),
			Updated: time.Now(),
		}
		e = txDB.Create(&model).Error //插入发布事件表。
		if e != nil {
			return
		}
		domainEvent.SetIdentity(model.ID)
	}

	//添加到到Worker，等待事务成功的动作
	manager.addPubToWorker(repo.GetWorker(), entity.GetPubEvent())
	return
}

//EventTransaction在事务成功后调用 .
func (manager *EventManager) push(event freedom.DomainEvent) {
	freedom.Logger().Infof("领域事件发布 Topic:%s, %+v", event.Topic(), event)
	eventID := event.Identity().(int)
	go func() {
		//Kafka组件创建消息
		msg := manager.kafkaProducer.NewMsg(event.Topic(), event.Marshal()).SetHeader(event.GetPrototypes())
		msg.SetMessageKey(fmt.Sprint(eventID)) //设置kafka消息key

		//Kafka消息发送
		if err := msg.Publish(); err != nil {
			//如果失败，定时器会扫表继续发布
			freedom.Logger().Error(err)
			return
		}

		publish := &domainEventPublish{ID: eventID}
		// Push成功后删除发布事件表
		if err := manager.db().Delete(&publish).Error; err != nil {
			freedom.Logger().Error(err)
		}
	}()
}
```

```go
package domainevent
//事务组件 .
type EventTransaction struct {
	transaction.GormImpl
}

// 事务处理 .
func (et *EventTransaction) Execute(fun func() error) (e error) {
	if e = et.GormImpl.Execute(fun); e != nil {
		//如果事务失败
		return
	}
	//如果成功触发消息Push动作
	et.pushEvent()
	return
}

func (et *EventTransaction) pushEvent() {
	pubs := et.Worker().Store().Get(workerStorePubEventKey)
	if pubs == nil {
		return
	}
	pubEvents, ok := pubs.([]freedom.DomainEvent)
	if !ok {
		return
	}

	for _, pubEvent := range pubEvents {
		eventManager.push(pubEvent) //使用manager 推送消息
	}
	return
}
```

#### 如何自定义基础设施组件

- 单例组件入口是 Booting, 生命周期为常驻
- 多例组件入口是 BeginRequest，生命周期为一个请求会话
- 框架已提供的组件目录 github.com/8treenet/freedom/infra
- 用户自定义的组件目录 [project]/infra/[custom]

##### 单例的组件

```go
func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		/*
			绑定组件
			single 是否单例
			com :如果是单例com是组件指针， 如果是多例 com是创建组件的函数
			BindInfra(single bool, com interface{})
		*/
		initiator.BindInfra(true, &Single{})

		/*
			该组件注入到控制器, 默认仅注入到service和repository
			如果不调用 initiator.InjectController, 控制器无法使用。
		*/
		initiator.InjectController(func(ctx freedom.Context) (com *Single) {
			initiator.FetchInfra(ctx, &com)
			return
		})
	})
}

type Single struct {
	life int
}

// Booting 单例组件入口, 启动时调用一次。
func (c *Single) Booting(boot freedom.BootManager) {
	freedom.Logger().Info("Single.Booting")
	c.life = rand.Intn(100)
}

func (mu *Single) GetLife() int {
	//所有请求的访问 都是一样的life
	return mu.life
}
```

##### 多例组件

###### 实现一个读取 json 数据的组件，并且做数据验证。

```go
//使用展示
type GoodsController struct {
	JSONRequest *infra.JSONRequest
}
// GetBy handles the PUT: /goods/stock route 增加商品库存.
func (goods *GoodsController) PutStock() freedom.Result {
	var request struct {
		GoodsID int `json:"goodsId" validate:"required"` //商品id
		Num     int `validate:"min=1,max=15"`//只能增加的范围1-15，其他报错
	}

	//使用自定义的json组件读取请求数据, 并且处理数据验证。
	if e := goods.JSONRequest.ReadJSON(&request); e != nil {
		return &infra.JSONResponse{Err: e}
	}
}
```

```go
//组件的实现
func init() {
	validate = validator.New()
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *JSONRequest {
			//绑定1个New多例组件的回调函数，多例目的是为每个请求独立服务。
			return &JSONRequest{}
		})
		initiator.InjectController(func(ctx freedom.Context) (com *JSONRequest) {
			//从Infra池里取出注入到控制器。
			initiator.FetchInfra(ctx, &com)
			return
		})
	})
}

// Transaction .
type JSONRequest struct {
	freedom.Infra	//多例需继承freedom.Infra
}

// BeginRequest 每一个请求只会触发一次
func (req *JSONRequest) BeginRequest(worker freedom.Worker) {
	// 调用基类初始化请求运行时
	req.Infra.BeginRequest(worker)
}

// ReadJSON .
func (req *JSONRequest) ReadJSON(obj interface{}) error {
	//从上下文读取io数据
	rawData, err := ioutil.ReadAll(req.Worker().Ctx().Request().Body)
	if err != nil {
		return err
	}

	/*
		使用第三方 validate 做数据验证
	*/
	err = json.Unmarshal(rawData, obj)
	if err != nil {
		return err
	}

	return validate.Struct(obj)
}
```
