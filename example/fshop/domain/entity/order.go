package entity

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/adapter/po"
)

const (
	OrderStatusPAID       = "PAID"
	OrderStatusNonPayment = "NON_PAYMENT"
	OrderStatusShipment   = "SHIPMENT"
)

// 订单实体
type Order struct {
	freedom.Entity
	po.Order
	Details []*po.OrderDetail
}

// Identity 唯一
func (o *Order) Identity() string {
	return o.OrderNo
}

// AddOrderDetal 增加订单详情
func (o *Order) AddOrderDetal(detal *po.OrderDetail) {
	o.Details = append(o.Details, detal)
}

// Pay 实体触发付款
func (o *Order) Pay() {
	o.SetStatus(OrderStatusPAID)
}

// Shipment 实体触发发货
func (o *Order) Shipment() {
	o.SetStatus(OrderStatusShipment)
}

// IsPay 是否支付
func (o *Order) IsPay() bool {
	if o.Status != OrderStatusPAID {
		return false
	}
	return true
}
