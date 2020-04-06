package entity

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/application/object"
)

const (
	OrderStatusPAID       = "PAID"
	OrderStatusNonPayment = "NON_PAYMENT"
)

// 订单实体
type Order struct {
	freedom.Entity
	object.Order
	Details []*object.OrderDetail
}

// Identity 唯一
func (o *Order) Identity() string {
	return o.OrderNo
}

// AddOrderDetal 增加订单详情
func (o *Order) AddOrderDetal(detal *object.OrderDetail) {
	o.Details = append(o.Details, detal)
}

// Pay 付款
func (o *Order) Pay() {
	o.SetStatus(OrderStatusPAID)
}
