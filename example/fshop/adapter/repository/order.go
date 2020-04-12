package repository

import (
	"fmt"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/application/entity"
	"github.com/8treenet/freedom/example/fshop/application/object"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *Order {
			return &Order{}
		})
	})
}

var _ OrderRepo = new(Order)

// Order .
type Order struct {
	freedom.Repository
}

// 新建订单实体
func (repo *Order) New() (orderEntity *entity.Order, e error) {
	orderNo := fmt.Sprint(time.Now().Unix())
	orderEntity = &entity.Order{Order: object.Order{OrderNo: orderNo, Status: "NON_PAYMENT", Created: time.Now(), Updated: time.Now()}}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(orderEntity)
	return
}

// Save 保存订单实体
func (repo *Order) Save(orderEntity *entity.Order) (e error) {
	if orderEntity.Id == 0 {
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
	return
}

// Find .
func (repo *Order) Find(orderNo string, userId int) (orderEntity *entity.Order, e error) {
	orderEntity = &entity.Order{}
	e = findOrder(repo, object.Order{OrderNo: orderNo, UserId: userId}, orderEntity)
	if e != nil {
		return
	}

	e = findOrderDetails(repo, object.OrderDetail{OrderNo: orderEntity.OrderNo}, &orderEntity.Details)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(orderEntity)
	return
}

// Finds .
func (repo *Order) Finds(userId int, page, pageSize int) (entitys []*entity.Order, totalPage int, e error) {
	pager := repo.NewORMDescBuilder("id").NewPageBuilder(page, pageSize)

	e = findOrders(repo, object.Order{UserId: userId}, &entitys, pager)
	if e != nil {
		return
	}

	for i := 0; i < len(entitys); i++ {
		e = findOrderDetails(repo, object.OrderDetail{OrderNo: entitys[i].OrderNo}, &entitys[i].Details)
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
	orderEntity = &entity.Order{}
	e = findOrder(repo, object.Order{OrderNo: orderNo}, orderEntity)
	if e != nil {
		return
	}

	e = findOrderDetails(repo, object.OrderDetail{OrderNo: orderEntity.OrderNo}, &orderEntity.Details)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(orderEntity)
	return
}
