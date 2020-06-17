# freedom
### 事务组件和自定义组件
---

#### 本篇概要 
- CRUD代码生成 
- 事务组件
- 单元测试
- 如何自定义基础设施组件

#### CRUD代码生成
###### 生成Goods和Order的模型
```sh
#前置1. 导入./freedom.sql 到数据库
#前置2. 编辑./server/conf/db.toml
#创建表模型 freedom new-po
$ freedom new-po
```

#### 事务组件
###### 事务组件也是以依赖注入的方式使用，简单的来说定义成员变量后开箱即用。本示例在service定义使用，也可以在repository定义使用。
```go
import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/repository"
	"github.com/8treenet/freedom/example/infra-example/objects"
	"github.com/8treenet/freedom/infra/transaction"
)
func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//初始化绑定该服务实例
		initiator.BindService(func() *OrderService {
			return &OrderService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *OrderService) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// OrderService .
type OrderService struct {
	Worker   freedom.Worker
	GoodsRepo repository.GoodsInterface	//repositorys包声明的商品仓库接口
	OrderRepo repository.OrderInterface	//repositorys包声明的订单仓库接口
	Tx        transaction.Transaction 	    //transaction包下的 事务组件接口
}

func (srv *OrderService) Add(goodsID, num, userId int) (resp string, e error) {
	goodsObj, e := srv.GoodsRepo.Get(goodsID)
	if e != nil {
		return
	}
	if goodsObj.Stock < num {
		resp = "库存不足"
		return
	}
	goodsObj.AddStock(-num)

	/*
		事务执行
		Transaction.Execute() error
		如果错误为 nil, 触发 db.Commit
		如果错误不为nil, 触发 db.Rollback
	*/
	e = srv.Tx.Execute(func() error {
		if err := srv.GoodsRepo.Save(&goodsObj); err != nil {
			return err
		}

		return srv.OrderRepo.Create(goodsObj.ID, num, userId)
	})
	return
}
```

#### 单元测试
```go
func TestGoodsService_Get(t *testing.T) {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallGorm(func() (db *gorm.DB) {
		//这是连接数据库方式，mock方式参见goods_test.go
		conf := conf.Get().DB
		db, _ = gorm.Open("mysql", conf.Addr)
		return
	})

	//启动单元测试
	unitTest.Run()

	var srv *GoodsService
	//获取领域服务
	unitTest.GetService(&srv)
	//调用服务方法执行测试
	rep, err := srv.Get(1)
	if err != nil {
		panic(rep)
	}
}
```
###### 单测工具介绍
```go
type UnitTest interface {
	//获取领域服务
	GetService(service interface{})
	//获取资源库
	GetRepository(repository interface{})
	//安装gorm
	InstallGorm(f func() (db *gorm.DB))
	//安装redis
	InstallRedis(f func() (client redis.Cmdable))
	//启动
	Run()
	//设置请求
	SetRequest(request *http.Request)
	//创建实体
	InjectBaseEntity(entity interface{})
	//安装领域事件
	InstallDomainEventInfra(eventInfra DomainEventInfra)
	//创建一个领域事件的回调函数
	NewDomainEventInfra(call ...func(producer, topic string, data []byte, header map[string]string)) DomainEventInfra
}
```

#### 如何自定义基础设施组件
- 单例组件入口是 Booting, 生命周期为常驻
- 多例组件入口是 BeginRequest，生命周期为一个请求会话
- 多例组件一个请求会话中共享一个实例，简单的讲 controller、service、repository 如果引用同一多例组件，本质使用同一实例。
- 框架已提供的组件目录 github.com/8treenet/freedom/infra
- 用户自定义的组件目录 [project]/infra/[custom]
- 组件可以独立使用组件的配置文件, 配置文件放在 [project]/server/conf/infra/[custom.toml]


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
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

type Single struct {
	life int //生命
}

// Booting 单例组件入口, 启动时调用一次。
func (c *Single) Booting(boot freedom.SingleBoot) {
	freedom.Logger().Info("Single.Booting")
	c.life = rand.Intn(100)
}

func (mu *Single) GetLife() int {
	//所有请求的访问 都是一样的life
	return mu.life
}
```

##### 多例组件
###### 实现一个读取json数据的组件，并且做数据验证。
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
			//绑定该组件到基础设施池。
			return &JSONRequest{}
		})
		initiator.InjectController(func(ctx freedom.Context) (com *JSONRequest) {
			//从基础设施池里取出注入到控制器。
			initiator.GetInfra(ctx, &com)
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
	rawData, err := ioutil.ReadAll(req.Worker.Ctx().Request().Body)
	if err != nil {
		return err
	}

	/*
		使用第三方 extjson 反序列化
		使用第三方 validate 做数据验证
	*/
	err = extjson.Unmarshal(rawData, obj)
	if err != nil {
		return err
	}

	return validate.Struct(obj)
}
```