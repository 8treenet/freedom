package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/application"
	"github.com/8treenet/freedom/example/infra-example/infra"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindController("/order", &OrderController{})
	})
}

type OrderController struct {
	Runtime  freedom.Runtime
	OrderSev *application.OrderService
}

// GetBy handles the GET: /order/:id route 通过id获取订单.
func (c *OrderController) GetBy(goodsID int) freedom.Result {
	userId, err := c.Runtime.Ctx().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Err: err}
	}
	obj, err := c.OrderSev.Get(goodsID, userId)
	if err != nil {
		return &infra.JSONResponse{Err: err}
	}

	return &infra.JSONResponse{Object: obj}
}

// GetBy handles the GET: /order route 获取全部订单.
func (c *OrderController) Get() freedom.Result {
	userId, err := c.Runtime.Ctx().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Err: err}
	}
	objs, err := c.OrderSev.GetAll(userId)
	if err != nil {
		return &infra.JSONResponse{Err: err}
	}
	return &infra.JSONResponse{Object: objs}
}
