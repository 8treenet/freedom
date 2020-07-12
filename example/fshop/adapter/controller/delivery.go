package controller

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain"
	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/example/fshop/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		controller := &Delivery{}
		initiator.BindController("/delivery", controller)
		//监听订单支付事件
		initiator.ListenEvent("order-pay", "Delivery.PostOrderPayBy")
	})
}

type Delivery struct {
	Worker   freedom.Worker
	OrderSrv *domain.Order  //订单领域服务
	Request  *infra.Request //基础设施 用于处理客户端请求io的json数据和验证
}

// PostOrderPayBy 返货提醒, POST: /delivery/order/pay route.
func (d *Delivery) PostOrderPayBy(eventID string) error {
	var msg dto.OrderPayMsg
	d.Request.ReadJSON(&msg)
	d.Worker.Logger().Info("发货提醒 eventId:", eventID, "msg:", msg)
	return nil
}

// Post 发货, POST: /delivery route.
func (d *Delivery) Post() freedom.Result {
	var req dto.DeliveryReq
	if e := d.Request.ReadJSON(&req); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e := d.OrderSrv.Delivery(req)
	return &infra.JSONResponse{Error: e}
}
