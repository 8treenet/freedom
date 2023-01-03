package controller

import (
	domain "github.com/8treenet/freedom/example/fshop/domain"
	"github.com/8treenet/freedom/example/fshop/infra"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/vo"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/cart", &CartController{})
	})
}

// CartController .
type CartController struct {
	Worker  freedom.Worker
	CartSev *domain.CartService //购物车领域服务
	Request *infra.Request      //基础设施 用于处理客户端请求io的json数据和验证
}

// GetItems 获取购物车商品列表, GET: /cart/items route.
func (c *CartController) GetItems() freedom.Result {
	userID, err := c.Worker.IrisContext().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	vo, err := c.CartSev.Items(userID)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: vo}
}

// Post 添加商品到购物车, POST: /cart route.
func (c *CartController) Post() freedom.Result {
	var req vo.CartAddReq
	e := c.Request.ReadJSON(&req, true)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	c.CartSev.Add(req.UserID, req.GoodsID, req.GoodsNum)
	return &infra.JSONResponse{Error: e}
}

// DeleteAll 删除购物车全部商品, DELETE: /cart/all route.
func (c *CartController) DeleteAll() freedom.Result {
	userID, err := c.Worker.IrisContext().URLParamInt("userId")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	err = c.CartSev.DeleteAll(userID)
	return &infra.JSONResponse{Error: err}
}

// PostShop 购物购物车全部买商品, POST: /cart/shop route.
func (c *CartController) PostShop() freedom.Result {
	var req vo.CartShopReq
	e := c.Request.ReadJSON(&req, true)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = c.CartSev.Shop(req.UserID)
	return &infra.JSONResponse{Error: e}
}
