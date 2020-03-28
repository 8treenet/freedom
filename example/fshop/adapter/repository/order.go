package repository

import (
	"fmt"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/application/entity"
	"github.com/8treenet/freedom/example/fshop/application/object"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *Order {
			return &Order{}
		})
	})
}

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
