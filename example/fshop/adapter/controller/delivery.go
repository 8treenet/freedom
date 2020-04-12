package controller

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/application/dto"
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
	Runtime     freedom.Runtime
	JSONRequest *infra.JSONRequest //基础设施 用于处理客户端请求io的json数据和验证
}

// PostOrderPayBy 返货提醒, POST: /delivery/order/pay route.
func (d *Delivery) PostOrderPayBy(eventID string) error {
	var msg dto.OrderPayMsg
	d.JSONRequest.ReadBodyJSON(&msg)
	d.Runtime.Logger().Info("发货提醒 eventId:", eventID, "msg:", msg)
	return nil
}
