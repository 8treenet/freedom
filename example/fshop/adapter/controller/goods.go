package controller

import (
	application "github.com/8treenet/freedom/example/fshop/domain"
	"github.com/8treenet/freedom/example/fshop/domain/vo"
	"github.com/8treenet/freedom/example/fshop/infra"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/goods", &GoodsController{})
	})
}

// GoodsController .
type GoodsController struct {
	GoodsSev *application.GoodsService //商品领域服务
	Worker   freedom.Worker            //运行时，一个请求绑定一个运行时
	Request  *infra.Request            //基础设施 用于处理客户端请求io的json数据和验证
}

// GetItems 获取商品列表, GET: /goods/items route.
func (g *GoodsController) GetItems() freedom.Result {
	tag := g.Worker.IrisContext().URLParamDefault("tag", "")
	page := g.Worker.IrisContext().URLParamIntDefault("page", 1)
	pageSize := g.Worker.IrisContext().URLParamIntDefault("pageSize", 10)

	items, err := g.GoodsSev.Items(page, pageSize, tag)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: items}
}

// Post 添加商品, POST: /goods
func (g *GoodsController) Post() freedom.Result {
	var req vo.GoodsAddReq
	e := g.Request.ReadJSON(&req, true)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = g.GoodsSev.New(req.Name, req.Price)
	return &infra.JSONResponse{Error: e}
}

// PutStockByBy 增加库存, PUT: /goods/stock/:id/:num
func (g *GoodsController) PutStockByBy(id, num int) freedom.Result {
	e := g.GoodsSev.AddStock(id, num)
	return &infra.JSONResponse{Error: e}
}

// PutTag 为商品打tag, PUT: /goods/tag
func (g *GoodsController) PutTag() freedom.Result {
	var req vo.GoodsTagReq
	e := g.Request.ReadJSON(&req, true)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = g.GoodsSev.MarkedTag(req.ID, req.Tag)
	return &infra.JSONResponse{Error: e}
}

// PostShop 为商品打tag, POST: /goods/shop
func (g *GoodsController) PostShop() freedom.Result {
	var req vo.GoodsShopReq
	e := g.Request.ReadJSON(&req, true)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = g.GoodsSev.Shop(req.ID, req.Num, req.UserID)
	return &infra.JSONResponse{Error: e}
}
