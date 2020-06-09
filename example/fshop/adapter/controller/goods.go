package controller

import (
	"github.com/8treenet/freedom/example/fshop/adapter/dto"
	application "github.com/8treenet/freedom/example/fshop/domain"
	"github.com/8treenet/freedom/example/fshop/infra"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/goods", &Goods{})
	})
}

type Goods struct {
	GoodsSev *application.Goods //商品领域服务
	Worker   freedom.Worker     //运行时，一个请求绑定一个运行时
	Request  *infra.Request     //基础设施 用于处理客户端请求io的json数据和验证
}

// GetItems 获取商品列表, GET: /goods/items route.
func (g *Goods) GetItems() freedom.Result {
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
func (g *Goods) Post() freedom.Result {
	var req dto.GoodsAddReq
	e := g.Request.ReadJSON(&req)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = g.GoodsSev.New(req.Name, req.Price)
	return &infra.JSONResponse{Error: e}
}

// PutStock 增加库存, PUT: /goods/stock/:id/:num
func (g *Goods) PutStockByBy(id, num int) freedom.Result {
	e := g.GoodsSev.AddStock(id, num)
	return &infra.JSONResponse{Error: e}
}

// PutTag 为商品打tag, PUT: /goods/tag
func (g *Goods) PutTag() freedom.Result {
	var req dto.GoodsTagReq
	e := g.Request.ReadJSON(&req)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = g.GoodsSev.MarkedTag(req.Id, req.Tag)
	return &infra.JSONResponse{Error: e}
}

// PostShop 为商品打tag, POST: /goods/shop
func (g *Goods) PostShop() freedom.Result {
	var req dto.GoodsShopReq
	e := g.Request.ReadJSON(&req)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}

	e = g.GoodsSev.Shop(req.Id, req.Num, req.UserId)
	return &infra.JSONResponse{Error: e}
}
