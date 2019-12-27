# freedom
### repository和事务组件
---

#### 本篇概要 
- 围绕购物项目
- 事务组件
- repository
- CRUD代码生成 
- 如何自定义基础设施组件

#### CRUD代码生成
###### 生成Goods和Order的模型
```sh
#前置1. 导入./freedom.sql 到数据库
#前置2. 编辑./cmd/conf/db.toml
$ freedom new-project [project-name]
```

#### repository
###### 数据仓库，用来读写Goods和Order的模型，限于篇幅 示例只使用到mysql数据。
```go
package repositorys

import (
	"github.com/8treenet/freedom"
	//models 目录下的模型 已通过crud生成
	"github.com/8treenet/freedom/example/infra-example/models"
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

func (repo *GoodsRepository) Get(id int) (result models.Goods, e error) {
	//通过主键id 查询该商品
	result, e = models.FindGoodsByPrimary(repo, id)
	return
}

func (repo *GoodsRepository) GetAll() (result []models.Goods, e error) {
	//查看全部商品
	result, e = models.FindGoodss(repo, models.Goods{})
	return
}

func (repo *GoodsRepository) ChangeStock(goods *models.Goods, num int) (e error) {
	//修改该商品库存
	models.UpdateGoods(repo, goods, models.Goods{
		Stock: goods.Stock + num,
	})
	return
}
```


#### 事务组件
###### 事务组件也是以依赖注入的方式使用，简单的来说定义成员变量后开箱即用。本示例在service定义使用，也可以在repository定义使用。
```go
import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/business/repositorys"
	"github.com/8treenet/freedom/example/infra-example/models"
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
	Tx        *transaction.Transaction 	    //transaction包下的 事务组件
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

	/*
		事务执行
		Transaction.Execute() error
		如果错误为 nil, 触发 db.Commit
		如果错误不为nil, 触发 db.Rollback
	*/
	e = srv.Tx.Execute(func() error {
		if err := srv.GoodsRepo.ChangeStock(&goodsObj, -num); err != nil {
			return err
		}

		return srv.OrderRepo.Create(goodsObj.ID, num, userId)
	})
	return
}
```


#### 如何自定义基础设施组件
- 单例组件入口是 Booting, 生命周期为常驻
- 多利组件入口是 BeginRequest，生命周期为一个请求会话
- 多利组件一个请求会话中共享一个实例，简单的讲 controller、service、repository 如果引用同一多利组件，本质使用同一实例。
- 框架已提供的组件目录 "github.com/8treenet/freedom/infra"
- 用户自定义的组件目录 "[project]/infra/[custom]"
- 组件可以独立使用组件的配置文件, 配置文件放在"[project]/cmd/conf/infra/[custom.toml]"


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
	freedom.Infra
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

##### 多例的组件
```go
func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *Multiton {
			return &Multiton{}
		})
		initiator.InjectController(func(ctx freedom.Context) (com *Multiton) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

// Multiton .
type Multiton struct {
	freedom.Infra
	life int //生命
}

// BeginRequest 多例组件入口，每个请求只调用一次。
func (mu *Multiton) BeginRequest(rt freedom.Runtime) {
	mu.Infra.BeginRequest(rt)
	rt.Logger().Info("Multiton 初始化拉")
	mu.life = rand.Intn(100)
}

func (mu *Multiton) GetLife() int {
	//每个请求不一样的life
	return mu.life
}
```