package repository

import (
	"fmt"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/8treenet/freedom/infra/store"
	"github.com/jinzhu/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建资源库函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindRepository(func() *Order {
			return &Order{} //创建Order资源库
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.OrderRepo = (*Order)(nil)

// Order .
type Order struct {
	freedom.Repository
	Cache store.EntityCache //实体缓存组件
}

// BeginRequest .
func (repo *Order) BeginRequest(worker freedom.Worker) {
	repo.Repository.BeginRequest(worker)
	//设置缓存的持久化数据源
	repo.Cache.SetSource(func(result freedom.Entity) error {
		orderEntity := result.(*entity.Order)
		if e := findOrder(repo, orderEntity); e != nil {
			return e
		}
		return findOrderDetailList(repo, po.OrderDetail{OrderNo: orderEntity.OrderNo}, &orderEntity.Details)
	})
}

//New 新建订单实体 .
func (repo *Order) New() (orderEntity *entity.Order, e error) {
	orderNo := fmt.Sprint(time.Now().Unix())
	orderEntity = &entity.Order{Order: po.Order{OrderNo: orderNo, Status: "NON_PAYMENT", Created: time.Now(), Updated: time.Now()}}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(orderEntity)
	return
}

// Save 保存订单实体
func (repo *Order) Save(orderEntity *entity.Order) (e error) {
	if orderEntity.ID == 0 {
		//新建
		_, e = createOrder(repo, &orderEntity.Order)
		if e != nil {
			return
		}
		for i := 0; i < len(orderEntity.Details); i++ {
			_, e = createOrderDetail(repo, orderEntity.Details[i])
			if e != nil {
				return
			}
		}
		return
	}

	_, e = saveOrder(repo, &orderEntity.Order)
	if e != nil {
		return
	}

	for i := 0; i < len(orderEntity.Details); i++ {
		_, e = saveOrderDetail(repo, orderEntity.Details[i])
		if e != nil {
			return
		}
	}
	repo.Cache.Delete(orderEntity)
	return
}

// Find .
func (repo *Order) Find(orderNo string, userID int) (orderEntity *entity.Order, e error) {
	orderEntity = &entity.Order{Order: po.Order{OrderNo: orderNo, UserID: userID}}
	e = findOrder(repo, orderEntity)
	if e != nil {
		return
	}

	e = findOrderDetailList(repo, po.OrderDetail{OrderNo: orderEntity.OrderNo}, &orderEntity.Details)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(orderEntity)
	return
}

// Finds .
func (repo *Order) Finds(userID int, page, pageSize int) (entitys []*entity.Order, totalPage int, e error) {
	pager := NewDescPager("id").SetPage(page, pageSize)

	e = findOrderList(repo, po.Order{UserID: userID}, &entitys, pager)
	if e != nil {
		return
	}

	for i := 0; i < len(entitys); i++ {
		e = findOrderDetailList(repo, po.OrderDetail{OrderNo: entitys[i].OrderNo}, &entitys[i].Details)
		if e != nil {
			return
		}
	}

	totalPage = pager.TotalPage()
	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntitys(entitys)
	return
}

// Get .
func (repo *Order) Get(orderNo string) (orderEntity *entity.Order, e error) {
	orderEntity = &entity.Order{Order: po.Order{OrderNo: orderNo}}
	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(orderEntity)

	return orderEntity, repo.Cache.GetEntity(orderEntity)
}

func (repo *Order) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	db = db.New()
	db.SetLogger(repo.Worker.Logger())
	return db
}
