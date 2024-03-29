package repository

import (
	"fmt"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/8treenet/freedom/example/fshop/infra/domainevent"
	"github.com/8treenet/freedom/infra/store"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建资源库函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindRepository(func() *OrderRepository {
			return &OrderRepository{} //创建Order资源库
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.OrderRepo = (*OrderRepository)(nil)

// OrderRepository .
type OrderRepository struct {
	freedom.Repository
	Cache        store.EntityCache         //实体缓存组件
	EventManager *domainevent.EventManager //领域事件组件
}

// BeginRequest .
func (repo *OrderRepository) BeginRequest(worker freedom.Worker) {
	repo.Repository.BeginRequest(worker)
	//设置缓存的持久化数据源
	repo.Cache.SetSource(func(result freedom.Entity) error {
		orderEntity := result.(*entity.Order)
		if e := findOrder(repo, &orderEntity.Order); e != nil {
			return e
		}
		details, e := findOrderDetailList(repo, po.OrderDetail{OrderNo: orderEntity.OrderNo})
		if e != nil {
			return e
		}
		for _, v := range details {
			orderEntity.Details = append(orderEntity.Details, &v)
		}
		return nil
	})
}

//New 新建订单实体 .
func (repo *OrderRepository) New() (orderEntity *entity.Order, e error) {
	orderNo := fmt.Sprint(time.Now().Unix())
	orderEntity = &entity.Order{Order: po.Order{OrderNo: orderNo, Status: "NON_PAYMENT", Created: time.Now(), Updated: time.Now()}}

	//注入基础Entity
	repo.InjectBaseEntity(orderEntity)
	return
}

// Save 保存订单实体
func (repo *OrderRepository) Save(orderEntity *entity.Order) (e error) {
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
		return repo.EventManager.Save(&repo.Repository, orderEntity) //持久化实体的事件
	}

	_, e = saveOrder(repo, orderEntity)
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
func (repo *OrderRepository) Find(orderNo string, userID int) (orderEntity *entity.Order, e error) {
	orderEntity = &entity.Order{Order: po.Order{OrderNo: orderNo, UserID: userID}}
	e = findOrder(repo, &orderEntity.Order)
	if e != nil {
		return
	}

	list, e := findOrderDetailList(repo, po.OrderDetail{OrderNo: orderEntity.OrderNo})
	if e != nil {
		return
	}
	for _, v := range list {
		orderEntity.Details = append(orderEntity.Details, &v)
	}

	//注入基础Entity
	repo.InjectBaseEntity(orderEntity)
	return
}

// Finds .
func (repo *OrderRepository) Finds(userID int, page, pageSize int) (entitys []*entity.Order, totalPage int, e error) {
	pager := NewDescPager("id").SetPage(page, pageSize)

	list, e := findOrderList(repo, po.Order{UserID: userID}, pager)
	if e != nil {
		return
	}
	for _, v := range list {
		entitys = append(entitys, &entity.Order{Order: v})
	}

	for i := 0; i < len(entitys); i++ {
		details, err := findOrderDetailList(repo, po.OrderDetail{OrderNo: entitys[i].OrderNo})
		if err != nil {
			e = err
			return
		}
		for _, detail := range details {
			entitys[i].Details = append(entitys[i].Details, &detail)
		}
	}

	totalPage = pager.TotalPage()
	//注入基础Entity
	repo.InjectBaseEntitys(entitys)
	return
}

// Get .
func (repo *OrderRepository) Get(orderNo string) (orderEntity *entity.Order, e error) {
	orderEntity = &entity.Order{Order: po.Order{OrderNo: orderNo}}
	//注入基础Entity
	repo.InjectBaseEntity(orderEntity)

	return orderEntity, repo.Cache.GetEntity(orderEntity)
}

func (repo *OrderRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
