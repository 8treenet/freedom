package controller

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain"
	"github.com/8treenet/freedom/example/infra-example/domain/event"
	"github.com/8treenet/freedom/example/infra-example/domain/vo"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	//普通消息示例，领域事件请参考GoodsController
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/order", &OrderController{})
		initiator.ListenEvent((&event.OrderPay{}).Topic(), "OrderController.OrderPayEvent", false)
	})
}

// OrderController .
type OrderController struct {
	Worker   freedom.Worker
	OrderSev *domain.OrderService
	Request  *infra.Request
}

// BeforeActivation .
func (controller *OrderController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("PUT", "/orderPay", "OrderPay")
	b.Handle("POST", "/orderPayEvent", "OrderPayEvent")
}

// GetBy handles the GET: /order/:id route 通过id获取订单.
func (c *OrderController) GetBy(goodsID int) freedom.Result {
	userID, err := c.Worker.IrisContext().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	obj, err := c.OrderSev.Get(goodsID, userID)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: obj}
}

// Get handles the GET: /order route 获取全部订单.
func (c *OrderController) Get() freedom.Result {
	userID, err := c.Worker.IrisContext().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	objs, err := c.OrderSev.GetAll(userID)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: objs}
}

// PutPay handles the PUT: /order/orderPay
func (c *OrderController) OrderPay() freedom.Result {
	var req vo.OrderPayReq
	if e := c.Request.ReadJSON(&req); e != nil {
		return &infra.JSONResponse{Error: e}
	}
	err := c.OrderSev.Pay(req.OrderID, req.UserID)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

// PostConsumeOrderPay handles POST: /order/orderPayEvent
func (c *OrderController) OrderPayEvent() freedom.Result {
	var payEvent event.OrderPay
	if e := c.Request.ReadJSON(&payEvent); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	// 处理支付消息，这里可以添加其他业务逻辑
	c.Worker.Logger().Infof("收到订单支付消息: OrderID=%d, UserID=%d", payEvent.OrderID, payEvent.UserID)
	return &infra.JSONResponse{}
}
