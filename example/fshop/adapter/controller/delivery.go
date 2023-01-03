package controller

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain"
	"github.com/8treenet/freedom/example/fshop/domain/vo"
	"github.com/8treenet/freedom/example/fshop/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		controller := &DeliveryController{}
		initiator.BindController("/delivery", controller)
		//监听订单支付事件
		initiator.ListenEvent("order-pay", "Delivery.PostOrderPayBy")
	})
}

// DeliveryController .
type DeliveryController struct {
	Worker   freedom.Worker
	OrderSrv *domain.OrderService //订单领域服务
	Request  *infra.Request       //基础设施 用于处理客户端请求io的json数据和验证
}

// PostOrderPayBy 返货提醒, POST: /delivery/order/pay route.
func (d *DeliveryController) PostOrderPayBy(eventID string) error {
	var msg vo.OrderPayMsg
	d.Request.ReadJSON(&msg)
	d.Worker.Logger().Info("发货提醒:", freedom.LogFields{"eventId": eventID, "msgBody": msg})
	return nil
}

// Post 发货, POST: /delivery route.
func (d *DeliveryController) Post() freedom.Result {
	var req vo.DeliveryReq
	if e := d.Request.ReadJSON(&req); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e := d.OrderSrv.Delivery(req)
	return &infra.JSONResponse{Error: e}
}
