package controller

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/application"
	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/example/fshop/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/order", &Order{})
	})
}

type Order struct {
	Runtime     freedom.Runtime
	JSONRequest *infra.JSONRequest //基础设施 用于处理客户端请求io的json数据和验证
	OrderSrv    *application.Order //订单领域服务
}

// PutPay 支付订单, PUT: /order/pay route.
func (o *Order) PutPay() freedom.Result {
	var req dto.OrderPayReq
	e := o.JSONRequest.ReadBodyJSON(&req)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = o.OrderSrv.Pay(req.OrderNo, req.UserId)
	return &infra.JSONResponse{Error: e}
}

// GetItems 获取商品列表, GET: /order/items route.
func (o *Order) GetItems() freedom.Result {
	page := o.Runtime.Ctx().URLParamIntDefault("page", 1)
	pageSize := o.Runtime.Ctx().URLParamIntDefault("pageSize", 10)
	userId, err := o.Runtime.Ctx().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	dtoItems, totalPage, err := o.OrderSrv.Items(userId, page, pageSize)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	o.Runtime.Ctx().Header("X-Total-Page", strconv.Itoa(totalPage))

	return &infra.JSONResponse{Object: dtoItems}
}
