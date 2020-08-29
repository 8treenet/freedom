package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/order", &OrderController{})
	})
}

// OrderController .
type OrderController struct {
	Worker   freedom.Worker
	OrderSev *domain.OrderService
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
