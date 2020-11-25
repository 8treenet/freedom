package entity

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/po"
)

const (
	//OrderStatusPAID 已支付
	OrderStatusPAID = "PAID"
	//OrderStatusNonPayment 未支付
	OrderStatusNonPayment = "NON_PAYMENT"
	//OrderStatusShipment 已发货
	OrderStatusShipment = "SHIPMENT"
)

//Order 订单实体
type Order struct {
	freedom.Entity
	po.Order
	Details []*po.OrderDetail
}

// Identity 唯一
func (o *Order) Identity() string {
	return o.OrderNo
}

// AddOrderDetail 增加订单详情
func (o *Order) AddOrderDetail(detail *po.OrderDetail) {
	o.Details = append(o.Details, detail)
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
