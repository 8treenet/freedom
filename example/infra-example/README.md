# freedom
### repository和事务组件
---

#### 本篇概要 
- 围绕购物项目
- CRUD代码生成 
- repository
- 事务组件
- 单元测试
- 如何自定义基础设施组件

#### CRUD代码生成
###### 生成Goods和Order的模型
```sh
#前置1. 导入./freedom.sql 到数据库
#前置2. 编辑./server/conf/db.toml
#创建表模型 freedom new-crud
$ freedom new-crud
```

#### repository
###### 数据仓库，用来读写Goods和Order的模型，限于篇幅 示例只使用到mysql数据。
```go
package repositorys

import (
	"github.com/8treenet/freedom"
	//objects 目录下的模型 和repositorys/generate.go 由crud生成
	"github.com/8treenet/freedom/example/infra-example/application/objects"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		//初始化绑定该仓库实例
		initiator.BindRepository(func() *GoodsRepository {
			return &GoodsRepository{}
		})
	})
}

// GoodsRepository .
type GoodsRepository struct {
	freedom.Repository	//仓库必须继承 freedom.Repository，否则无法使用数据源
}

func (repo *GoodsRepository) Get(id int) (result objects.Goods, e error) {
	/*
		通过主键id 查询该商品
		findGoodsByPrimary 由代码生成
	*/
	result, e = findGoodsByPrimary(repo, id)
	return
}

func (repo *GoodsRepository) GetAll() (result []objects.Goods, e error) {
	/*
		查看全部商品
		findGoodss由代码生成
	*/
	result, e = findGoodss(repo, objects.Goods{})
	return
}

func (repo *GoodsRepository) Save(goods *objects.Goods) (e error) {
	/*
		保存更新商品
		updateGoods 由代码生成
	*/
	_, e = updateGoods(repo, goods)
	return
}
```


#### 事务组件
###### 事务组件也是以依赖注入的方式使用，简单的来说定义成员变量后开箱即用。本示例在service定义使用，也可以在repository定义使用。
```go
import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/repositorys"
	"github.com/8treenet/freedom/example/infra-example/objects"
	"github.com/8treenet/freedom/infra/transaction"
)
func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
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
	Runtime   freedom.Runtime
	GoodsRepo repositorys.GoodsInterface	//repositorys包声明的商品仓库接口
	OrderRepo repositorys.OrderInterface	//repositorys包声明的订单仓库接口

	/*
		github.com/8treenet/freedom/infra/transaction
		type Transaction interface {
			Execute(fun func() error) (e error)
			ExecuteTx(fun func() error, ctx context.Context, opts *sql.TxOptions) (e error)
		}
	*/
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
// application/goods_test.go
func TestGoodsService_Get(t *testing.T) {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallGorm(func() (db *gorm.DB) {
		//这是连接数据库方式，mock方式参见TestGoodsService_MockGet
		conf := config.Get().DB
		var e error
		db, e = gorm.Open("mysql", conf.Addr)
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}

		db.DB().SetMaxIdleConns(conf.MaxIdleConns)
		db.DB().SetMaxOpenConns(conf.MaxOpenConns)
		db.DB().SetConnMaxLifetime(time.Duration(conf.ConnMaxLifeTime) * time.Second)
		return
	})

	//启动单元测试
	unitTest.Run()

	var srv *GoodsService
	//获取领域服务
	unitTest.GetService(&srv)
	//调用服务方法。 篇幅有限，没有具体逻辑 只是读取个数据
	rep, err := srv.Get(1)
	if err != nil {
		panic(rep)
	}
	//在这里做 rep的case，如果不满足条件 触发panic
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
	MadeEntity(entity interface{})
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
	freedom.Booting(func(initiator freedom.Initiator) {
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

/*
	不论单例组件还是多例组件 必须继承freedom.Infra, freedom.Infra提供了一些基础设施相关的功能。
*/
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
	if e := goods.JSONRequest.ReadBodyJSON(&request); e != nil {
		return &infra.JSONResponse{Err: e}
	}
}
```

```go
//组件的实现
func init() {
	validate = validator.New()
	freedom.Booting(func(initiator freedom.Initiator) {
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
	freedom.Infra	//继承freedom.Infra
}

// BeginRequest 每一个请求只会触发一次
func (req *JSONRequest) BeginRequest(rt freedom.Runtime) {
	// 调用基类初始化请求运行时
	req.Infra.BeginRequest(rt)
}

// ReadBodyJSON .
func (req *JSONRequest) ReadBodyJSON(obj interface{}) error {
	//从上下文读取io数据
	rawData, err := ioutil.ReadAll(req.Runtime.Ctx().Request().Body)
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