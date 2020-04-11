package controller

import (
	"github.com/8treenet/freedom/example/fshop/infra"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/application"
	"github.com/8treenet/freedom/example/fshop/application/dto"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/cart", &Cart{})
	})
}

type Cart struct {
	Runtime     freedom.Runtime
	CartSev     *application.Cart  //购物车领域服务
	JSONRequest *infra.JSONRequest //基础设施 用于处理客户端请求io的json数据和验证
}

// GetItems 获取购物车商品列表, GET: /cart/items route.
func (c *Cart) GetItems() freedom.Result {
	userId, err := c.Runtime.Ctx().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	dto, err := c.CartSev.Items(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: dto}
}

// Post 添加商品到购物车, POST: /cart route.
func (c *Cart) Post() freedom.Result {
	var req dto.CartAddReq
	e := c.JSONRequest.ReadBodyJSON(&req)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	c.CartSev.Add(req.UserId, req.GoodsId, req.GoodsNum)
	return &infra.JSONResponse{Error: e}
}

// DeleteAll 删除购物车全部商品, DELETE: /cart/all route.
func (c *Cart) DeleteAll() freedom.Result {
	userId, err := c.Runtime.Ctx().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	err = c.CartSev.DeleteAll(userId)
	return &infra.JSONResponse{Error: err}
}

// PostShop 购物购物车全部买商品, POST: /cart/shop route.
func (c *Cart) PostShop() freedom.Result {
	var req dto.CartShopReq
	e := c.JSONRequest.ReadBodyJSON(&req)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = c.CartSev.Shop(req.UserId)
	return &infra.JSONResponse{Error: e}
}
