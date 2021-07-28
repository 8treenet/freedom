package controller

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain"
	"github.com/8treenet/freedom/example/fshop/domain/vo"
	"github.com/8treenet/freedom/example/fshop/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/order", &Order{})
	})
}

// Order .
type Order struct {
	Worker   freedom.Worker
	Request  *infra.Request //基础设施 用于处理客户端请求io的json数据和验证
	OrderSrv *domain.Order  //订单领域服务
}

// PutPay 支付订单, PUT: /order/pay route.
func (o *Order) PutPay() freedom.Result {
	var req vo.OrderPayReq
	e := o.Request.ReadJSON(&req, true)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = o.OrderSrv.Pay(req.OrderNo, req.UserID)
	return &infra.JSONResponse{Error: e}
}

// GetItems 获取商品列表, GET: /order/items route.
func (o *Order) GetItems() freedom.Result {
	page := o.Worker.IrisContext().URLParamIntDefault("page", 1)
	pageSize := o.Worker.IrisContext().URLParamIntDefault("pageSize", 10)
	userID, err := o.Worker.IrisContext().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	voItems, totalPage, err := o.OrderSrv.Items(userID, page, pageSize)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	o.Worker.IrisContext().Header("X-Total-Page", strconv.Itoa(totalPage))

	return &infra.JSONResponse{Object: voItems}
}
